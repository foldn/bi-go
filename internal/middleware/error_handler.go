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
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors[0].Err

		var appErr *apierror.AppError
		if errors.As(err, &appErr) {
			c.AbortWithStatusJSON(appErr.StatusCode, v1.APIError{
				Code:    appErr.Code,
				Message: appErr.Message,
			})
			return
		}

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
