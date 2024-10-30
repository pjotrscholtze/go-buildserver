package repo

import (
	"time"

	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/process"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/websocketmanager"
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
