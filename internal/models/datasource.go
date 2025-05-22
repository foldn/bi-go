package models

import "gorm.io/gorm"

type DataSourceType string

type IsDeleteType int

const (
	PostgreSQL DataSourceType = "postgresql"
	MySQL      DataSourceType = "mysql"
	CSV        DataSourceType = "csv"
	ClickHouse DataSourceType = "clickhouse"
	Sqlite     DataSourceType = "sqlite"
)

const (
	IS_DELETE  int = 1
	NOT_DELETE int = 0
)

type DataSource struct {
	gorm.Model
	Name        string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	Type        DataSourceType `gorm:"type:varchar(50);not null"`
	Host        string         `gorm:"type:varchar(255)"`
	Port        string         `gorm:"type:varchar(10)"`
	Username    string         `gorm:"type:varchar(255)"`
	Password    string         `gorm:"type:varchar(255)"`
	DBName      string         `gorm:"type:varchar(255)"`
	FilePath    string         `gorm:"type:text"`
	OtherParams string         `gorm:"type:text"`
	Description string         `gorm:"type:text"`
	IsDelete    IsDeleteType   `gorm:"type:tinyint"`
}
