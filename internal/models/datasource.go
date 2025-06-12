package models

import (
	"fmt"
	"gorm.io/gorm"
	"strings"
)

type DataSourceType string

type IsDeleteType int

const (
	PostgreSQL DataSourceType = "postgresql"
	MySQL      DataSourceType = "mysql"
	CSV        DataSourceType = "csv"
	ClickHouse DataSourceType = "clickhouse"
	SQLite     DataSourceType = "sqlite"
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
}

func (ds *DataSource) QuoteIdentifier(name string) string {
	if name == "" {
		return "" // Handle empty identifier if necessary, or error
	}
	switch ds.Type {
	case PostgreSQL, SQLite:
		return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	case MySQL, ClickHouse:
		return "`" + strings.ReplaceAll(name, "`", "``") + "`"
	case CSV:
		return name
	default:
		fmt.Printf("Warning: QuoteIdentifier used for unhandled DataSourceType: %s. Identifier '%s' returned as is.\n", ds.Type, name)
		return name
	}
}

func (ds *DataSource) IsDatabase() bool {
	return ds.Type == PostgreSQL || ds.Type == MySQL || ds.Type == SQLite || ds.Type == ClickHouse
}
