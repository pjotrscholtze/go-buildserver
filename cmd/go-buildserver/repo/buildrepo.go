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

type buildRepo struct {
	repos            []Repo
	config           *config.Config
	websocketmanager *websocketmanager.WebsocketManager
}
type BuildRepo interface {
	GetRepoByName(name string) Repo
	GetRepoBySlug(name string) Repo
	List() []Repo
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

type repo struct {
	repo         config.Repo
	results      []*BuildResult
	resultsMutex sync.Mutex
	buildRepo    *buildRepo
}
type Repo interface {
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
func (r *repo) GetLastNBuildResults(n int) []BuildResult {
	r.resultsMutex.Lock()
	defer r.resultsMutex.Unlock()

	if n < 0 {
		n = len(r.results)
	}
	res := make([]BuildResult, min(len(r.results), n))
	for i := 0; i < len(res); i++ {
		ind := len(r.results) - n + i
		res[i] = *r.results[ind]
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

func (r *repo) GetPath() string {
	return r.repo.Path
}

func (r *repo) GetURL() string {
	return r.repo.URL
}
func (r *repo) GetTriggers() []config.Trigger {
	return r.repo.Triggers
}

func (r *repo) GetTriggersOfKind(filterKind string) []config.Trigger {
	triggers := []config.Trigger{}
	for _, trigger := range r.repo.Triggers {
		if trigger.Kind != filterKind {
			continue
		}
		triggers = append(triggers, trigger)
	}
	return triggers
}

func (r *repo) printBuildStart(reason, origin, queueTime string) {
	isRepoBased := len(r.GetURL()) > 0
	log.Printf("Starting build for '%s', reason: %s, origin: %s, queuetime: %s", r.repo.Name, reason, origin, queueTime)
	log.Println("Build configuration:")
	log.Printf("- Is repo based:%s\n", strconv.FormatBool(isRepoBased))
	if isRepoBased {
		log.Printf("- URL:%s\n", r.repo.URL)
	} else {
		log.Printf("- Path:%s\n", r.repo.Path)
	}
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

func (r *repo) Build(reason, origin, queueTime string) {
	r.printBuildStart(reason, origin, queueTime)
	r.resultsMutex.Lock()
	os.MkdirAll(r.buildRepo.config.WorkspaceDirectory, 0777)

	isRepoBased := len(r.GetURL()) > 0
	repoPath := path.Join(r.buildRepo.config.WorkspaceDirectory, r.repo.Name)
	doClone := !fileExists(repoPath)
	if isRepoBased && r.ForceCleanBuild() && !doClone {
		doClone = true
		os.RemoveAll(repoPath)
	}
	os.MkdirAll(repoPath, 0777)

	results := &BuildResult{
		Lines:            []BuildResultLine{},
		Reason:           reason,
		Starttime:        time.Now(),
		Status:           RUNNING,
		Websocketmanager: r.buildRepo.websocketmanager,
	}
	r.results = append(r.results, results)
	if len(r.results) > int(r.buildRepo.config.MaxHistoryInMemory) {
		r.results[0].Lines = nil
		r.results = r.results[1:]
	}

	r.resultsMutex.Unlock()
	gitPath := path.Join(repoPath, r.repo.Name)

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
		jobPath = r.GetPath()
	}
	if isRepoBased {
		bootScript = append(bootScript, []string{
			"eval `ssh-agent`",
			"ssh-agent &",
			"ssh-add " + (*r).repo.SSHKeyLocation,
			"git clone --depth 1 " + r.repo.URL + " " + gitPath,
		}...)
	}

	bootScript = append(bootScript, []string{
		"chmod +x " + path.Join(jobPath, r.repo.BuildScript),
		path.Join(jobPath, r.repo.BuildScript),
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
			r.resultsMutex.Lock()
			results.addLine(BuildResultLine{
				Line: s,
				pipe: pt,
				Pipe: map[process.PipeType]string{
					process.STDOUT: "STDOUT",
					process.STDERR: "STDERR",
				}[pt],
				Time: t,
			})
			r.resultsMutex.Unlock()
		})
	results.Status = FINISHED
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
func (br *buildRepo) GetRepoBySlug(name string) Repo {
	for _, repo := range br.repos {
		if util.StringToSlug(repo.GetName()) == name {
			return repo
		}
	}
	log.Fatalln("Repo not found!") // @todo decent error handling here!
	return &repo{repo: br.config.Repos[0]}
}
func (br *buildRepo) List() []Repo {
	return br.repos
}

func NewBuildRepo(config *config.Config, wm *websocketmanager.WebsocketManager) BuildRepo {
	br := &buildRepo{
		config:           config,
		websocketmanager: wm,
	}
	res := make([]Repo, len(br.config.Repos))
	for i, elem := range br.config.Repos {
		r := repo{repo: elem, buildRepo: br}
		res[i] = &r
	}
	br.repos = res
	return br
}
