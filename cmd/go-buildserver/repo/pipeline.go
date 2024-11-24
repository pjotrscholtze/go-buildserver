package repo

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/config"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/entity"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/process"
	"github.com/pjotrscholtze/go-buildserver/models"
)

type pipeline struct {
	pipeline        config.Pipeline
	buildRepo       *pipelineRepo
	buildResultRepo BuildResultRepo
	db              DatabaseRepo
}

func NewPipeline(pl config.Pipeline, buildRepo *pipelineRepo, buildResultRepo BuildResultRepo, db DatabaseRepo) Pipeline {
	return &pipeline{
		pipeline:        pl,
		buildRepo:       buildRepo,
		buildResultRepo: buildResultRepo,
		db:              db,
	}
}

type Pipeline interface {
	Build(job *models.Job)
	GetBuildScript() string
	ForceCleanBuild() bool
	GetName() string
	GetPath() string
	GetURL() string
	GetTriggers() []config.Trigger
	GetTriggersOfKind(filterKind string) []config.Trigger
	GetLastNBuildResults(n int) []entity.BuildResult
	GetBuildResultForJobID(job *models.Job) *entity.BuildResult
}

func (p *pipeline) GetBuildResultForJobID(job *models.Job) *entity.BuildResult {
	br, err := p.buildResultRepo.GetBuildResultForJobID(job.ID)
	if err != nil {
		return nil
	}
	return br
}

func (p *pipeline) GetLastNBuildResults(n int) []entity.BuildResult {
	jobs, err := p.db.ListNLastJobsOfPipeline(p.GetName(), n)
	if err != nil {
		return nil
	}
	res := []entity.BuildResult{}
	for _, job := range jobs {
		br, err := p.db.GetBuildResult(job.ID)
		if err != nil {
			return nil
		}
		res = append(res, *br)
	}

	return res
}

func (p *pipeline) GetBuildScript() string {
	return p.pipeline.BuildScript
}

func (p *pipeline) ForceCleanBuild() bool {
	return p.pipeline.ForceCleanBuild
}

func (p *pipeline) GetName() string {
	return p.pipeline.Name
}

func (p *pipeline) GetPath() string {
	return p.pipeline.Path
}

func (p *pipeline) GetURL() string {
	return p.pipeline.URL
}
func (p *pipeline) GetTriggers() []config.Trigger {
	return p.pipeline.Triggers
}

func (p *pipeline) GetTriggersOfKind(filterKind string) []config.Trigger {
	triggers := []config.Trigger{}
	for _, trigger := range p.pipeline.Triggers {
		if trigger.Kind != filterKind {
			continue
		}
		triggers = append(triggers, trigger)
	}
	return triggers
}

func (p *pipeline) printBuildStart(reason, origin, queueTime string) {
	isRepoBased := len(p.GetURL()) > 0
	log.Printf("Starting build for '%s', reason: %s, origin: %s, queuetime: %s", p.pipeline.Name, reason, origin, queueTime)
	log.Println("Build configuration:")
	log.Printf("- Is repo based:%s\n", strconv.FormatBool(isRepoBased))
	if isRepoBased {
		log.Printf("- URL:%s\n", p.pipeline.URL)
	} else {
		log.Printf("- Path:%s\n", p.pipeline.Path)
	}
	log.Printf("- Name:%s\n", p.pipeline.Name)
	log.Printf("- BuildScript:%s\n", p.pipeline.BuildScript)
	log.Printf("- ForceCleanBuild:%s\n", p.pipeline.ForceCleanBuild)
	log.Println("")
}
func (p *pipeline) Build(job *models.Job) {
	if job == nil {
		// @todo error handling here.
	}
	reason, origin, queueTime := job.BuildReason, job.Origin, job.QueueTime.String()
	p.printBuildStart(reason, origin, queueTime)
	os.MkdirAll(p.buildRepo.config.WorkspaceDirectory, 0777)

	isRepoBased := len(p.GetURL()) > 0
	repoPath := path.Join(p.buildRepo.config.WorkspaceDirectory, p.pipeline.Name)
	doClone := !fileExists(repoPath)
	if isRepoBased && p.ForceCleanBuild() && !doClone {
		doClone = true
		os.RemoveAll(repoPath)
	}
	os.MkdirAll(repoPath, 0777)

	gitPath := path.Join(repoPath, p.pipeline.Name)

	f, err := os.Create(path.Join(repoPath, "boot.sh"))
	defer f.Close()
	if err != nil {
		p.buildResultRepo.AddLines(job.ID, []entity.BuildResultLine{
			entity.NewBuildResultLine(
				"Failed to create boot script:",
				process.STDERR,
				time.Now(),
			),
			entity.NewBuildResultLine(
				err.Error(),
				process.STDERR,
				time.Now(),
			),
		})
		fmt.Println(err)
		return
	}

	bootScript := []string{"#!/bin/sh"}
	jobPath := gitPath
	if !isRepoBased {
		jobPath = p.GetPath()
	}
	if isRepoBased {
		bootScript = append(bootScript, []string{
			"eval `ssh-agent`",
			"ssh-agent &",
			"ssh-add " + (*p).pipeline.SSHKeyLocation,
			"git clone --depth 1 " + p.pipeline.URL + " " + gitPath,
		}...)
	}

	bootScript = append(bootScript, []string{
		"chmod +x " + path.Join(jobPath, p.pipeline.BuildScript),
		path.Join(jobPath, p.pipeline.BuildScript),
	}...)
	if isRepoBased {
		bootScript = append(bootScript, "pkill ssh-agent")
	}

	_, err = f.WriteString(strings.Join(bootScript, "\n"))
	if err != nil {
		p.buildResultRepo.AddLines(job.ID, []entity.BuildResultLine{
			entity.NewBuildResultLine(
				"Failed to write boot script:",
				process.STDERR,
				time.Now(),
			),
			entity.NewBuildResultLine(
				err.Error(),
				process.STDERR,
				time.Now(),
			),
		})
		fmt.Println(err)
		return
	}

	process.StartProcessWithStdErrStdOutCallback(
		"/bin/sh",
		[]string{path.Join(repoPath, "boot.sh")},
		func(pt process.PipeType, t time.Time, s string) {
			p.buildResultRepo.AddLine(job.ID, entity.NewBuildResultLine(s, pt, t))
		})
	p.buildResultRepo.SetStatus(job.ID, entity.FINISHED)
}
