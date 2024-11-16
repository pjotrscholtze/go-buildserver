package repo

import (
	"strconv"
	"time"

	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/process"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/websocketmanager"
	"github.com/pjotrscholtze/go-buildserver/models"
)

type BuildResultLine struct {
	Line string
	pipe process.PipeType
	Pipe string
	Time time.Time
}
type BuildResult struct {
	PipelineName     string
	Lines            []BuildResultLine
	Reason           string
	Starttime        time.Time
	Status           ResultStatus
	Websocketmanager *websocketmanager.WebsocketManager
	Job              *models.Job
}

func (br *BuildResult) addLines(lines []BuildResultLine) {
	for i := range lines {
		br.addLine(lines[i])
	}
}
func (br *BuildResult) addLine(line BuildResultLine) {
	br.Lines = append(br.Lines, line)
	br.Websocketmanager.BroadcastOnEndpoint("build", strconv.FormatInt(br.Job.ID, 10), br)
	br.Websocketmanager.BroadcastOnEndpoint("repo-build-live", (*(*br).Job).RepoName, br)
}
