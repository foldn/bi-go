package models
import "gorm.io/gorm"

type DataSourceType string
const (
    PostgreSQL DataSourceType = "postgresql"
    MySQL      DataSourceType = "mysql"
    CSV        DataSourceType = "csv"
)

type DataSource struct {
    gorm.Model
    Name        string `gorm:"type:varchar(255);uniqueIndex;not null"`
    Type        DataSourceType `gorm:"type:varchar(50);not null"`
    Host        string `gorm:"type:varchar(255)"`
    Port        string `gorm:"type:varchar(10)"`
    Username    string `gorm:"type:varchar(255)"`
    Password    string `gorm:"type:varchar(255)"` 
    DBName      string `gorm:"type:varchar(255)"` 
    FilePath    string `gorm:"type:text"`         
    OtherParams string `gorm:"type:text"`         
    Description string `gorm:"type:text"`
}