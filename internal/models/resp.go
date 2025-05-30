package models

import "time"

type JobStatusOutput struct {
	UUID          string       `json:"uuid"`
	Status        JobStatus    `json:"status"`
	ResultMessage string       `json:"resultMessage,omitempty"`
	ResultPath    string       `json:"resultPath,omitempty"`
	CreatedAt     time.Time    `json:"createdAt"`
	StartedAt     *time.Time   `json:"startedAt,omitempty"`
	CompletedAt   *time.Time   `json:"completedAt,omitempty"`
	OutputFormat  OutputFormat `json:"outputFormat"`
}

// PaginatedJobResultOutput DTO for returning paginated job results
// For Phase 2, results are primarily files, so direct data pagination is less emphasized for async.
// This DTO might be more relevant if results were stored in a queryable way.
type PaginatedJobResultOutput struct {
	Data         interface{}  `json:"data"` // Could be []map[string]interface{} or a message
	Total        int64        `json:"total"`
	Page         int          `json:"page"`
	PageSize     int          `json:"pageSize"`
	FilePath     string       `json:"filePath,omitempty"` // Path to download if result is a file
	IsFile       bool         `json:"isFile"`
	OutputFormat OutputFormat `json:"outputFormat"`
}
