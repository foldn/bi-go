package config

import (
	"log"
	"os"
	"path/filepath"
)

// 应用配置
var (
	// OutputDir 报表输出目录
	OutputDir string
)

// Init 初始化配置
func Init() {
	// 初始化输出目录
	OutputDir = getEnv("OUTPUT_DIR", "./output")

	// 确保输出目录存在
	if err := os.MkdirAll(OutputDir, 0755); err != nil {
		log.Fatalf("无法创建输出目录: %v", err)
	}

	// 转换为绝对路径
	absPath, err := filepath.Abs(OutputDir)
	if err == nil {
		OutputDir = absPath
	}

	log.Printf("配置初始化完成，报表输出目录: %s", OutputDir)
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}