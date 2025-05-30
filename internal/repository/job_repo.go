package repository

import (
	"github.com/foldn/bi-go/internal/models"
	"gorm.io/gorm"
)

type JobRepository interface {
	CreateJob(job *models.Job) error
	GetJobById(id string) (*models.Job, error)
	UpdateJobStatus(id string, status models.JobStatus) error
	UpdateJob(job *models.Job) error
	GetJobs(offset, limit int) (*[]models.Job, int64, error)
	GetJobByUUID(uuid string) (*models.Job, error)
}

type jobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) JobRepository {
	return &jobRepository{db: db}
}

func (r *jobRepository) CreateJob(job *models.Job) error {
	return r.db.Create(job).Error
}

func (r *jobRepository) GetJobById(id string) (*models.Job, error) {
	var job models.Job
	if err := r.db.First(&job, id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *jobRepository) UpdateJobStatus(id string, status models.JobStatus) error {
	return r.db.Model(&models.Job{}).Where("id = ?", id).Update("Status", status).Error
}

func (r *jobRepository) UpdateJob(job *models.Job) error {
	return r.db.Save(&job).Error
}

func (r *jobRepository) GetJobs(offset, limit int) (*[]models.Job, int64, error) {
	var jobs []models.Job
	var total int64
	if err := r.db.Model(&models.Job{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Model(&models.Job{}).Offset(offset).Limit(limit).Find(&jobs).Error; err != nil {
		return nil, total, err
	}
	return &jobs, total, nil
}

func (r *jobRepository) GetJobByUUID(uuid string) (*models.Job, error) {
	var job models.Job
	if err := r.db.Where("uuid = ?", uuid).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}
