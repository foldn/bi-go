package service

import (
	"github.com/foldn/bi-go/internal/models"
	"github.com/foldn/bi-go/internal/repository"
)

type AnalysisService interface {
	CreateDataSource(input models.CreateAnalysisInput) (*models.AnalysisDefinition, error)
	GetDataSources(page, pageSize int) ([]models.AnalysisDefinition, int64, error)
	GetDataSourceByID(id uint) (*models.AnalysisDefinition, error)
	UpdateDataSource(id uint, input models.UpdateAnalysisInput) (*models.AnalysisDefinition, error)
	DeleteDataSource(id uint) error
	ExecuteAnalysis(uuid uint) models.Job
}

type analysisService struct {
	analysisRepo repository.AnalysisRepository
	jobService   JobService
}

func NewAnalysisService(analysisRepo repository.AnalysisRepository, jobService JobService) AnalysisService {
	return &analysisService{
		analysisRepo: analysisRepo,
		jobService:   jobService,
	}
}

func (r analysisService) CreateDataSource(input models.CreateAnalysisInput) (*models.AnalysisDefinition, error) {
	//TODO implement me
	panic("implement me")
}

func (r analysisService) GetDataSources(page, pageSize int) ([]models.AnalysisDefinition, int64, error) {
	//TODO implement me
	panic("implement me")
}

func (r analysisService) GetDataSourceByID(id uint) (*models.AnalysisDefinition, error) {
	//TODO implement me
	panic("implement me")
}

func (r analysisService) UpdateDataSource(id uint, input models.UpdateAnalysisInput) (*models.AnalysisDefinition, error) {
	//TODO implement me
	panic("implement me")
}

func (r analysisService) DeleteDataSource(id uint) error {
	//TODO implement me
	panic("implement me")
}

func (r analysisService) ExecuteAnalysis(uuid uint) models.Job {
	//TODO implement me
	panic("implement me")
}
