package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/foldn/bi-go/internal/models"
	"github.com/foldn/bi-go/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AnalysisService interface {
	CreateAnalysis(input models.CreateAnalysisInput) (*models.AnalysisDefinition, error)
	GetAnalysis(page, pageSize int) ([]models.AnalysisDefinition, int64, error)
	GetAnalysisByID(id uint) (*models.AnalysisDefinition, error)
	UpdateAnalysis(uuid string, input models.UpdateAnalysisInput) (*models.AnalysisDefinition, error)
	DeleteAnalysis(uuid string) error
	ExecuteAnalysis(uuid string, input models.ExecuteAnalysisInput) (*models.Job, error)
	GetAnalysisByUUID(uuid string) (*models.AnalysisDefinition, error)
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

func (s analysisService) CreateAnalysis(input models.CreateAnalysisInput) (*models.AnalysisDefinition, error) {
	// Optional: Check for duplicate analysis name.
	existing, err := s.analysisRepo.GetByName(input.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("error checking for existing analysis: %w", err)
	}
	if existing != nil {
		return nil, errors.New("an analysis with this name already exists")
	}

	// Serialize the definition part (DataSourceID, EntityName, Operations) into a JSON string.
	definitionJSON, err := json.Marshal(input.Definition)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize analysis definition: %w", err)
	}

	// Create the model instance.
	analysis := &models.AnalysisDefinition{
		UUID:        uuid.New().String(),
		Name:        input.Name,
		Description: input.Description,
		Definition:  string(definitionJSON),
	}

	// Persist to the database.
	if err := s.analysisRepo.CreateAnalysisDefinition(analysis); err != nil {
		return nil, fmt.Errorf("failed to save analysis definition: %w", err)
	}

	return analysis, nil
}

func (s analysisService) GetAnalysis(page, pageSize int) ([]models.AnalysisDefinition, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize
	return s.analysisRepo.GetAnalysisDefinitions(offset, pageSize)
}

func (s analysisService) GetAnalysisByID(id uint) (*models.AnalysisDefinition, error) {
	//TODO implement me
	panic("implement me")
}

func (s analysisService) UpdateAnalysis(uuid string, input models.UpdateAnalysisInput) (*models.AnalysisDefinition, error) {
	// Fetch the existing record.
	analysis, err := s.analysisRepo.GetAnalysisDefinitionByUUID(uuid)
	if err != nil {
		// This will handle gorm.ErrRecordNotFound from the repository.
		return nil, err
	}

	// Update fields if provided in the input (pointers are used for optional updates).
	if input.Name != nil {
		// If name is being changed, check for conflicts.
		if *input.Name != analysis.Name {
			existing, err := s.analysisRepo.GetByName(*input.Name)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("error checking for existing analysis name: %w", err)
			}
			if existing != nil && existing.UUID != uuid {
				return nil, errors.New("another analysis with this name already exists")
			}
		}
		analysis.Name = *input.Name
	}

	if input.Description != nil {
		analysis.Description = *input.Description
	}

	if input.Definition != nil {
		definitionJSON, err := json.Marshal(*input.Definition)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize updated analysis definition: %w", err)
		}
		analysis.Definition = string(definitionJSON)
	}

	// Persist the changes.
	if err := s.analysisRepo.UpdateAnalysisDefinition(analysis); err != nil {
		return nil, fmt.Errorf("failed to update analysis definition: %w", err)
	}

	return analysis, nil
}

func (s analysisService) DeleteAnalysis(uuid string) error {
	_, err := s.analysisRepo.GetAnalysisDefinitionByUUID(uuid)
	if err != nil {
		return err // Handles gorm.ErrRecordNotFound from the repository.
	}

	return s.analysisRepo.DeleteByUUID(uuid)
}

func (s analysisService) ExecuteAnalysis(uuid string, input models.ExecuteAnalysisInput) (*models.Job, error) {
	// 1. Fetch the saved analysis definition.
	analysisDef, err := s.analysisRepo.GetAnalysisDefinitionByUUID(uuid)
	if err != nil {
		return nil, fmt.Errorf("analysis definition with UUID %s not found: %w", uuid, err)
	}

	// 2. Deserialize the stored JSON definition into our DefinitionInput struct.
	var defInput models.DefinitionInput
	if err := json.Unmarshal([]byte(analysisDef.Definition), &defInput); err != nil {
		return nil, fmt.Errorf("failed to parse saved analysis definition for UUID %s: %w", uuid, err)
	}

	// 3. Construct the input for the JobService based on the saved definition.
	jobInput := models.ProcessJobInput{
		DataSourceID: defInput.DataSourceID,
		EntityName:   defInput.EntityName,
		Operations:   defInput.Operations,
		// When executing a saved analysis, we default to async mode as it's typically a background task.
		ExecutionMode: models.ExecutionModeAsync,
		// Default output format can be specified here or taken from the definition if it existed there.
		OutputFormat: models.OutputFormatJSON,
	}

	// 4. Apply any overrides from the execution-time input.
	if input.OutputFormat != "" {
		jobInput.OutputFormat = input.OutputFormat
	}

	// 5. Submit the job using the JobService.
	// We expect the 'syncResult' to be nil for async jobs.
	job, _, err := s.jobService.SubmitJob(jobInput)
	if err != nil {
		return nil, fmt.Errorf("failed to submit job for analysis UUID %s: %w", uuid, err)
	}

	// 6. Return the created job details.
	return job, nil
}

func (s analysisService) GetAnalysisByUUID(uuid string) (*models.AnalysisDefinition, error) {
	return s.analysisRepo.GetAnalysisDefinitionByUUID(uuid)
}
