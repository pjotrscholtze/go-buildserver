package repo

import (
	"fmt"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/jmoiron/sqlx"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/entity"
	"github.com/pjotrscholtze/go-buildserver/models"
)

func interfaceToString(val interface{}) string {
	if fmt.Sprintf("%T", val) == "string" {
		return val.(string)
	}

	return string(val.([]uint8))
}
func interfaceToDatetime(val interface{}) strfmt.DateTime {
	var queueTime strfmt.DateTime
	if fmt.Sprintf("%T", val) == "[]uint8" {
		qtString := string(val.([]uint8))
		qt, _ := time.Parse("2006-01-02 15:04:05", qtString)
		queueTime = strfmt.DateTime(qt)
	} else {
		queueTime = strfmt.DateTime(val.(time.Time))
	}

	return queueTime
}

type internalJob struct {
	ID           interface{}
	Buildreason  interface{}
	Origin       interface{}
	Queuetime    interface{}
	Pipelinename interface{}
	Status       interface{}
	Starttime    interface{}
}

func (ij *internalJob) AsModelJob() *models.Job {
	return &models.Job{
		BuildReason: interfaceToString(ij.Buildreason),
		ID:          ij.ID.(int64),
		QueueTime:   interfaceToDatetime(ij.Queuetime),
		Origin:      interfaceToString(ij.Origin),
		RepoName:    interfaceToString(ij.Pipelinename),
		Status:      interfaceToString(ij.Status),
	}
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
	qt := time.Time(job.QueueTime).Format("2006-01-02 15:04:05")

	_, err = st.Exec(map[string]interface{}{
		"BuildReason": job.BuildReason,
		"Origin":      job.Origin,
		"QueueTime":   qt,
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
			out = append(out, *row.AsModelJob())
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
			out = append(out, *row.AsModelJob())
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
			out = append(out, *row.AsModelJob())
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

	return job.AsModelJob(), nil
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
		qt := interfaceToDatetime(mres["time"])
		lines = append(lines, entity.NewBuildResultLinePipeString(
			interfaceToString(mres["line"]),
			interfaceToString(mres["pipe"]),
			time.Time(qt),
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
