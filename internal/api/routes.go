package api

import (
	"github.com/foldn/bi-go/internal/api/handlers"
	"github.com/gin-gonic/gin"
	"net/http"
)

// RegisterRoutes 注册所有API路由
func RegisterRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API路由组
	api := r.Group("/api")
	{
		// 数据源管理
		datasources := api.Group("/datasources")
		{
			datasources.GET("", handlers.ListDataSources)
			datasources.GET("/:id", handlers.GetDataSource)
			datasources.POST("", handlers.CreateDataSource)
			datasources.PUT("/:id", handlers.UpdateDataSource)
			datasources.DELETE("/:id", handlers.DeleteDataSource)
		}

		// 报表管理
		reports := api.Group("/reports")
		{
			reports.GET("", handlers.ListReports)
			reports.GET("/:id", handlers.GetReport)
			reports.POST("", handlers.CreateReport)
			reports.PUT("/:id", handlers.UpdateReport)
			reports.DELETE("/:id", handlers.DeleteReport)

			// 报表生成和下载
			reports.POST("/:id/generate", handlers.GenerateReport)
			reports.GET("/:id/status", handlers.GetReportStatus)
			reports.GET("/:id/download", handlers.DownloadReport)
		}
	}
}
