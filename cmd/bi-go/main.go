package main

import (
	"github.com/foldn/bi-go/internal/api"
	"github.com/foldn/bi-go/internal/config"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	// 初始化配置
	config.Init()

	// 创建Gin引擎
	r := gin.Default()

	// 注册API路由
	api.RegisterRoutes(r)

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
