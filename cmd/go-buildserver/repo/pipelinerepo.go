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
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/util"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/websocketmanager"
)

type ResultStatus string

const (
	PENDING  ResultStatus = "PENDING"
	RUNNING               = "RUNNING"
	FINISHED              = "FINISHED"
	ERROR                 = "ERROR"
)

type resultLine struct {
	pipeType  process.PipeType
	timestamp time.Time
	line      string
}
type ResultLine interface {
	GetPipeType() process.PipeType
	GetTimestamp() time.Time
	GetLine() string
}

func (rl *resultLine) GetPipeType() process.PipeType {
	return rl.pipeType
}
func (rl *resultLine) GetTimestamp() time.Time {
	return rl.timestamp
}
func (rl *resultLine) GetLine() string {
	return rl.line
}

type pipelineRepo struct {
	pipelines        []Pipeline
	config           *config.Config
	websocketmanager *websocketmanager.WebsocketManager
}
type PipelineRepo interface {
	GetRepoByName(name string) Pipeline
	GetRepoBySlug(name string) Pipeline
	List() []Pipeline
}
type BuildResultLine struct {
	Line string
	pipe process.PipeType
	Pipe string
	Time time.Time
}

type BuildResult struct {
	Lines            []BuildResultLine
	Reason           string
	Starttime        time.Time
	Status           ResultStatus
	Websocketmanager *websocketmanager.WebsocketManager
}

func (br *BuildResult) addLines(lines []BuildResultLine) {
	for i := range lines {
		br.addLine(lines[i])
	}
}
func (br *BuildResult) addLine(line BuildResultLine) {
	br.Lines = append(br.Lines, line)
	br.Websocketmanager.BroadcastOnEndpoint("repo", "Go-Buildserver_Repo_clone_example", br)
}

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return false
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

func (br *pipelineRepo) GetRepoByName(name string) Pipeline {
	for _, repo := range br.pipelines {
		if repo.GetName() == name {
			return repo
		}
	}
	log.Fatalln("Repo not found!") // @todo decent error handling here!
	return &pipeline{pipeline: br.config.Pipelines[0]}
}
func (br *pipelineRepo) GetRepoBySlug(name string) Pipeline {
	for _, repo := range br.pipelines {
		if util.StringToSlug(repo.GetName()) == name {
			return repo
		}
	}
	log.Fatalln("Repo not found!") // @todo decent error handling here!
	return &pipeline{pipeline: br.config.Pipelines[0]}
}
func (br *pipelineRepo) List() []Pipeline {
	return br.pipelines
}

func NewPipelineRepo(config *config.Config, wm *websocketmanager.WebsocketManager) PipelineRepo {
	br := &pipelineRepo{
		config:           config,
		websocketmanager: wm,
	}
	res := make([]Pipeline, len(br.config.Pipelines))
	for i, elem := range br.config.Pipelines {
		r := pipeline{pipeline: elem, buildRepo: br}
		res[i] = &r
	}
	br.pipelines = res
	return br
}