package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

// FormatTime 格式化时间为字符串
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// ParseTime 解析时间字符串
func ParseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// PrettyJSON 将对象格式化为美观的JSON字符串
func PrettyJSON(v interface{}) (string, error) {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ValidateFormat 验证报表格式是否支持
func ValidateFormat(format string) bool {
	validFormats := map[string]bool{
		"csv":  true,
		"json": true,
	}
	
	_, ok := validFormats[format]
	return ok
}

// GenerateFileName 生成报表文件名
func GenerateFileName(reportID, jobID, format string) string {
	return fmt.Sprintf("%s_%s_%s.%s", reportID, jobID, time.Now().Format("20060102150405"), format)
}