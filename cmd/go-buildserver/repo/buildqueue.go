package repo

import (
	"sync"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/pjotrscholtze/go-buildserver/models"
	"github.com/robfig/cron/v3"
)

//	type buildQueueItem struct {
//		RepoName    string
//		BuildReason string
//		Origin      string
//		QueueTime   time.Time
//	}
type buildQueue struct {
	items     []models.Job
	lock      sync.Locker
	buildRepo BuildRepo
}

type BuildQueue interface {
	Run()
	AddQueueItem(repoName, buildReason, origin string)
	List() []*models.Job
}

func (bq *buildQueue) List() []*models.Job {
	out := []*models.Job{}
	bq.lock.Lock()
	defer bq.lock.Unlock()
	for i := range bq.items {
		out = append(out, &bq.items[i])
	}

	return out
}

func (bq *buildQueue) tick() *models.Job {
	bq.lock.Lock()
	defer bq.lock.Unlock()
	if len(bq.items) == 0 {
		return nil
	}
	item := &bq.items[0]
	bq.items = bq.items[1:]
	return item
}
func (bq *buildQueue) Run() {
	for {
		if item := bq.tick(); item != nil {
			bq.buildRepo.GetRepoByName(item.RepoName).Build(item.BuildReason, item.Origin, item.QueueTime.String())
		}
		time.Sleep(50 * time.Millisecond)
	}
}
func (bq *buildQueue) AddQueueItem(repoName, buildReason, origin string) {
	bq.lock.Lock()
	defer bq.lock.Unlock()
	bq.items = append(bq.items, models.Job{
		RepoName:    repoName,
		BuildReason: buildReason,
		Origin:      origin,
		QueueTime:   strfmt.DateTime(time.Now()),
	})
}

func NewBuildQueue(buildRepo BuildRepo, cr *cron.Cron) BuildQueue {
	bq := &buildQueue{
		items:     []models.Job{},
		lock:      &sync.Mutex{},
		buildRepo: buildRepo,
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
