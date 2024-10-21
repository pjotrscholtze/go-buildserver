package repo

import (
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type buildQueueItem struct {
	RepoName    string
	BuildReason string
	Origin      string
	QueueTime   time.Time
}
type buildQueue struct {
	items     []buildQueueItem
	lock      sync.Locker
	buildRepo BuildRepo
}

type BuildQueue interface {
	Run()
	AddQueueItem(repoName, buildReason, origin string)
}

func (bq *buildQueue) tick() *buildQueueItem {
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
	bq.items = append(bq.items, buildQueueItem{
		RepoName:    repoName,
		BuildReason: buildReason,
		Origin:      origin,
		QueueTime:   time.Now(),
	})
}

func NewBuildQueue(buildRepo BuildRepo, cr *cron.Cron) BuildQueue {
	bq := &buildQueue{
		items:     []buildQueueItem{},
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
