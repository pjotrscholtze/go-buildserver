package entity

import (
	"time"

	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/websocketmanager"
	"github.com/pjotrscholtze/go-buildserver/models"
)

type BuildResult struct {
	PipelineName     string
	Lines            []BuildResultLine
	Reason           string
	Starttime        time.Time
	Status           ResultStatus
	Websocketmanager *websocketmanager.WebsocketManager
	Job              *models.Job
}

func NewBuildResult(PipelineName string, Lines []BuildResultLine, Reason string, Starttime time.Time, Status ResultStatus, Job *models.Job) BuildResult {
	return BuildResult{
		PipelineName: PipelineName,
		Lines:        Lines,
		Reason:       Reason,
		Starttime:    Starttime,
		Status:       Status,
		Job:          Job,
	}
}
