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

func (h AnalysisHandler) GetAnalyses(c *gin.Context) {
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

func (h AnalysisHandler) GetAnalysisByUUID(c *gin.Context) {
	asUUID := c.Param("analysis_uuid")

	as, err := h.analysisService.GetAnalysisByUUID(asUUID)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, as)
}

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

func (h AnalysisHandler) DeleteAnalysis(c *gin.Context) {
	asUUID := c.Param("analysis_uuid")

	err := h.analysisService.DeleteAnalysis(asUUID)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusNoContent)
}

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
