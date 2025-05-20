package models

import (
	"github.com/google/uuid"
	"time"
)

// DataSource 数据源配置
type DataSource struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // 例如: mysql, postgres, csv等
	Config      string    `json:"config"` // JSON格式的连接配置
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Report 报表定义
type Report struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	DataSourceID string    `json:"data_source_id"`
	Query        string    `json:"query"` // SQL查询或其他查询语句
	Columns      []string  `json:"columns"` // 输出列定义
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ReportJob 报表生成任务
type ReportJob struct {
	ID        string    `json:"id"`
	ReportID  string    `json:"report_id"`
	Status    string    `json:"status"` // pending, running, completed, failed
	Format    string    `json:"format"` // csv, json等
	FilePath  string    `json:"file_path,omitempty"` // 生成的报表文件路径
	Error     string    `json:"error,omitempty"` // 错误信息
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewDataSource 创建新的数据源
func NewDataSource(name, dsType, config string) *DataSource {
	now := time.Now()
	return &DataSource{
		ID:        uuid.New().String(),
		Name:      name,
		Type:      dsType,
		Config:    config,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewReport 创建新的报表定义
func NewReport(name, description, dataSourceID, query string, columns []string) *Report {
	now := time.Now()
	return &Report{
		ID:           uuid.New().String(),
		Name:         name,
		Description:  description,
		DataSourceID: dataSourceID,
		Query:        query,
		Columns:      columns,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// NewReportJob 创建新的报表生成任务
func NewReportJob(reportID, format string) *ReportJob {
	now := time.Now()
	return &ReportJob{
		ID:        uuid.New().String(),
		ReportID:  reportID,
		Status:    "pending",
		Format:    format,
		CreatedAt: now,
		UpdatedAt: now,
	}
}