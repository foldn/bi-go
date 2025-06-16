package v1

import (
	"errors"
	"github.com/foldn/bi-go/internal/models"
	"github.com/foldn/bi-go/internal/service"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JobHandler struct {
	jobService service.JobService
}

func NewJobHandler(js service.JobService) *JobHandler {
	return &JobHandler{jobService: js}
}

// SubmitJob godoc
// @Summary Submit a new data processing job
// @Description Creates and queues a new data processing job. Can run in sync or async mode.
// @Tags jobs
// @Accept  json
// @Produce  json
// @Param   job_input  body   models.ProcessJobInput  true  "Job Processing Input"
// @Success 200 {object} models.Job "Sync execution: Job details, result might be in response if small JSON"
// @Success 202 {object} models.Job "Async execution: Job details, job accepted for processing"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /jobs/process [post]
func (h *JobHandler) SubmitJob(c *gin.Context) {
	var input models.ProcessJobInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Set defaults if client doesn't specify
	if input.ExecutionMode == "" {
		input.ExecutionMode = models.ExecutionModeAsync
	}
	if input.OutputFormat == "" {
		input.OutputFormat = models.OutputFormatJSON
	}

	job, syncResult, err := h.jobService.SubmitJob(input)
	if err != nil {
		// Service layer should ideally return specific error types or messages
		// to allow better status code mapping here.
		if strings.Contains(err.Error(), "invalid datasource_id") || strings.Contains(err.Error(), "entity") {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		} else if strings.Contains(err.Error(), "failed to marshal operations") {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	if input.ExecutionMode == models.ExecutionModeSync {
		if job.Status == models.JobStatusCompleted {
			// For sync JSON that's small, it's in syncResult.
			// If it was CSV, syncResult is nil, and client uses job.ResultPath from the 'job' object.
			if syncResult != nil && job.OutputFormat == models.OutputFormatJSON {
				c.JSON(http.StatusOK, gin.H{"job": job, "result": syncResult})
			} else { // For CSV or if JSON result is file-based even in sync
				c.JSON(http.StatusOK, gin.H{"job": job}) // Client uses job.ResultPath
			}
		} else { // Sync job failed
			c.JSON(http.StatusInternalServerError, gin.H{"job": job, "error": job.ResultMessage})
		}
		return
	}

	// Async mode
	c.JSON(http.StatusAccepted, job)
}

// GetJobStatus godoc
// @Summary Get the status of a job
// @Description Retrieves the current status and details of a submitted job by its UUID.
// @Tags jobs
// @Produce  json
// @Param   job_uuid   path   string  true  "Job UUID"
// @Success 200 {object} models.JobStatusOutput
// @Failure 400 {object} ErrorResponse "Invalid Job UUID"
// @Failure 404 {object} ErrorResponse "Job not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /jobs/{job_uuid}/status [get]
func (h *JobHandler) GetJobStatus(c *gin.Context) {
	jobUUID := c.Param("job_uuid")
	if jobUUID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Job UUID is required"})
		return
	}

	status, err := h.jobService.GetJobStatus(jobUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Job not found"})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, status)
}

// GetJobResult godoc
// @Summary Get the result of a completed job
// @Description Retrieves the result of a completed job. If the result is a file, it will be served for download.
// @Tags jobs
// @Produce  json
// @Param   job_uuid   path   string  true  "Job UUID"
// @Success 200 {object} models.PaginatedJobResultOutput "If JSON and small, data might be in response"
// @Success 200 "File download if output is CSV or large JSON file"
// @Failure 400 {object} ErrorResponse "Invalid Job UUID"
// @Failure 404 {object} ErrorResponse "Job or result not found"
// @Failure 409 {object} ErrorResponse "Job not completed or failed"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /jobs/{job_uuid}/result [get]
func (h *JobHandler) GetJobResult(c *gin.Context) {
	jobUUID := c.Param("job_uuid")
	if jobUUID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Job UUID is required"})
		return
	}

	resultOutput, err := h.jobService.GetJobResult(jobUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Job not found"})
		} else if strings.Contains(err.Error(), "not completed") {
			c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	if resultOutput.IsFile && resultOutput.FilePath != "" {
		// Check if file exists
		if _, err := os.Stat(resultOutput.FilePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Result file not found."})
			return
		}

		// For CSV or if JSON is served as a file download
		if resultOutput.OutputFormat == models.OutputFormatCSV {
			c.Header("Content-Description", "File Transfer")
			c.Header("Content-Transfer-Encoding", "binary")
			c.Header("Content-Disposition", "attachment; filename="+filepath.Base(resultOutput.FilePath))
			c.Header("Content-Type", "text/csv")
			c.File(resultOutput.FilePath)
			return
		}
		if resultOutput.OutputFormat == models.OutputFormatJSON {
			// Option 1: Serve JSON file directly for download
			// c.Header("Content-Description", "File Transfer")
			// c.Header("Content-Transfer-Encoding", "binary")
			// c.Header("Content-Disposition", "attachment; filename="+filepath.Base(resultOutput.FilePath))
			// c.Header("Content-Type", "application/json")
			// c.File(resultOutput.FilePath)

			// Option 2: Try to read small JSON file and return as JSON response (as done in PaginatedJobResultOutput.Data by service)
			// For now, the service doesn't populate Data if IsFile is true.
			// So, we just serve the file.
			c.Header("Content-Type", "application/json")
			c.File(resultOutput.FilePath)
			return
		}
	}

	// If result is not a file (e.g., small sync JSON result directly in Data, though current service logic prioritizes files)
	if resultOutput.Data != nil {
		c.JSON(http.StatusOK, resultOutput)
		return
	}

	// Fallback if neither file nor direct data, which indicates an issue or incomplete result.
	c.JSON(http.StatusNotFound, ErrorResponse{Error: "Job result not available in the expected format."})
}
