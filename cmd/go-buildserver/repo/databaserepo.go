package repo

import (
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/jmoiron/sqlx"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/entity"
	"github.com/pjotrscholtze/go-buildserver/models"
)

type internalJob struct {
	ID           interface{}
	Buildreason  interface{}
	Origin       interface{}
	Queuetime    interface{}
	Pipelinename interface{}
	Status       interface{}
	Starttime    interface{}
}

type databaseRepo struct {
	db *sqlx.DB
}

func (dr *databaseRepo) AddJob(job models.Job) error {
	st, err := dr.db.PrepareNamed("INSERT INTO job " +
		"(BuildReason, Origin, QueueTime, Pipelinename, Status) VALUES " +
		"(:BuildReason, :Origin, :QueueTime, :RepoName, :Status)")

	if err != nil {
		return err
	}
	defer st.Close()

	_, err = st.Exec(map[string]interface{}{
		"BuildReason": job.BuildReason,
		"Origin":      job.Origin,
		"QueueTime":   job.QueueTime,
		"RepoName":    job.RepoName,
		"Status":      job.Status,
	})

	return err
}
func (dr *databaseRepo) ListNLastJobsOfPipeline(pipelineName string, n int) ([]models.Job, error) {
	out := []models.Job{}
	st, err := dr.db.PrepareNamed("SELECT * FROM job WHERE Pipelinename = :Pipelinename LIMIT :n")

	if err != nil {
		return out, err
	}
	defer st.Close()

	jobs := []internalJob{}
	args := map[string]interface{}{
		"Pipelinename": pipelineName,
		"n":            n,
	}
	err = st.Select(&jobs, args)

	if err == nil {
		for _, row := range jobs {
			out = append(out, models.Job{
				BuildReason: row.Buildreason.(string),
				ID:          row.ID.(int64),
				Origin:      row.Origin.(string),
				QueueTime:   strfmt.DateTime(row.Queuetime.(time.Time)),
				RepoName:    row.Pipelinename.(string),
				Status:      row.Status.(string),
			})
		}
	}

	return out, err
}
func (dr *databaseRepo) ListAllJobsOfPipeline(pipelineName string) ([]models.Job, error) {
	out := []models.Job{}
	st, err := dr.db.PrepareNamed("SELECT * FROM job WHERE Pipelinename = :Pipelinename")

	if err != nil {
		return out, err
	}
	defer st.Close()

	jobs := []internalJob{}
	args := map[string]interface{}{
		"Pipelinename": pipelineName,
	}
	err = st.Select(&jobs, args)

	if err == nil {
		for _, row := range jobs {
			out = append(out, models.Job{
				BuildReason: row.Buildreason.(string),
				ID:          row.ID.(int64),
				Origin:      row.Origin.(string),
				QueueTime:   strfmt.DateTime(row.Queuetime.(time.Time)),
				RepoName:    row.Pipelinename.(string),
				Status:      row.Status.(string),
			})
		}
	}

	return out, err
}

func (dr *databaseRepo) ListJobByStatus(status string) ([]models.Job, error) {
	out := []models.Job{}
	st, err := dr.db.PrepareNamed("SELECT * FROM job WHERE status = :status")

	if err != nil {
		return out, err
	}
	defer st.Close()

	jobs := []internalJob{}
	args := map[string]interface{}{
		"status": status,
	}
	err = st.Select(&jobs, args)

	if err == nil {
		for _, row := range jobs {
			out = append(out, models.Job{
				BuildReason: row.Buildreason.(string),
				ID:          row.ID.(int64),
				Origin:      row.Origin.(string),
				QueueTime:   strfmt.DateTime(row.Queuetime.(time.Time)),
				RepoName:    row.Pipelinename.(string),
				Status:      row.Status.(string),
			})
		}
	}

	return out, err
}

func (dr *databaseRepo) GetJobByID(ID int64) (*models.Job, error) {
	st, err := dr.db.PrepareNamed("SELECT * FROM job WHERE id = :id")

	if err != nil {
		return nil, err
	}
	defer st.Close()

	job := internalJob{}
	err = st.Get(&job, map[string]interface{}{
		"id": ID,
	})

	if err != nil {
		return nil, err
	}

	return &models.Job{
		BuildReason: job.Buildreason.(string),
		ID:          job.ID.(int64),
		Origin:      job.Origin.(string),
		QueueTime:   strfmt.DateTime(job.Queuetime.(time.Time)),
		RepoName:    job.Pipelinename.(string),
		Status:      job.Status.(string),
	}, nil
}

func (dr *databaseRepo) UpdateJobStatusByID(ID int64, status string) error {
	st, err := dr.db.PrepareNamed("UPDATE job SET status=:status WHERE id=:id")
	if err != nil {
		return err
	}
	defer st.Close()

	args := map[string]interface{}{}
	args["id"] = ID
	args["status"] = status

	_, err = st.Exec(args)
	return err
}

func (dr *databaseRepo) AddBuildResultLine(jobID int64, line entity.BuildResultLine) error {
	st, err := dr.db.PrepareNamed("INSERT INTO buildresultline " +
		"(jobid, line, pipe, time) VALUES " +
		"(:jobid, :line, :pipe, :time)")

	if err != nil {
		return err
	}
	defer st.Close()

	_, err = st.Exec(map[string]interface{}{
		"jobid": jobID,
		"line":  line.Line,
		"pipe":  line.Pipe,
		"time":  line.Time,
	})

	return err
}
func (dr *databaseRepo) GetBuildResult(jobID int64) (*entity.BuildResult, error) {
	job, err := dr.GetJobByID(jobID)
	if err != nil {
		return nil, err
	}
	st, err := dr.db.PrepareNamed("SELECT line, pipe, time FROM buildresultline WHERE jobid=:jobid")

	if err != nil {
		return nil, err
	}
	defer st.Close()

	lines := []entity.BuildResultLine{}
	args := map[string]interface{}{
		"jobid": jobID,
	}
	rows, err := st.Queryx(args)
	// err = st.Select(&lines, args)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		mres := map[string]interface{}{}
		err = rows.MapScan(mres)
		lines = append(lines, entity.NewBuildResultLinePipeString(
			mres["line"].(string),
			mres["pipe"].(string),
			mres["time"].(time.Time),
		))
	}
	var startTime time.Time
	if len(lines) > 0 {
		startTime = lines[0].Time
	}

	br := entity.NewBuildResult(
		job.RepoName,
		lines,
		job.BuildReason,
		startTime,
		entity.ResultStatus(job.Status),
		job,
	)
	return &br, nil
}

type DatabaseRepo interface {
	AddJob(job models.Job) error
	GetJobByID(ID int64) (*models.Job, error)
	ListJobByStatus(status string) ([]models.Job, error)
	UpdateJobStatusByID(ID int64, status string) error
	ListAllJobsOfPipeline(pipelineName string) ([]models.Job, error)
	ListNLastJobsOfPipeline(pipelineName string, n int) ([]models.Job, error)
	AddBuildResultLine(jobID int64, line entity.BuildResultLine) error
	GetBuildResult(jobID int64) (*entity.BuildResult, error)
}

func NewDatabaseRepo(db *sqlx.DB) DatabaseRepo {
	return &databaseRepo{
		db: db,
	}
}
