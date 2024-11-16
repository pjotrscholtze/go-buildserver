package repo

import (
	"sort"
	"sync"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/websocketmanager"
	"github.com/pjotrscholtze/go-buildserver/models"
	"github.com/robfig/cron/v3"
)

type jobQueue struct {
	items        []models.Job
	finishedJobs []*models.Job
	lock         sync.Locker
	buildRepo    PipelineRepo
	wm           *websocketmanager.WebsocketManager
	nextID       int64
}

type JobQueue interface {
	Run()
	GetJobById(buildId int64) *models.Job
	AddQueueItem(repoName, buildReason, origin string)
	List() []*models.Job
	ListAllJobsOfPipeline(pipelineName string) []*models.Job
}

func (bq *jobQueue) GetJobById(buildId int64) *models.Job {
	for _, item := range bq.items {
		if item.ID == buildId {
			return &item
		}
	}
	for _, item := range bq.finishedJobs {
		if item.ID == buildId {
			return item
		}
	}
	return nil
}

func (bq *jobQueue) ListAllJobsOfPipeline(pipelineName string) []*models.Job {
	bq.lock.Lock()
	defer bq.lock.Unlock()
	return bq.listAllJobsOfPipelineUnsafe(pipelineName)
}
func (bq *jobQueue) listAllJobsOfPipelineUnsafe(pipelineName string) []*models.Job {
	out := []*models.Job{}
	for i := range bq.items {
		if bq.items[i].RepoName == pipelineName {
			out = append(out, &bq.items[i])
		}
	}
	for i := range bq.finishedJobs {
		if bq.finishedJobs[i].RepoName == pipelineName {
			out = append(out, bq.finishedJobs[i])
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})

	return out
}
func (bq *jobQueue) List() []*models.Job {
	out := []*models.Job{}
	bq.lock.Lock()
	defer bq.lock.Unlock()
	for i := range bq.items {
		out = append(out, &bq.items[i])
	}

	return out
}

func (bq *jobQueue) tick() *models.Job {
	bq.lock.Lock()
	defer bq.lock.Unlock()
	if len(bq.items) == 0 {
		return nil
	}
	bq.items[0].Status = "running"
	item := &bq.items[0]
	bq.items = bq.items[1:]
	bq.finishedJobs = append(bq.finishedJobs, item)
	return item
}
func (bq *jobQueue) Run() {
	for {
		if item := bq.tick(); item != nil {
			bq.wm.BroadcastOnEndpoint("jobs", "", bq.items)
			bq.wm.BroadcastOnEndpoint("repo-builds", item.RepoName, struct {
				Jobs []*models.Job
			}{
				Jobs: bq.ListAllJobsOfPipeline(item.RepoName),
			})

			bq.buildRepo.GetRepoByName(item.RepoName).Build(item)
			item.Status = "finished"
			bq.wm.BroadcastOnEndpoint("repo-builds", item.RepoName, struct {
				Jobs []*models.Job
			}{
				Jobs: bq.ListAllJobsOfPipeline(item.RepoName),
			})
		}
		time.Sleep(50 * time.Millisecond)
	}
}
func (bq *jobQueue) AddQueueItem(repoName, buildReason, origin string) {
	bq.lock.Lock()
	defer bq.lock.Unlock()
	bq.items = append(bq.items, models.Job{
		RepoName:    repoName,
		BuildReason: buildReason,
		Origin:      origin,
		QueueTime:   strfmt.DateTime(time.Now()),
		ID:          bq.nextID,
		Status:      "pending",
	})
	bq.nextID++
	bq.wm.BroadcastOnEndpoint("jobs", "", bq.items)
	bq.wm.BroadcastOnEndpoint("repo-builds", repoName, struct {
		Jobs []*models.Job
	}{
		Jobs: bq.listAllJobsOfPipelineUnsafe(repoName),
	})
}

func NewJobQueue(buildRepo PipelineRepo, cr *cron.Cron, wm *websocketmanager.WebsocketManager) JobQueue {
	bq := &jobQueue{
		items:     []models.Job{},
		lock:      &sync.Mutex{},
		buildRepo: buildRepo,
		wm:        wm,
		nextID:    1,
	}
	for _, repo := range buildRepo.List() {
		for _, trigger := range repo.GetTriggersOfKind("Cron") {
			cr.AddFunc(trigger.Schedule, func() {
				bq.AddQueueItem(repo.GetName(), trigger.Schedule, "cron")
			})
		}
	}

	return bq
}
