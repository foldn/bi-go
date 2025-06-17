package api

import (
	_ "github.com/foldn/bi-go/docs"
	"github.com/foldn/bi-go/internal/api/v1"
	"github.com/foldn/bi-go/internal/middleware"
	"github.com/foldn/bi-go/internal/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(dsService service.DataSourceService,
	jobService service.JobService,
	analysisService service.AnalysisService) *gin.Engine {
	// gin.SetMode(gin.ReleaseMode) // Uncomment for production
	router := gin.New() // Includes logger and recovery middleware

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.ErrorHandler())

	// Instantiate handlers
	dsHandler := v1.NewDataSourceHandler(dsService)
	jobHandler := v1.NewJobHandler(jobService) // <--- 实例化 JobHandler
	analysisHandler := v1.NewAnalysisHandler(analysisService)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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

		analysisRoutes := apiV1.Group("/analysis")
		{
			analysisRoutes.POST("", analysisHandler.CreateAnalysis)
			analysisRoutes.GET("", analysisHandler.GetAnalysis)
			analysisRoutes.GET("/:analysis_uuid", analysisHandler.GetAnalysisByUUID)
			analysisRoutes.PUT("/:analysis_uuid", analysisHandler.UpdateAnalysis)
			analysisRoutes.DELETE("/:analysis_uuid", analysisHandler.DeleteAnalysis)
			analysisRoutes.POST("/:analysis_uuid/execute", analysisHandler.ExecuteAnalysis) // 执行分析
		}

	}

	return router
}
