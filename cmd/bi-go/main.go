package main

import (
	"github.com/foldn/bi-go/internal/api"        // Update
	"github.com/foldn/bi-go/internal/config"     // Update
	"github.com/foldn/bi-go/internal/database"   // Update
	"github.com/foldn/bi-go/internal/repository" // Update
	"github.com/foldn/bi-go/internal/service"
	"log"
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

	// 4. Initialize Services
	dsService := service.NewDataSourceService(dsRepo /*, pass other dependencies if any, like schemaService */)

	// 5. Setup Router (and inject services into handlers via router setup)
	router := api.SetupRouter(dsService)
	log.Printf("Starting server on port %s", cfg.Server.Port)

	// 6. Start Server
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
