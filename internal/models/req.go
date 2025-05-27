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
	DataSourceID  uint            `json:"datasource_id" binding:"required"`
	EntityName    string          `json:"entity_name" binding:"required"`
	Operations    []OperationStep `json:"operations" binding:"required,dive"` // dive ensures nested validation
	OutputFormat  OutputFormat    `json:"output_format" binding:"omitempty,oneof=json csv"`
	ExecutionMode ExecutionMode   `json:"execution_mode" binding:"omitempty,oneof=sync async"`
}
