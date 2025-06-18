package middleware

import (
	"github.com/foldn/bi-go/internal/metrice"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.FullPath()

		// 更新指标
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(latency.Seconds())
		metrics.HTTPRequestsTotal.WithLabelValues(method, path, strconv.Itoa(statusCode)).Inc()
	}
}
