package main

import (
	"fmt"
	"github.com/foldn/bi-go/internal/config"
	"github.com/foldn/bi-go/internal/models"
	"github.com/foldn/bi-go/internal/services"
	"github.com/foldn/bi-go/internal/storage"
	"time"
)

func main() {
	// 初始化配置
	config.Init()

	// 创建示例数据源
	dataSource := createExampleDataSource()
	fmt.Printf("创建数据源: %s\n", dataSource.Name)

	// 创建示例报表
	report := createExampleReport(dataSource.ID)
	fmt.Printf("创建报表: %s\n", report.Name)

	// 创建报表生成任务
	job := models.NewReportJob(report.ID, "csv")
	storage.SaveReportJob(job)
	fmt.Printf("创建报表任务: %s, 格式: %s\n", job.ID, job.Format)

	// 生成报表
	fmt.Println("开始生成报表...")
	services.GenerateReport(job)

	// 等待报表生成完成
	time.Sleep(1 * time.Second)

	// 获取更新后的任务状态
	updatedJob, _ := storage.GetReportJob(job.ID)
	fmt.Printf("报表生成状态: %s\n", updatedJob.Status)

	if updatedJob.Status == "completed" {
		fmt.Printf("报表文件路径: %s\n", updatedJob.FilePath)
	} else if updatedJob.Status == "failed" {
		fmt.Printf("报表生成失败: %s\n", updatedJob.Error)
	}
}

// 创建示例数据源
func createExampleDataSource() *models.DataSource {
	// MySQL数据源配置示例
	configJSON := `{
		"host": "localhost",
		"port": 3306,
		"username": "root",
		"password": "password",
		"database": "example_db"
	}`

	dataSource := models.NewDataSource(
		"示例MySQL数据源",
		"mysql",
		configJSON,
	)

	storage.SaveDataSource(dataSource)
	return dataSource
}

// 创建示例报表
func createExampleReport(dataSourceID string) *models.Report {
	// 定义报表列
	columns := []string{"id", "name", "value", "date"}

	// 创建报表定义
	report := models.NewReport(
		"月度销售报表",
		"展示每月销售数据统计",
		dataSourceID,
		"SELECT id, name, value, date FROM sales WHERE date >= '2023-01-01' AND date <= '2023-01-31'",
		columns,
	)

	storage.SaveReport(report)
	return report
}
