package main

import (
	"errors"
	"fmt"
	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/foldn/bi-go/internal/api"        // Update
	"github.com/foldn/bi-go/internal/config"     // Update
	"github.com/foldn/bi-go/internal/database"   // Update
	"github.com/foldn/bi-go/internal/repository" // Update
	"github.com/foldn/bi-go/internal/service"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// @title           数据处理平台 API
// @version         1.0
// @description     这是一个基于Go语言的数据处理与分析平台API服务.
// @termsOfService  http://example.com/terms/
//
// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com
//
// @license.name   Apache 2.0
// @license.url    http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http https
func main() {

	// Create a logger that outputs to standard output in JSON format
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	//  Load Configuration
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		slog.Error("Failed to load configuration", "error", err, "details", fmt.Sprintf("%v", err))
	}

	//  Initialize Database (GORM)
	db, err := database.Connect(cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err, "details", fmt.Sprintf("%v", err))
	}
	// Auto-migrate schema
	err = database.AutoMigrate(db)
	if err != nil {
		slog.Error("Failed to migrate database", "error", err, "details", fmt.Sprintf("%v", err))
	}
	slog.Info("Database connected and migrated successfully.")

	//  Initialize Repositories
	dsRepo := repository.NewDataSourceRepository(db)
	jobRepo := repository.NewJobRepository(db)
	analysisRepo := repository.NewAnalysisRepository(db)

	//  Initialize Services
	dsService := service.NewDataSourceService(dsRepo)
	jobService := service.NewJobService(jobRepo, dsService, cfg.Job)
	analysisService := service.NewAnalysisService(analysisRepo, jobService)

	jobService.StartWorkers()

	//  Setup Router (and inject services into handlers via router setup)
	router := api.SetupRouter(dsService, jobService, analysisService)
	slog.Info("Starting server on port :" + cfg.Server.Port)

	//  Start Server
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		slog.Error("Failed to start server", "error", err, "details", fmt.Sprintf("%v", err))
	}

	go func() {
		if err := router.Run(":" + cfg.Server.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Unable to start HTTP server", "error", err, "details", fmt.Sprintf("%v", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	sig := <-quit
	slog.Info(fmt.Sprintf("he signal is received %s，The server is ready to shut down...", sig.String()))

	slog.Info(" The server was successfully shut down。")
}
