package service

import (
	"errors"
	"fmt"
	"github.com/foldn/bi-go/internal/models"
	"github.com/foldn/bi-go/internal/repository"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type DataSourceService interface {
	CreateDataSource(input CreateDataSourceInput) (*models.DataSource, error)
	GetDataSources(page, pageSize int) ([]models.DataSource, int64, error)
	GetDataSourceByID(id uint) (*models.DataSource, error)
	UpdateDataSource(id uint, input UpdateDataSourceInput) (*models.DataSource, error)
	DeleteDataSource(id uint) error

	// Schema discovery methods - to be detailed in schema_service.go or here
	GetDataSourceSchema(dataSourceID uint) (interface{}, error)
	GetDataSourceEntitySchema(dataSourceID uint, entityName string) (interface{}, error)
}

type dataSourceService struct {
	repo repository.DataSourceRepository
}

func NewDataSourceService(repo repository.DataSourceRepository) DataSourceService {
	return &dataSourceService{repo: repo}
}

type CreateDataSourceInput struct {
	Name        string                `json:"name" binding:"required"`
	Type        models.DataSourceType `json:"type" binding:"required,oneof=postgresql mysql csv"`
	Host        string                `json:"host"`
	Port        string                `json:"port"`
	Username    string                `json:"username"`
	Password    string                `json:"password"`
	DBName      string                `json:"dbName"`
	FilePath    string                `json:"filePath"`
	OtherParams string                `json:"otherParams"`
	Description string                `json:"description"`
}

type UpdateDataSourceInput struct {
	Name        *string                `json:"name"` // Use pointers for optional updates
	Type        *models.DataSourceType `json:"type" binding:"omitempty,oneof=postgresql mysql csv clickhouse sqlite"`
	Host        *string                `json:"host"`
	Port        *string                `json:"port"`
	Username    *string                `json:"username"`
	Password    *string                `json:"password"`
	DBName      *string                `json:"dbName"`
	FilePath    *string                `json:"filePath"`
	OtherParams *string                `json:"otherParams"`
	Description *string                `json:"description"`
}

func (s *dataSourceService) CreateDataSource(input CreateDataSourceInput) (*models.DataSource, error) {
	// Check for duplicate name
	existing, err := s.repo.GetByName(input.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("error checking existing datasource: %w", err)
	}
	if existing != nil {
		return nil, errors.New("datasource with this name already exists")
	}

	ds := &models.DataSource{
		Name:        input.Name,
		Type:        input.Type,
		Host:        input.Host,
		Port:        input.Port,
		Username:    input.Username,
		Password:    input.Password, // Remember security!
		DBName:      input.DBName,
		FilePath:    input.FilePath,
		OtherParams: input.OtherParams,
		Description: input.Description,
	}
	if err := s.repo.Create(ds); err != nil {
		return nil, err
	}
	return ds, nil
}

func (s *dataSourceService) GetDataSources(page, pageSize int) ([]models.DataSource, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	return s.repo.GetAll(offset, pageSize)
}

func (s *dataSourceService) GetDataSourceByID(id uint) (*models.DataSource, error) {
	return s.repo.GetByID(id)
}

func (s *dataSourceService) UpdateDataSource(id uint, input UpdateDataSourceInput) (*models.DataSource, error) {
	ds, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err // handles gorm.ErrRecordNotFound appropriately
	}

	// Update fields if provided in input
	if input.Name != nil {
		// Check for duplicate name if changed
		if *input.Name != ds.Name {
			existing, err := s.repo.GetByName(*input.Name)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("error checking existing datasource: %w", err)
			}
			if existing != nil && existing.ID != id { // if another DS has this new name
				return nil, errors.New("datasource with this name already exists")
			}
		}
		ds.Name = *input.Name
	}
	if input.Type != nil {
		ds.Type = *input.Type
	}
	if input.Host != nil {
		ds.Host = *input.Host
	}

	if input.Port != nil {
		ds.Port = *input.Port
	}
	// ... update other fields similarly

	if input.Username != nil {
		ds.Username = *input.Username
	}
	if input.Password != nil {
		ds.Password = *input.Password
	} // Security!
	if input.Description != nil {
		ds.Description = *input.Description
	}

	if input.DBName != nil {
		ds.DBName = *input.DBName
	}

	if input.FilePath != nil {
		ds.FilePath = *input.FilePath
	}

	if input.OtherParams != nil {
		ds.OtherParams = *input.OtherParams
	}

	if err := s.repo.Update(ds); err != nil {
		return nil, err
	}
	return ds, nil
}

func (s *dataSourceService) DeleteDataSource(id uint) error {
	// Optionally check if datasource exists before deleting
	_, err := s.repo.GetByID(id)
	if err != nil {
		return err // handles gorm.ErrRecordNotFound appropriately
	}
	return s.repo.Delete(id)
}

// Placeholder for schema service methods - actual implementation is complex
func (s *dataSourceService) GetDataSourceSchema(dataSourceID uint) (interface{}, error) {
	// 1. Get DataSource config by dataSourceID using s.repo
	// 2. Based on ds.Type, connect to the actual data source (NOT the metadata DB)
	// 3. Fetch schema (tables for DBs, columns for CSVs)
	// 4. Return formatted schema
	ds, err := s.repo.GetByID(dataSourceID)
	if err != nil {
		return nil, err
	}

	dataSourceType := ds.Type
	switch dataSourceType {
	case models.PostgreSQL:
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
			ds.Host, ds.Username, ds.Password, ds.DBName, ds.Port)

		newLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: time.Second,
				LogLevel:      logger.Info, // Or logger.Silent for less noise
				Colorful:      true,
			},
		)
		postgresDb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: newLogger,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to postgresDb: %w", err)
		}
		return postgresDb.Exec("SHOW DATABASES"), err
	case models.ClickHouse:
		dsn := fmt.Sprintf("tcp://%s:%s?username=%s&password=%s&database=%s",
			ds.Host, ds.Port, ds.Username, ds.Password, ds.DBName)

		newLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: time.Second,
				LogLevel:      logger.Info, // Or logger.Silent for less noise
				Colorful:      true,
			},
		)
		clickhouseDb, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{
			Logger: newLogger,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to clickhouse:%w ,%w", dsn, err)
		}
		return clickhouseDb.Exec("show databases;"), err
	case models.Sqlite:
		dsn := fmt.Sprintf("tcp://%s:%s?username=%s&password=%s&database=%s",
			ds.Host, ds.Port, ds.Username, ds.Password, ds.DBName)

		newLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: time.Second,
				LogLevel:      logger.Info, // Or logger.Silent for less noise
				Colorful:      true,
			},
		)
		sqlLiteDb, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{
			Logger: newLogger,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to clickhouse:%w ,%w", dsn, err)
		}
		return sqlLiteDb.Exec("show databases;"), err
	default:
	}
	return nil, errors.New("GetDataSourceSchema not implemented yet")
}

func (s *dataSourceService) GetDataSourceEntitySchema(dataSourceID uint, entityName string) (interface{}, error) {
	// 1. Get DataSource config
	// 2. Connect to actual data source
	// 3. Fetch specific entity (table/CSV) schema (columns with types)
	// 4. Return formatted schema
	return nil, errors.New("GetDataSourceEntitySchema not implemented yet")
}
