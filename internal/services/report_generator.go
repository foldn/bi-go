package services

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/foldn/bi-go/internal/config"
	"github.com/foldn/bi-go/internal/models"
	"github.com/foldn/bi-go/internal/storage"
	"os"
	"path/filepath"
	"time"
)

// 模拟数据结果
type DataRow map[string]interface{}

// GenerateReport 异步生成报表
func GenerateReport(job *models.ReportJob) {
	// 更新任务状态为运行中
	job.Status = "running"
	job.UpdatedAt = time.Now()
	storage.SaveReportJob(job)

	// 获取报表定义
	report, err := storage.GetReport(job.ReportID)
	if err != nil {
		handleJobError(job, fmt.Sprintf("获取报表定义失败: %v", err))
		return
	}

	// 获取数据源
	dataSource, err := storage.GetDataSource(report.DataSourceID)
	if err != nil {
		handleJobError(job, fmt.Sprintf("获取数据源失败: %v", err))
		return
	}

	// 执行查询获取数据
	data, err := executeQuery(dataSource, report)
	if err != nil {
		handleJobError(job, fmt.Sprintf("执行查询失败: %v", err))
		return
	}

	// 生成报表文件
	filePath, err := generateReportFile(job, report, data)
	if err != nil {
		handleJobError(job, fmt.Sprintf("生成报表文件失败: %v", err))
		return
	}

	// 更新任务状态为完成
	job.Status = "completed"
	job.FilePath = filePath
	job.UpdatedAt = time.Now()
	storage.SaveReportJob(job)
}

// handleJobError 处理任务错误
func handleJobError(job *models.ReportJob, errMsg string) {
	job.Status = "failed"
	job.Error = errMsg
	job.UpdatedAt = time.Now()
	storage.SaveReportJob(job)
}

// executeQuery 执行查询获取数据
// 注意：这里是模拟实现，实际应用中需要根据数据源类型连接实际数据库
func executeQuery(dataSource *models.DataSource, report *models.Report) ([]DataRow, error) {
	// 这里仅作为演示，返回模拟数据
	// 实际应用中，应该根据dataSource.Type和dataSource.Config连接到实际数据源
	// 然后执行report.Query查询，并返回结果

	// 模拟数据
	result := []DataRow{
		{
			"id":    1,
			"name":  "示例1",
			"value": 100,
			"date":  "2023-01-01",
		},
		{
			"id":    2,
			"name":  "示例2",
			"value": 200,
			"date":  "2023-01-02",
		},
		{
			"id":    3,
			"name":  "示例3",
			"value": 300,
			"date":  "2023-01-03",
		},
	}

	return result, nil
}

// generateReportFile 生成报表文件
func generateReportFile(job *models.ReportJob, report *models.Report, data []DataRow) (string, error) {
	// 使用配置的输出目录
	outputDir := config.OutputDir
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", err
	}

	// 生成文件路径
	fileName := fmt.Sprintf("%s_%s.%s", report.ID, job.ID, job.Format)
	filePath := filepath.Join(outputDir, fileName)

	// 根据格式生成文件
	switch job.Format {
	case "csv":
		return generateCSV(filePath, report.Columns, data)
	case "json":
		return generateJSON(filePath, data)
	default:
		return "", errors.New("不支持的报表格式")
	}
}

// generateCSV 生成CSV格式报表
func generateCSV(filePath string, columns []string, data []DataRow) (string, error) {
	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 创建CSV写入器
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	if err := writer.Write(columns); err != nil {
		return "", err
	}

	// 写入数据行
	for _, row := range data {
		values := make([]string, len(columns))
		for i, col := range columns {
			// 获取列值并转换为字符串
			if val, ok := row[col]; ok {
				values[i] = fmt.Sprintf("%v", val)
			} else {
				values[i] = ""
			}
		}

		if err := writer.Write(values); err != nil {
			return "", err
		}
	}

	return filePath, nil
}

// generateJSON 生成JSON格式报表
func generateJSON(filePath string, data []DataRow) (string, error) {
	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 将数据编码为JSON并写入文件
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return "", err
	}

	return filePath, nil
}
