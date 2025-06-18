package middleware

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"time"
)

func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		logger := slog.With(
			"status_code", statusCode,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"latency", latency.String(),
		)

		if len(c.Errors) > 0 {
			logger.Error("请求处理时发生错误", "error", c.Errors.String())
		} else {
			if statusCode >= 500 {
				logger.Error("请求成功但服务器端状态码异常")
			} else if statusCode >= 400 {
				logger.Warn("请求成功但客户端状态码异常")
			} else {
				logger.Info("请求处理成功")
			}
		}
	}
}
