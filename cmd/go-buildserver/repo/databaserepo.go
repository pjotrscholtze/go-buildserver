package repo

import (
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/jmoiron/sqlx"
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
	st, err := dr.db.PrepareNamed("INSERT INTO Job " +
		"(BuildReason, Origin, QueueTime, Pipelinename, Status) VALUES " +
		"(:BuildReason, :Origin, :QueueTime, :RepoName, :Status)")

	if err != nil {
		return err
	}

	_, err = st.Exec(map[string]interface{}{
		"BuildReason": job.BuildReason,
		"Origin":      job.Origin,
		"QueueTime":   job.QueueTime,
		"RepoName":    job.RepoName,
		"Status":      job.Status,
	})

	return err
}
func (dr *databaseRepo) ListAllJobsOfPipeline(pipelineName string) ([]models.Job, error) {
	out := []models.Job{}
	st, err := dr.db.PrepareNamed("SELECT * FROM Job WHERE Pipelinename = :Pipelinename")

	if err != nil {
		return out, err
	}

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
	st, err := dr.db.PrepareNamed("SELECT * FROM Job WHERE status = :status")

	if err != nil {
		return out, err
	}

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
	st, err := dr.db.PrepareNamed("SELECT * FROM Job WHERE id = :id")

	if err != nil {
		return nil, err
	}

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
	st, err := dr.db.PrepareNamed("UPDATE Job SET status=:status WHERE id=:id")
	if err != nil {
		return err
	}
	args := map[string]interface{}{}
	args["id"] = ID
	args["status"] = status

	_, err = st.Exec(args)
	return err
}

type DatabaseRepo interface {
	AddJob(job models.Job) error
	GetJobByID(ID int64) (*models.Job, error)
	ListJobByStatus(status string) ([]models.Job, error)
	UpdateJobStatusByID(ID int64, status string) error
	ListAllJobsOfPipeline(pipelineName string) ([]models.Job, error)
}

func NewDatabaseRepo(db *sqlx.DB) DatabaseRepo {
	return &databaseRepo{
		db: db,
	}
}
