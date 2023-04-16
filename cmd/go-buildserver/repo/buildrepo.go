package repo

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/config"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/process"
	"github.com/robfig/cron/v3"
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

type buildRepo struct {
	repos  []Repo
	config *config.Config
	cron   *cron.Cron
}
type BuildRepo interface {
	GetRepoByName(name string) Repo
	List() []Repo
}
type buildResultLine struct {
	line string
	pipe process.PipeType
	time time.Time
}
type BuildResultLine interface {
	Line() string
	Pipe() process.PipeType
	Time() time.Time
}

func (brl *buildResultLine) Line() string {
	return brl.line
}
func (brl *buildResultLine) Pipe() process.PipeType {
	return brl.pipe
}
func (brl *buildResultLine) Time() time.Time {
	return brl.time
}

type buildResult struct {
	lines     []buildResultLine
	reason    string
	starttime time.Time
	status    ResultStatus
}
type BuildResult interface {
	Lines() []BuildResultLine
	Reason() string
	Starttime() time.Time
	Status() ResultStatus
}

func (br *buildResult) Lines() []BuildResultLine {
	res := make([]BuildResultLine, len(br.lines))
	for i, _ := range br.lines {
		res[i] = &br.lines[i]
	}
	return res
}
func (br *buildResult) Reason() string {
	return br.reason
}
func (br *buildResult) Starttime() time.Time {
	return br.starttime
}
func (br *buildResult) Status() ResultStatus {
	return br.status
}

type repo struct {
	repo         config.Repo
	results      []*buildResult
	resultsMutex sync.Mutex
	buildRepo    *buildRepo
}
type Repo interface {
	Build(reason string)
	GetBuildScript() string
	ForceCleanBuild() bool
	GetName() string
	GetURL() string
	GetTriggers() []config.Trigger
	GetLastNBuildResults(n int) []BuildResult
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func (r *repo) init(cr *cron.Cron) {
	for _, trigger := range r.repo.Triggers {
		if trigger.Kind != "Cron" {
			continue
		}

		reason := "Cron: " + trigger.Schedule
		cr.AddFunc(trigger.Schedule, func() {
			r.Build(reason)
		})
	}

}
func (r *repo) GetLastNBuildResults(n int) []BuildResult {
	r.resultsMutex.Lock()
	defer r.resultsMutex.Unlock()

	if n < 0 {
		n = len(r.results)
	}
	res := make([]BuildResult, min(len(r.results), n))
	for i := 0; i < len(res); i++ {
		ind := len(r.results) - n + i
		res[i] = r.results[ind]
	}

	return res
}

func (r *repo) GetBuildScript() string {
	return r.repo.BuildScript
}

func (r *repo) ForceCleanBuild() bool {
	return r.repo.ForceCleanBuild
}

func (r *repo) GetName() string {
	return r.repo.Name
}

func (r *repo) GetURL() string {
	return r.repo.URL
}
func (r *repo) GetTriggers() []config.Trigger {
	return r.repo.Triggers
}

func (r *repo) printBuildStart(reason string) {
	log.Printf("Starting build for '%s', reason: %s", r.repo.Name, reason)
	log.Println("Build configuration:")
	log.Printf("- URL:%s\n", r.repo.URL)
	log.Printf("- Name:%s\n", r.repo.Name)
	log.Printf("- BuildScript:%s\n", r.repo.BuildScript)
	log.Printf("- ForceCleanBuild:%s\n", r.repo.ForceCleanBuild)
	log.Println("")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return false
}

func (r *repo) Build(reason string) {
	r.printBuildStart(reason)
	r.resultsMutex.Lock()
	os.MkdirAll(r.buildRepo.config.WorkspaceDirectory, 0777)

	repoPath := path.Join(r.buildRepo.config.WorkspaceDirectory, r.repo.Name)
	doClone := !fileExists(repoPath)
	if r.ForceCleanBuild() && !doClone {
		doClone = true
		os.RemoveAll(repoPath)
	}
	os.MkdirAll(repoPath, 0777)

	results := &buildResult{
		lines:     []buildResultLine{},
		reason:    reason,
		starttime: time.Now(),
		status:    RUNNING,
	}
	r.results = append(r.results, results)
	if len(r.results) > int(r.buildRepo.config.MaxHistoryInMemory) {
		r.results[0].lines = nil
		r.results = r.results[1:]
	}

	r.resultsMutex.Unlock()
	gitPath := path.Join(repoPath, r.repo.Name)

	f, err := os.Create(path.Join(repoPath, "boot.sh"))
	defer f.Close()
	if err != nil {
		results.lines = append(results.lines, []buildResultLine{
			{
				line: "Failed to create boot script:",
				pipe: process.STDERR,
				time: time.Now(),
			},
			{
				line: err.Error(),
				pipe: process.STDERR,
				time: time.Now(),
			},
		}...)
		fmt.Println(err)
		return
	}
	_, err = f.WriteString(strings.Join([]string{
		"#!/bin/sh",
		"eval `ssh-agent`",
		"ssh-agent &",
		"ssh-add " + (*r).repo.SSHKeyLocation,
		"git clone " + r.repo.URL + " " + gitPath,
		"chmod +x " + path.Join(gitPath, r.repo.BuildScript),
		path.Join(gitPath, r.repo.BuildScript),
		"pkill ssh-agent",
	}, "\n"))
	if err != nil {
		results.lines = append(results.lines, []buildResultLine{
			{
				line: "Failed to write boot script:",
				pipe: process.STDERR,
				time: time.Now(),
			},
			{
				line: err.Error(),
				pipe: process.STDERR,
				time: time.Now(),
			},
		}...)
		fmt.Println(err)
		return
	}

	process.StartProcessWithStdErrStdOutCallback(
		"/bin/sh",
		[]string{path.Join(repoPath, "boot.sh")},
		func(pt process.PipeType, t time.Time, s string) {
			r.resultsMutex.Lock()
			results.lines = append(results.lines, buildResultLine{
				line: s,
				pipe: pt,
				time: t,
			})
			r.resultsMutex.Unlock()
		})
	results.status = FINISHED
}

func (br *buildRepo) GetRepoByName(name string) Repo {
	for _, repo := range br.repos {
		if repo.GetName() == name {
			return repo
		}
	}
	log.Fatalln("Repo not found!") // @todo decent error handling here!
	return &repo{repo: br.config.Repos[0]}
}
func (br *buildRepo) List() []Repo {
	return br.repos
}

func NewBuildRepo(config *config.Config, cr *cron.Cron) BuildRepo {
	br := &buildRepo{
		config: config,
		cron:   cr,
	}
	res := make([]Repo, len(br.config.Repos))
	for i, elem := range br.config.Repos {
		r := repo{repo: elem, buildRepo: br}
		res[i] = &r
		r.init(cr)
	}
	br.repos = res
	return br
}
