package models

type CreateDataSourceInput struct {
	Name        string         `json:"name" binding:"required"`
	Type        DataSourceType `json:"type" binding:"required,oneof=postgresql mysql csv"`
	Host        string         `json:"host"`
	Port        string         `json:"port"`
	Username    string         `json:"username"`
	Password    string         `json:"password"`
	DBName      string         `json:"dbName"`
	FilePath    string         `json:"filePath"`
	OtherParams string         `json:"otherParams"`
	Description string         `json:"description"`
}

type UpdateDataSourceInput struct {
	Name        *string         `json:"name"` // Use pointers for optional updates
	Type        *DataSourceType `json:"type" binding:"omitempty,oneof=postgresql mysql csv clickhouse sqlite"`
	Host        *string         `json:"host"`
	Port        *string         `json:"port"`
	Username    *string         `json:"username"`
	Password    *string         `json:"password"`
	DBName      *string         `json:"dbName"`
	FilePath    *string         `json:"filePath"`
	OtherParams *string         `json:"otherParams"`
	Description *string         `json:"description"`
}

type ProcessJobInput struct {
	DataSourceID  uint            `json:"datasourceId" binding:"required"`
	EntityName    string          `json:"entityName" binding:"required"`
	Operations    []OperationStep `json:"operations" binding:"required,dive"` // dive ensures nested validation
	OutputFormat  OutputFormat    `json:"outputFormat" binding:"omitempty,oneof=json csv"`
	ExecutionMode ExecutionMode   `json:"executionMode" binding:"omitempty,oneof=sync async"`
}

type DefinitionInput struct {
	DataSourceID uint            `json:"datasource_id" binding:"required"`
	EntityName   string          `json:"entity_name" binding:"required"`
	Operations   []OperationStep `json:"operations" binding:"required,dive"`
}

// CreateAnalysisInput is the DTO for creating a new analysis definition.
type CreateAnalysisInput struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description"`
	Definition  DefinitionInput `json:"definition" binding:"required"`
}

// UpdateAnalysisInput is the DTO for updating an analysis definition.
type UpdateAnalysisInput struct {
	Name        *string          `json:"name"` // Pointers for optional updates
	Description *string          `json:"description"`
	Definition  *DefinitionInput `json:"definition"`
}

// ExecuteAnalysisInput allows overriding some parameters when executing a saved analysis.
type ExecuteAnalysisInput struct {
	// Override the output format of the saved analysis for this execution.
	OutputFormat OutputFormat `json:"output_format,omitempty" binding:"omitempty,oneof=json csv"`
	// Always run as async when triggered via this endpoint.
	// ExecutionMode models.ExecutionMode `json:"execution_mode" binding:"omitempty,oneof=sync async"`
}
