package api

import (
	"github.com/foldn/bi-go/internal/api/v1"
	"github.com/foldn/bi-go/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(dsService service.DataSourceService /*, other services can be passed here */) *gin.Engine {
	// gin.SetMode(gin.ReleaseMode) // Uncomment for production
	router := gin.Default() // Includes logger and recovery middleware

	// TODO: Add CORS middleware if needed
	// router.Use(cors.Default())

	// TODO: Add any other global middleware (e.g., authentication, custom logging)

	// Instantiate handlers
	dsHandler := v1.NewDataSourceHandler(dsService)

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

	}

	return router
}
