# BI-GO 报表管理系统

这是一个基于Golang的后端服务，用于管理和生成报表。

## 功能特点

- 数据源配置管理
- 报表定义管理
- 异步报表生成
- 报表状态查询
- 报表下载（支持CSV和JSON格式）

## 项目结构

本项目已按照标准Go项目布局进行了重构，新的目录结构如下：

```
.
├── cmd/                    # 主要应用程序入口点
│   └── bi-go/              # 主应用程序
│       └── main.go         # 主应用程序入口点
├── internal/               # 私有应用程序和库代码
│   ├── api/                # API服务器实现
│   │   ├── handlers/       # HTTP处理程序
│   │   └── routes.go       # 路由定义
│   ├── config/             # 配置管理
│   ├── models/             # 数据模型
│   ├── services/           # 业务逻辑服务
│   └── storage/            # 数据存储实现
├── pkg/                    # 可以被外部应用使用的库代码
│   └── utils/              # 通用工具函数
├── examples/               # 示例应用
├── configs/                # 配置文件模板或默认配置
├── docs/                   # 设计和用户文档
└── scripts/                # 构建、安装、分析等脚本
```
## 构建和运行

```bash
# 构建应用
go build -o bin/bi-go ./cmd/bi-go

# 运行应用
./bin/bi-go
```

## 快速开始

1. 克隆仓库
2. 安装依赖: `go mod tidy`
3. 运行服务: `go run main.go`
4. 访问API: `http://localhost:8080`

## API文档

### 数据源管理
- `GET /api/datasources` - 获取所有数据源
- `GET /api/datasources/:id` - 获取特定数据源
- `POST /api/datasources` - 创建数据源
- `PUT /api/datasources/:id` - 更新数据源
- `DELETE /api/datasources/:id` - 删除数据源

### 报表管理
- `GET /api/reports` - 获取所有报表
- `GET /api/reports/:id` - 获取特定报表
- `POST /api/reports` - 创建报表
- `PUT /api/reports/:id` - 更新报表
- `DELETE /api/reports/:id` - 删除报表

### 报表生成
- `POST /api/reports/:id/generate` - 触发报表生成
- `GET /api/reports/:id/status` - 查询报表生成状态
- `GET /api/reports/:id/download` - 下载生成的报表