package repo

import (
	"log"
	"os"

	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/config"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/util"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/websocketmanager"
)

type pipelineRepo struct {
	pipelines        []Pipeline
	config           *config.Config
	websocketmanager *websocketmanager.WebsocketManager
	buildResultRepo  BuildResultRepo
}
type PipelineRepo interface {
	GetRepoByName(name string) Pipeline
	GetRepoBySlug(name string) Pipeline
	List() []Pipeline
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return false
}

func (br *pipelineRepo) GetRepoByName(name string) Pipeline {
	for _, repo := range br.pipelines {
		if repo.GetPipelineConfig().Name == name {
			return repo
		}
	}
	log.Fatalln("Repo not found!") // @todo decent error handling here!
	return nil
}
func (br *pipelineRepo) GetRepoBySlug(name string) Pipeline {
	for _, repo := range br.pipelines {
		if util.StringToSlug(repo.GetPipelineConfig().Name) == name {
			return repo
		}
	}
	log.Fatalln("Repo not found!") // @todo decent error handling here!
	return nil
}
func (br *pipelineRepo) List() []Pipeline {
	return br.pipelines
}

func NewPipelineRepo(config *config.Config, wm *websocketmanager.WebsocketManager, buildResultRepo BuildResultRepo, db DatabaseRepo) PipelineRepo {
	br := &pipelineRepo{
		config:           config,
		websocketmanager: wm,
		buildResultRepo:  buildResultRepo,
	}
	res := make([]Pipeline, len(br.config.Pipelines))
	for i, elem := range br.config.Pipelines {
		res[i] = NewPipeline(elem, br, br.buildResultRepo, db)
	}
	br.pipelines = res
	return br
}
