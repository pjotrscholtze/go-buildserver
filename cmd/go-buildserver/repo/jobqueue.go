package repo

import (
	"sync"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/websocketmanager"
	"github.com/pjotrscholtze/go-buildserver/models"
	"github.com/robfig/cron/v3"
)

type jobQueue struct {
	items        []models.Job
	finishedJobs []models.Job
	lock         sync.Locker
	buildRepo    PipelineRepo
	wm           *websocketmanager.WebsocketManager
	nextID       int64
}

type JobQueue interface {
	Run()
	AddQueueItem(repoName, buildReason, origin string)
	List() []*models.Job
	ListAllJobsOfPipeline(pipelineName string) []*models.Job
}

func (bq *jobQueue) ListAllJobsOfPipeline(pipelineName string) []*models.Job {
	out := []*models.Job{}
	bq.lock.Lock()
	defer bq.lock.Unlock()
	for i := range bq.items {
		if bq.items[i].RepoName == pipelineName {
			out = append(out, &bq.items[i])
		}
	}
	for i := range bq.finishedJobs {
		if bq.finishedJobs[i].RepoName == pipelineName {
			out = append(out, &bq.finishedJobs[i])
		}
	}

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
	bq.finishedJobs = append(bq.finishedJobs, *item)
	return item
}
func (bq *jobQueue) Run() {
	for {
		if item := bq.tick(); item != nil {
			bq.wm.BroadcastOnEndpoint("jobs", "", bq.items)
			bq.buildRepo.GetRepoByName(item.RepoName).Build(item.BuildReason, item.Origin, item.QueueTime.String())
			item.Status = "finished"
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
