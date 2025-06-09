package models

import "gorm.io/gorm"

// AnalysisDefinition stores a reusable data processing workflow.
type AnalysisDefinition struct {
	gorm.Model
	UUID        string `gorm:"type:varchar(36);uniqueIndex;not null"`
	Name        string `gorm:"type:varchar(255);not null"`
	Description string `gorm:"type:text"`
	// Definition stores the core processing input (DataSourceID, EntityName, Operations) as a JSON string.
	Definition string `gorm:"type:text;not null"`
	// UserID can be added later for multi-tenancy.
	// UserID      uint
}
