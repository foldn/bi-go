package database

import (
	"fmt"
	"github.com/foldn/bi-go/internal/config" // Update with your module path
	"github.com/foldn/bi-go/internal/models" // Update with your module path
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

var DB *gorm.DB

func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
	// user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info, // Or logger.Silent for less noise
			Colorful:      true,
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = db // Assign to global var if needed, or pass around
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(&models.DataSource{})
	if err != nil {
		return fmt.Errorf("failed to auto-migrate database: %w", err)
	}
	log.Println("Database migration completed.")
	return nil
}
