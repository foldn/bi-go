package apierror

import "net/http"

type AppError struct {
	StatusCode int
	Code       string
	Message    string
	Err        error
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(statusCode int, code, message string) *AppError {
	return &AppError{StatusCode: statusCode, Code: code, Message: message}
}

func NewFromError(statusCode int, code string, err error) *AppError {
	return &AppError{StatusCode: statusCode, Code: code, Message: err.Error(), Err: err}
}

// 预定义一些常用错误
var (
	ErrNotFound       = New(http.StatusNotFound, "RESOURCE_NOT_FOUND", "请求的资源未找到")
	ErrInternalServer = New(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "服务器内部错误")
	ErrConflict       = New(http.StatusConflict, "RESOURCE_CONFLICT", "资源已存在或状态冲突")
)

func NewBadRequest(message string) *AppError {
	return New(http.StatusBadRequest, "INVALID_INPUT", message)
}
