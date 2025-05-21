# BI-GO 报表管理系统

这是一个基于Golang的后端服务，用于管理和生成报表。

## 项目结构

本项目已按照标准Go项目布局进行了布局，目录结构如下：

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

# Go Data Processing & Analysis API Platform (Gin + GORM)

## 概述

本项目旨在开发一个基于Go语言的、API优先的数据处理与分析平台，使用 **Gin-Gonic** 作为Web框架，**GORM** 作为ORM。专注于提供强大的后端API服务，使开发者能够通过编程方式连接数据源、执行数据转换、进行数据分析，并获取结构化的结果。

## 核心理念

* **API优先:** 所有功能均通过定义良好的RESTful API暴露。
* **模块化:** 后端服务被拆分为逻辑清晰、可独立演进的模块。
* **可扩展性:** 设计考虑未来功能（如更复杂的分析、更广泛的数据源支持）和性能的扩展。
* **高性能:** 利用Go语言的并发特性和高效处理能力，结合Gin的性能优势，应对数据密集型任务。


## 开发阶段规划

以下是项目的后端API开发阶段规划，每个阶段都包含其核心目标、关键功能和主要的API端点。

---

### 阶段一：基础数据连接与元数据服务 API

* **目标：** 建立与数据源的连接能力，并通过API提供数据源的结构信息。
* **关键功能：**
    1.  **数据源配置管理:** 注册、更新、删除和列出数据源配置 (使用GORM操作元数据存储)。
    2.  **数据源模式 (Schema) 发现:** 获取数据源的表/视图/列信息。
* **主要API端点：**
    * `POST /api/v1/datasources`: 创建数据源配置。
        * *请求示例: `{ "name": "my_db", "type": "postgresql", "connection_string": "..." }`*
    * `GET /api/v1/datasources`: 列出所有数据源。
    * `GET /api/v1/datasources/{datasource_id}`: 获取特定数据源配置。
    * `PUT /api/v1/datasources/{datasource_id}`: 更新数据源配置。
    * `DELETE /api/v1/datasources/{datasource_id}`: 删除数据源配置。
    * `GET /api/v1/datasources/{datasource_id}/schema`: 获取数据源的顶层结构 (如表列表)。
    * `GET /api/v1/datasources/{datasource_id}/schema/{entity_name}`: 获取特定实体的详细列结构。

---

### 阶段二：数据提取与基本转换 API，引入异步任务处理

* **目标：** 实现从数据源提取数据，支持基本的列选择、行过滤和简单聚合，并引入异步任务处理机制。
* **关键功能：**
    1.  **数据提取与查询执行:** 提交数据处理任务，支持同步/异步模式。
    2.  **异步任务状态与结果获取:** 查询任务状态，获取处理结果 (任务信息使用GORM存取)。
* **主要API端点：**
    * `POST /api/v1/jobs/process`: 提交数据处理任务。
    * `GET /api/v1/jobs/{job_id}/status`: 获取任务状态。
    * `GET /api/v1/jobs/{job_id}/result`: 获取任务结果 (支持分页)。

---

### 阶段三：高级数据转换与“分析定义”持久化 API

* **目标：** 扩展转换引擎，支持Join、复杂计算字段等高级操作，并允许用户保存和加载“分析定义”。
* **关键功能：**
    1.  **高级转换操作支持:** Join, Calculate Field, Sort, Limit 等。
    2.  **分析定义管理:** 保存、加载、更新、执行已保存的分析流程 (分析定义使用GORM存取)。
* **主要API端点：**
    * `POST /api/v1/jobs/process` (增强 `operations`):
    * `POST /api/v1/analyses`: 创建/保存分析定义。
        * *请求示例: `{ "name": "Q1 Sales Analysis", "definition": { "datasource_id": "...", "operations": [...] } }`*
    * `GET /api/v1/analyses`: 列出所有分析定义。
    * `GET /api/v1/analyses/{analysis_id}`: 获取特定分析定义。
    * `PUT /api/v1/analyses/{analysis_id}`: 更新分析定义。
    * `DELETE /api/v1/analyses/{analysis_id}`: 删除分析定义。
    * `POST /api/v1/analyses/{analysis_id}/execute`: 执行已保存的分析定义。
* **GORM模型示例 (元数据存储):**
    ```go
    // models/analysis.go
    package models
    import "gorm.io/gorm"
    type AnalysisDefinition struct {
        gorm.Model
        Name        string `gorm:"uniqueIndex"`
        Description string
        Definition  string // JSON string or other serialized format of {datasource_id, operations, ...}
        UserID      uint   // Optional: for multi-user systems
    }
    ```

---

### 阶段四：API 成熟度、性能优化与安全性

* **目标：** 提升API的易用性、健壮性、性能和安全性，为生产环境做准备。
* **关键功能/改进：**
    1.  **API版本控制:** (例如 `/api/v2/...`)
    2.  **认证与授权:** 实现API Key或OAuth 2.0认证 (Gin中间件)，以及RBAC。
    3.  **输入校验与错误处理:** (Gin的请求绑定与校验) 完善的校验和标准化的错误响应。
    4.  **速率限制与配额:** (Gin中间件) 防止API滥用。
    5.  **性能监控与日志:** 集成Prometheus、结构化日志 (Gin中间件记录请求日志)、`pprof`。
    6.  **查询优化:** 下推操作到数据库、优化内存算法。GORM的 Preload, Joins, Smart Select Fields 等特性。
    7.  **可配置的数据输出格式:** 支持Parquet, Arrow IPC等。
    8.  **API文档:** 使用OpenAPI (Swagger) 规范 (例如 `swaggo/gin-swagger`)。

## 技术栈

* **语言:** Go
* **Web框架/HTTP路由:** **Gin-Gonic (`github.com/gin-gonic/gin`)**
* **ORM:** **GORM (`gorm.io/gorm`)**
* **数据库驱动 (由GORM管理):** `gorm.io/driver/postgres`, `gorm.io/driver/mysql`, `gorm.io/driver/sqlite`
* **元数据存储数据库:** PostgreSQL, MySQL, SQLite (选择一个与GORM配合使用)
* **配置文件管理:** `github.com/spf13/viper`
* **JSON处理:** `encoding/json` (标准库), Gin内置的JSON处理
* **任务队列 (可选，用于Job Management):** Go channels + Goroutines (内置), Asynq, RabbitMQ, Kafka
* **表达式引擎 (可选，用于Calculate Field):** `github.com/antonmedv/expr` 或自定义
* **日志:** `slog` (Go 1.21+), `logrus`, `zap`; Gin有自己的日志中间件
* **测试:** `testing` (标准库), `testify/assert`, `net/http/httptest` (用于测试Gin处理程序)
* **API文档 (Gin集成):** `github.com/swaggo/gin-swagger`, `github.com/swaggo/swag`

## (占位符) 如何开始

```bash
# 克隆仓库
git clone [https://github.com/foldn/bi-go.git]
cd [bi-go]

# 安装依赖 (Go Modules会自动处理)
go mod tidy

# 配置 (例如，创建 .env 文件或配置 config.yaml, 特别是数据库连接)

# 运行数据库迁移 (如果使用GORM的AutoMigrate)

# 运行服务