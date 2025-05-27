package models

import (
	"gorm.io/gorm"
	"time"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "PENDING"
	JobStatusRunning   JobStatus = "RUNNING"
	JobStatusCompleted JobStatus = "COMPLETED"
	JobStatusFailed    JobStatus = "FAILED"
	JobStatusCancelled JobStatus = "CANCELLED"
)

type ExecutionMode string

const (
	ExecutionModeSync  ExecutionMode = "sync"
	ExecutionModeAsync ExecutionMode = "async"
)

type OutputFormat string

const (
	OutputFormatJSON OutputFormat = "json"
	OutputFormatCSV  OutputFormat = "csv"
)

type Job struct {
	gorm.Model
	UUID          string        `gorm:"type:varchar(36);uniqueIndex;not null"`
	DataSourceID  uint          `gorm:"not null"`
	DataSource    DataSource    // GORM 关联
	EntityName    string        `gorm:"type:varchar(255);not null"`
	Operations    string        `gorm:"type:text"` // JSON marshalled operations
	Status        JobStatus     `gorm:"type:varchar(50);default:'PENDING'"`
	OutputFormat  OutputFormat  `gorm:"type:varchar(10);default:'json'"`
	ExecutionMode ExecutionMode `gorm:"type:varchar(10);default:'async'"`
	ResultPath    string        `gorm:"type:text"`
	ResultMessage string        `gorm:"type:text"`
	StartedAt     *time.Time
	CompletedAt   *time.Time
}
