package main

import (
	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/foldn/bi-go/internal/api"        // Update
	"github.com/foldn/bi-go/internal/config"     // Update
	"github.com/foldn/bi-go/internal/database"   // Update
	"github.com/foldn/bi-go/internal/repository" // Update
	"github.com/foldn/bi-go/internal/service"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig("./configs") // Or a different path
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Initialize Database (GORM)
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// Auto-migrate schema
	err = database.AutoMigrate(db) // Pass the *gorm.DB instance
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database connected and migrated successfully.")

	// 3. Initialize Repositories
	dsRepo := repository.NewDataSourceRepository(db)
	jobRepo := repository.NewJobRepository(db)

	// 4. Initialize Services
	dsService := service.NewDataSourceService(dsRepo)
	jobService := service.NewJobService(jobRepo, dsService, cfg.JobConfig)

	jobService.StartWorkers()

	// 5. Setup Router (and inject services into handlers via router setup)
	router := api.SetupRouter(dsService, jobService)
	log.Printf("Starting server on port %s", cfg.Server.Port)

	// 6. Start Server
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	go func() {
		if err := router.Run(":" + cfg.Server.Port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("FATAL: Unable to start HTTP server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	sig := <-quit
	log.Printf("INFO: The signal is received %s，The server is ready to shut down...", sig.String())

	log.Println("INFO: The server was successfully shut down。")
}
