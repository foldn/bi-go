package repository

import (
	"github.com/foldn/bi-go/internal/models"
	"gorm.io/gorm"
)

type AnalysisRepository interface {
	CreateAnalysisDefinition(job *models.AnalysisDefinition) error
	GetAnalysisDefinitionById(id string) (*models.AnalysisDefinition, error)
	UpdateAnalysisDefinition(job *models.AnalysisDefinition) error
	GetAnalysisDefinitions(offset, limit int) (*[]models.AnalysisDefinition, int64, error)
	GetAnalysisDefinitionByUUID(uuid string) (*models.AnalysisDefinition, error)
}

type analysisRepository struct {
	db *gorm.DB
}

func NewAnalysisRepository(db *gorm.DB) AnalysisRepository {
	return &analysisRepository{db: db}
}

func (r *analysisRepository) CreateAnalysisDefinition(analysisDefinition *models.AnalysisDefinition) error {
	return r.db.Create(analysisDefinition).Error
}

func (r *analysisRepository) GetAnalysisDefinitionById(id string) (*models.AnalysisDefinition, error) {
	var analysisDefinition models.AnalysisDefinition
	if err := r.db.First(&analysisDefinition, id).Error; err != nil {
		return nil, err
	}
	return &analysisDefinition, nil
}

func (r *analysisRepository) UpdateAnalysisDefinition(job *models.AnalysisDefinition) error {
	return r.db.Save(&job).Error
}

func (r *analysisRepository) GetAnalysisDefinitions(offset, limit int) (*[]models.AnalysisDefinition, int64, error) {
	var analysisDefinitions []models.AnalysisDefinition
	var total int64
	if err := r.db.Model(&models.AnalysisDefinition{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Model(&models.AnalysisDefinition{}).Offset(offset).Limit(limit).Find(&analysisDefinitions).Error; err != nil {
		return nil, total, err
	}
	return &analysisDefinitions, total, nil
}

func (r *analysisRepository) GetAnalysisDefinitionByUUID(uuid string) (*models.AnalysisDefinition, error) {
	var analysisDefinition models.AnalysisDefinition
	if err := r.db.Where("uuid = ?", uuid).First(&analysisDefinition).Error; err != nil {
		return nil, err
	}
	return &analysisDefinition, nil
}
