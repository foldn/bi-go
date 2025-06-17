package v1

type APIError struct {
	Code    string      `json:"code,omitempty"`    // 自定义的错误码，方便客户端程序处理
	Message string      `json:"message"`           // 对错误的描述
	Details interface{} `json:"details,omitempty"` // 可选，提供更详细的错误信息，如字段校验失败详情
}
