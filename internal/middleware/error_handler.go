package middleware

import (
	"errors"
	v1 "github.com/foldn/bi-go/internal/api/v1"
	"github.com/foldn/bi-go/internal/models/apierror"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // 先执行后续的 handlers

		// c.Next() 执行完毕后，检查上下文中是否有错误
		if len(c.Errors) == 0 {
			return
		}

		// 我们只处理第一个错误
		err := c.Errors[0].Err

		// 检查是否是我们自定义的 AppError
		var appErr *apierror.AppError
		if errors.As(err, &appErr) {
			c.AbortWithStatusJSON(appErr.StatusCode, v1.APIError{
				Code:    appErr.Code,
				Message: appErr.Message,
			})
			return
		}

		// 检查是否是 GORM 的 RecordNotFound 错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, v1.APIError{
				Code:    "RESOURCE_NOT_FOUND",
				Message: "请求的资源未找到",
			})
			return
		}

		// 其他未预料到的错误，返回 500
		// 在生产环境中，不应暴露原始错误信息
		c.AbortWithStatusJSON(http.StatusInternalServerError, v1.APIError{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "服务器内部发生未知错误",
		})
	}
}
