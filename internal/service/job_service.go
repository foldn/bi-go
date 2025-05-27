package service

import (
	"github.com/foldn/bi-go/internal/models"
	"github.com/foldn/bi-go/internal/repository"
)

type JobService interface {
	SubmitJob(input models.ProcessJobInput) (*models.Job, interface{}, error)
	GetJobStatus(jobUUID string) (*models.Job, error)
	GetJobResult(jobUUID string, page, pageSize int) (interface{}, int64, error)
}

type jobService struct {
	repo *repository.JobRepository
}

func NewJobService(repo *repository.JobRepository) JobService {
	return &jobService{repo: repo}
}

func (r *jobService) SubmitJob(input models.ProcessJobInput) (*models.Job, interface{}, error) {

	return nil, nil, nil
}

func (r *jobService) GetJobStatus(jobUUID string) (*models.Job, error) {
	return nil, nil
}

func (r *jobService) GetJobResult(jobUUID string, page, pageSize int) (interface{}, int64, error) {
	return nil, 0, nil
}
