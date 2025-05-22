package repository

import (
	"github.com/foldn/bi-go/internal/models"
	"gorm.io/gorm"
)

type DataSourceRepository interface {
	Create(ds *models.DataSource) error
	GetAll(offset, limit int) ([]models.DataSource, int64, error)
	GetByID(id uint) (*models.DataSource, error)
	Update(ds *models.DataSource) error
	Delete(id uint) error
	GetByName(name string) (*models.DataSource, error)
}

type dataSourceRepository struct {
	db *gorm.DB
}

func NewDataSourceRepository(db *gorm.DB) DataSourceRepository {
	return &dataSourceRepository{db: db}
}

func (r *dataSourceRepository) Create(ds *models.DataSource) error {
	return r.db.Create(ds).Error
}

func (r *dataSourceRepository) GetAll(offset, limit int) ([]models.DataSource, int64, error) {
	var dataSources []models.DataSource
	var total int64
	if err := r.db.Model(&models.DataSource{}).Where("is_delete = ?", models.NOT_DELETE).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.Where("is_delete = ?", models.NOT_DELETE).Offset(offset).Limit(limit).Find(&dataSources).Error; err != nil {
		return nil, total, err
	}
	return dataSources, total, nil
}

func (r *dataSourceRepository) GetByID(id uint) (*models.DataSource, error) {
	var ds models.DataSource
	if err := r.db.Where("is_delete = ?", models.NOT_DELETE).First(&ds, id).Error; err != nil {
		return nil, err
	}
	return &ds, nil
}

func (r *dataSourceRepository) Update(ds *models.DataSource) error {
	return r.db.Save(ds).Error
}

func (r *dataSourceRepository) Delete(id uint) error {
	return r.db.Model(&models.DataSource{}).Where("id = ?", id).Update("is_delete", models.IS_DELETE).Error
}

func (r *dataSourceRepository) GetByName(name string) (*models.DataSource, error) {
	var ds models.DataSource
	if err := r.db.Where("name = ? and is_delete = ?", name, models.NOT_DELETE).First(&ds).Error; err != nil {
		return nil, err
	}
	return &ds, nil
}
