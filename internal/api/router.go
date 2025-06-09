package api

import (
	"github.com/foldn/bi-go/internal/api/v1"
	"github.com/foldn/bi-go/internal/service"
	"github.com/gin-gonic/gin"
)

func SetupRouter(dsService service.DataSourceService,
	jobService service.JobService,
	analysisService service.AnalysisService) *gin.Engine {
	// gin.SetMode(gin.ReleaseMode) // Uncomment for production
	router := gin.Default() // Includes logger and recovery middleware

	// TODO: Add CORS middleware if needed
	// router.Use(cors.Default())

	// TODO: Add any other global middleware (e.g., authentication, custom logging)

	// Instantiate handlers
	dsHandler := v1.NewDataSourceHandler(dsService)
	jobHandler := v1.NewJobHandler(jobService) // <--- 实例化 JobHandler
	analysisHandler := v1.NewAnalysisHandler(analysisService)

	// Base API group
	apiV1 := router.Group("/api/v1")
	{
		// Datasource routes
		dsRoutes := apiV1.Group("/datasources")
		{
			dsRoutes.POST("", dsHandler.CreateDataSource)
			dsRoutes.GET("", dsHandler.GetDataSources)
			dsRoutes.GET("/:id", dsHandler.GetDataSourceByID)
			dsRoutes.PUT("/:id", dsHandler.UpdateDataSource)
			dsRoutes.DELETE("/:id", dsHandler.DeleteDataSource)
			dsRoutes.GET("/:id/schema", dsHandler.GetDataSourceSchema)
			dsRoutes.GET("/:id/schema/:entity_name", dsHandler.GetDataSourceEntitySchema)
		}

		// Job routes
		jobRoutes := apiV1.Group("/job")
		{
			jobRoutes.POST("/process", jobHandler.SubmitJob)
			jobRoutes.GET("/:job_uuid/status", jobHandler.GetJobStatus)
			jobRoutes.GET("/:job_uuid/result", jobHandler.GetJobResult)
		}

		analysisRoutes := apiV1.Group("/analyses")
		{
			analysisRoutes.POST("", analysisHandler.CreateAnalysis)
			analysisRoutes.GET("", analysisHandler.GetAnalyses)
			analysisRoutes.GET("/:analysis_uuid", analysisHandler.GetAnalysisByUUID)
			analysisRoutes.PUT("/:analysis_uuid", analysisHandler.UpdateAnalysis)
			analysisRoutes.DELETE("/:analysis_uuid", analysisHandler.DeleteAnalysis)
			analysisRoutes.POST("/:analysis_uuid/execute", analysisHandler.ExecuteAnalysis) // 执行分析
		}

	}

	return router
}
