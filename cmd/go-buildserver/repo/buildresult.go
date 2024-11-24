package repo

import (
	"strconv"

	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/entity"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/websocketmanager"
)

// type BuildResult struct {
// 	PipelineName     string
// 	Lines            []entity.BuildResultLine
// 	Reason           string
// 	Starttime        time.Time
// 	Status           entity.ResultStatus
// 	Websocketmanager *websocketmanager.WebsocketManager
// 	Job              *models.Job
// }

type buildResultRepo struct {
	db               DatabaseRepo
	Websocketmanager *websocketmanager.WebsocketManager
}
type BuildResultRepo interface {
	AddLines(jobID int64, lines []entity.BuildResultLine) error
	AddLine(jobID int64, line entity.BuildResultLine) error
	SetStatus(jobID int64, line entity.ResultStatus) error
	GetBuildResultForJobID(jobID int64) (*entity.BuildResult, error)
}

func (brr *buildResultRepo) GetBuildResultForJobID(jobID int64) (*entity.BuildResult, error) {
	// p.buildResultRepo.GetBuildResultForJobID(job.ID)
	return brr.db.GetBuildResult(jobID)
}

func (brr *buildResultRepo) SetStatus(jobID int64, line entity.ResultStatus) error {
	err := brr.db.UpdateJobStatusByID(jobID, string(line))
	br, err := brr.db.GetBuildResult(jobID)
	if err != nil {
		return err
	}
	brr.Websocketmanager.BroadcastOnEndpoint("build", strconv.FormatInt(jobID, 10), br)
	return err
}
func (brr *buildResultRepo) AddLines(jobID int64, lines []entity.BuildResultLine) error {
	for i := range lines {
		err := brr.AddLine(jobID, lines[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (brr *buildResultRepo) AddLine(jobID int64, line entity.BuildResultLine) error {
	// br.Lines = append(br.Lines, line)
	err := brr.db.AddBuildResultLine(jobID, line)
	if err != nil {
		return err
	}
	br, err := brr.db.GetBuildResult(jobID)
	if err != nil {
		return err
	}
	brr.Websocketmanager.BroadcastOnEndpoint("build", strconv.FormatInt(br.Job.ID, 10), *br)
	brr.Websocketmanager.BroadcastOnEndpoint("repo-build-live", (*br).PipelineName, *br)
	return nil
}
func NewBuildResultRepo(db DatabaseRepo, Websocketmanager *websocketmanager.WebsocketManager) BuildResultRepo {
	return &buildResultRepo{
		db:               db,
		Websocketmanager: Websocketmanager,
	}
}
