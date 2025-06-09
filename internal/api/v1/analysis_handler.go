package v1

import (
	"github.com/foldn/bi-go/internal/service"
	"github.com/gin-gonic/gin"
)

type AnalysisHandler struct {
	analysisService service.AnalysisService
}

func NewAnalysisHandler(as service.AnalysisService) *AnalysisHandler {
	return &AnalysisHandler{analysisService: as}
}

func (h AnalysisHandler) CreateAnalysis(context *gin.Context) {

}

func (h AnalysisHandler) GetAnalyses(context *gin.Context) {

}

func (h AnalysisHandler) GetAnalysisByUUID(context *gin.Context) {

}

func (h AnalysisHandler) UpdateAnalysis(context *gin.Context) {

}

func (h AnalysisHandler) DeleteAnalysis(context *gin.Context) {

}

func (h AnalysisHandler) ExecuteAnalysis(context *gin.Context) {

}
