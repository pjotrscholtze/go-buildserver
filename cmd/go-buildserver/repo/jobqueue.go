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
	lock      sync.Locker
	buildRepo PipelineRepo
	wm        *websocketmanager.WebsocketManager
	dbRepo    DatabaseRepo
}

type JobQueue interface {
	Run()
	GetJobById(buildId int64) *models.Job
	AddQueueItem(repoName, buildReason, origin string)
	List() []*models.Job
	ListAllJobsOfPipeline(pipelineName string) []*models.Job
}

func (bq *jobQueue) GetJobById(buildId int64) *models.Job {
	job, _ := bq.dbRepo.GetJobByID(buildId)
	return job
}

func (bq *jobQueue) ListAllJobsOfPipeline(pipelineName string) []*models.Job {
	bq.lock.Lock()
	defer bq.lock.Unlock()
	return bq.listAllJobsOfPipelineUnsafe(pipelineName)
}
func (bq *jobQueue) listAllJobsOfPipelineUnsafe(pipelineName string) []*models.Job {
	jobs, _ := bq.dbRepo.ListAllJobsOfPipeline(pipelineName)
	out := []*models.Job{}
	for i := range jobs {
		out = append(out, &jobs[i])
	}

	return out
}
func (bq *jobQueue) List() []*models.Job {
	out := []*models.Job{}
	bq.lock.Lock()
	defer bq.lock.Unlock()
	jobs, _ := bq.dbRepo.ListJobByStatus("pending")
	for i := range jobs {
		out = append(out, &jobs[i])
	}

	return out
}

func (bq *jobQueue) tick() *models.Job {
	bq.lock.Lock()
	defer bq.lock.Unlock()
	jobs, _ := bq.dbRepo.ListJobByStatus("pending")
	if len(jobs) == 0 {
		return nil
	}
	bq.dbRepo.UpdateJobStatusByID(jobs[0].ID, "running")

	return &jobs[0]
}
func (bq *jobQueue) Run() {
	for {
		if item := bq.tick(); item != nil {
			jobs, _ := bq.dbRepo.ListJobByStatus("pending")
			bq.wm.BroadcastOnEndpoint("jobs", "", jobs)
			bq.wm.BroadcastOnEndpoint("repo-builds", item.RepoName, struct {
				Jobs []*models.Job
			}{
				Jobs: bq.ListAllJobsOfPipeline(item.RepoName),
			})

			bq.buildRepo.GetRepoByName(item.RepoName).Build(item)
			bq.dbRepo.UpdateJobStatusByID(item.ID, "finished")

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
	bq.dbRepo.AddJob(models.Job{
		RepoName:    repoName,
		BuildReason: buildReason,
		Origin:      origin,
		QueueTime:   strfmt.DateTime(time.Now()),
		Status:      "pending",
	})
	jobs, _ := bq.dbRepo.ListJobByStatus("pending")

	bq.wm.BroadcastOnEndpoint("jobs", "", jobs)
	bq.wm.BroadcastOnEndpoint("repo-builds", repoName, struct {
		Jobs []*models.Job
	}{
		Jobs: bq.listAllJobsOfPipelineUnsafe(repoName),
	})
}

func NewJobQueue(buildRepo PipelineRepo, cr *cron.Cron, wm *websocketmanager.WebsocketManager, dbRepo DatabaseRepo) JobQueue {
	bq := &jobQueue{
		lock:      &sync.Mutex{},
		buildRepo: buildRepo,
		wm:        wm,
		dbRepo:    dbRepo,
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
