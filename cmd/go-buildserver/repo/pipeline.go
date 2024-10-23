package repo

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/config"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/process"
)

type pipeline struct {
	pipeline     config.Pipeline
	results      []*BuildResult
	resultsMutex sync.Mutex
	buildRepo    *pipelineRepo
}

type Pipeline interface {
	Build(reason, origin, queueTime string)
	GetBuildScript() string
	ForceCleanBuild() bool
	GetName() string
	GetPath() string
	GetURL() string
	GetTriggers() []config.Trigger
	GetTriggersOfKind(filterKind string) []config.Trigger
	GetLastNBuildResults(n int) []BuildResult
}

func (p *pipeline) GetLastNBuildResults(n int) []BuildResult {
	p.resultsMutex.Lock()
	defer p.resultsMutex.Unlock()

	if n < 0 {
		n = len(p.results)
	}
	res := make([]BuildResult, min(len(p.results), n))
	for i := 0; i < len(res); i++ {
		ind := len(p.results) - n + i
		res[i] = *p.results[ind]
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
func (p *pipeline) Build(reason, origin, queueTime string) {
	p.printBuildStart(reason, origin, queueTime)
	p.resultsMutex.Lock()
	os.MkdirAll(p.buildRepo.config.WorkspaceDirectory, 0777)

	isRepoBased := len(p.GetURL()) > 0
	repoPath := path.Join(p.buildRepo.config.WorkspaceDirectory, p.pipeline.Name)
	doClone := !fileExists(repoPath)
	if isRepoBased && p.ForceCleanBuild() && !doClone {
		doClone = true
		os.RemoveAll(repoPath)
	}
	os.MkdirAll(repoPath, 0777)

	results := &BuildResult{
		Lines:            []BuildResultLine{},
		Reason:           reason,
		Starttime:        time.Now(),
		Status:           RUNNING,
		Websocketmanager: p.buildRepo.websocketmanager,
	}
	p.results = append(p.results, results)
	if len(p.results) > int(p.buildRepo.config.MaxHistoryInMemory) {
		p.results[0].Lines = nil
		p.results = p.results[1:]
	}

	p.resultsMutex.Unlock()
	gitPath := path.Join(repoPath, p.pipeline.Name)

	f, err := os.Create(path.Join(repoPath, "boot.sh"))
	defer f.Close()
	if err != nil {
		results.addLines([]BuildResultLine{
			{
				Line: "Failed to create boot script:",
				pipe: process.STDERR,
				Pipe: "STDERR",
				Time: time.Now(),
			},
			{
				Line: err.Error(),
				pipe: process.STDERR,
				Pipe: "STDERR",
				Time: time.Now(),
			},
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
		results.addLines([]BuildResultLine{
			{
				Line: "Failed to write boot script:",
				pipe: process.STDERR,
				Pipe: "STDERR",
				Time: time.Now(),
			},
			{
				Line: err.Error(),
				pipe: process.STDERR,
				Pipe: "STDERR",
				Time: time.Now(),
			},
		})
		fmt.Println(err)
		return
	}

	process.StartProcessWithStdErrStdOutCallback(
		"/bin/sh",
		[]string{path.Join(repoPath, "boot.sh")},
		func(pt process.PipeType, t time.Time, s string) {
			p.resultsMutex.Lock()
			results.addLine(BuildResultLine{
				Line: s,
				pipe: pt,
				Pipe: map[process.PipeType]string{
					process.STDOUT: "STDOUT",
					process.STDERR: "STDERR",
				}[pt],
				Time: t,
			})
			p.resultsMutex.Unlock()
		})
	results.Status = FINISHED
}
