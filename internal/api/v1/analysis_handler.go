package v1

import (
	"github.com/foldn/bi-go/internal/models"
	"github.com/foldn/bi-go/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type AnalysisHandler struct {
	analysisService service.AnalysisService
}

func NewAnalysisHandler(as service.AnalysisService) *AnalysisHandler {
	return &AnalysisHandler{analysisService: as}
}

// CreateAnalysis godoc
// @Summary Create a new Analysis Definition
// @Description Add a new Analysis Definition configuration to the system
// @Tags analysis
// @Accept  json
// @Produce  json
// @Param   Analysis  body   models.CreateAnalysisInput  true  "Data Source Configuration"
// @Success 201 {object} models.AnalysisDefinition
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /analysis [post]
func (h AnalysisHandler) CreateAnalysis(c *gin.Context) {
	var input models.CreateAnalysisInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	as, err := h.analysisService.CreateAnalysis(input)
	if err != nil {
		if err.Error() == "analysis with this name already exists" {
			c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
			return
		}
		handleError(c, err, http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusCreated, as)
}

// GetAnalysis godoc
// @Summary Get all analysis definition
// @Description Retrieve a paginated list of analysis definition
// @Tags analysis
// @Produce  json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Number of items per page" default(10)
// @Success 200 {object} map[string]interface{} "data, total, page, pageSize"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /analysis [get]
func (h AnalysisHandler) GetAnalysis(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	as, total, err := h.analysisService.GetAnalysis(page, pageSize)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     as,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetAnalysisByUUID godoc
// @Summary Get a analysis definition by UUID
// @Description Retrieve a specific analysis definition configuration by its UUID
// @Tags analysis
// @Produce  json
// @Param   id   path   int  true  "analysis definition UUID"
// @Success 200 {object} models.AnalysisDefinition
// @Failure 404 {object} ErrorResponse "Data source not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /analysis/{analysis_uuid} [get]
func (h AnalysisHandler) GetAnalysisByUUID(c *gin.Context) {
	asUUID := c.Param("analysis_uuid")

	as, err := h.analysisService.GetAnalysisByUUID(asUUID)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, as)
}

// UpdateAnalysis godoc
// @Summary Update an existing analysis definition
// @Description Update an existing analysis definition  configuration by its UUID
// @Tags analysis
// @Accept  json
// @Produce  json
// @Param   id   path   int  true  "analysis definition UUID"
// @Param   datasource  body   models.UpdateDataSourceInput  true  "analysis definition Configuration Update"
// @Success 200 {object} models.AnalysisDefinition
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 404 {object} ErrorResponse "Data source not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /analysis/{analysis_uuid} [put]
func (h AnalysisHandler) UpdateAnalysis(c *gin.Context) {
	asUUID := c.Param("analysis_uuid")

	var input models.UpdateAnalysisInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	as, err := h.analysisService.UpdateAnalysis(asUUID, input)
	if err != nil {
		if err.Error() == "analysis with this name already exists" {
			c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
			return
		}
		handleError(c, err, http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, as)
}

// DeleteAnalysis godoc
// @Summary Delete a analysis definition
// @Description Delete a analysis definition configuration by its UUID
// @Tags analysis
// @Produce  json
// @Param   id   path   int  true  "analysis definition UUID"
// @Success 204 "Successfully deleted"
// @Failure 404 {object} ErrorResponse "Data source not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /datasources/{id} [delete]
func (h AnalysisHandler) DeleteAnalysis(c *gin.Context) {
	asUUID := c.Param("analysis_uuid")

	err := h.analysisService.DeleteAnalysis(asUUID)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusNoContent)
}

// ExecuteAnalysis godoc
// @Summary submit an analysis definition
// @Description Creates a new data processing job by uuid.
// @Tags analysis
// @Produce  json
// @Param   id   path   int  true  "analysis definition UUID"
// @Success 200 {object} models.Job
// @Failure 400 {object} ErrorResponse "Invalid ID format"
// @Failure 404 {object} ErrorResponse "Data source not found"
// @Failure 500 {object} ErrorResponse "Error fetching schema"
// @Router /analysis/{analysis_uuid}/execute [post]
func (h AnalysisHandler) ExecuteAnalysis(c *gin.Context) {
	asUUID := c.Param("analysis_uuid")

	var input models.ExecuteAnalysisInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	job, err := h.analysisService.ExecuteAnalysis(asUUID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, job)

}
