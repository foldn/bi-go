package handlers

import (
	"fmt"
	"github.com/foldn/bi-go/internal/models"
	"github.com/foldn/bi-go/internal/services"
	"github.com/foldn/bi-go/internal/storage"
	"github.com/gin-gonic/gin"
	"net/http"
	// "path/filepath"
	"time"
)

// 数据源处理程序

// ListDataSources 获取所有数据源
func ListDataSources(c *gin.Context) {
	dataSources := storage.ListDataSources()
	c.JSON(http.StatusOK, dataSources)
}

// GetDataSource 获取特定数据源
func GetDataSource(c *gin.Context) {
	id := c.Param("id")
	dataSource, err := storage.GetDataSource(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dataSource)
}

// CreateDataSource 创建数据源
func CreateDataSource(c *gin.Context) {
	var request struct {
		Name   string `json:"name" binding:"required"`
		Type   string `json:"type" binding:"required"`
		Config string `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dataSource := models.NewDataSource(request.Name, request.Type, request.Config)
	if err := storage.SaveDataSource(dataSource); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dataSource)
}

// UpdateDataSource 更新数据源
func UpdateDataSource(c *gin.Context) {
	id := c.Param("id")
	dataSource, err := storage.GetDataSource(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var request struct {
		Name   string `json:"name"`
		Type   string `json:"type"`
		Config string `json:"config"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Name != "" {
		dataSource.Name = request.Name
	}
	if request.Type != "" {
		dataSource.Type = request.Type
	}
	if request.Config != "" {
		dataSource.Config = request.Config
	}

	dataSource.UpdatedAt = time.Now()

	if err := storage.SaveDataSource(dataSource); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dataSource)
}

// DeleteDataSource 删除数据源
func DeleteDataSource(c *gin.Context) {
	id := c.Param("id")
	if err := storage.DeleteDataSource(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// 报表处理程序

// ListReports 获取所有报表
func ListReports(c *gin.Context) {
	reports := storage.ListReports()
	c.JSON(http.StatusOK, reports)
}

// GetReport 获取特定报表
func GetReport(c *gin.Context) {
	id := c.Param("id")
	report, err := storage.GetReport(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// CreateReport 创建报表
func CreateReport(c *gin.Context) {
	var request struct {
		Name         string   `json:"name" binding:"required"`
		Description  string   `json:"description"`
		DataSourceID string   `json:"data_source_id" binding:"required"`
		Query        string   `json:"query" binding:"required"`
		Columns      []string `json:"columns" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证数据源是否存在
	_, err := storage.GetDataSource(request.DataSourceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的数据源ID"})
		return
	}

	report := models.NewReport(
		request.Name,
		request.Description,
		request.DataSourceID,
		request.Query,
		request.Columns,
	)

	if err := storage.SaveReport(report); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, report)
}

// UpdateReport 更新报表
func UpdateReport(c *gin.Context) {
	id := c.Param("id")
	report, err := storage.GetReport(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var request struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		DataSourceID string   `json:"data_source_id"`
		Query        string   `json:"query"`
		Columns      []string `json:"columns"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Name != "" {
		report.Name = request.Name
	}
	if request.Description != "" {
		report.Description = request.Description
	}
	if request.DataSourceID != "" {
		// 验证数据源是否存在
		_, err := storage.GetDataSource(request.DataSourceID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的数据源ID"})
			return
		}
		report.DataSourceID = request.DataSourceID
	}
	if request.Query != "" {
		report.Query = request.Query
	}
	if request.Columns != nil {
		report.Columns = request.Columns
	}

	report.UpdatedAt = time.Now()

	if err := storage.SaveReport(report); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// DeleteReport 删除报表
func DeleteReport(c *gin.Context) {
	id := c.Param("id")
	if err := storage.DeleteReport(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// 报表生成和下载处理程序

// GenerateReport 触发报表生成
func GenerateReport(c *gin.Context) {
	id := c.Param("id")
	report, err := storage.GetReport(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var request struct {
		Format string `json:"format" binding:"required,oneof=csv json"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建报表任务
	job := models.NewReportJob(report.ID, request.Format)
	if err := storage.SaveReportJob(job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 异步生成报表
	go services.GenerateReport(job)

	c.JSON(http.StatusAccepted, gin.H{
		"job_id": job.ID,
		"status": job.Status,
	})
}

// GetReportStatus 获取报表生成状态
func GetReportStatus(c *gin.Context) {
	reportID := c.Param("id")
	jobID := c.Query("job_id")

	if jobID == "" {
		// 如果没有指定任务ID，返回该报表的所有任务
		jobs := storage.ListReportJobs(reportID)
		c.JSON(http.StatusOK, jobs)
		return
	}

	// 获取特定任务
	job, err := storage.GetReportJob(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 确保任务属于指定的报表
	if job.ReportID != reportID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务不属于指定的报表"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// DownloadReport 下载生成的报表
func DownloadReport(c *gin.Context) {
	jobID := c.Query("job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少job_id参数"})
		return
	}

	// 获取任务信息
	job, err := storage.GetReportJob(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 检查任务状态
	if job.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "报表尚未生成完成"})
		return
	}

	// 检查文件是否存在
	if job.FilePath == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "报表文件路径不存在"})
		return
	}

	// 设置Content-Type和Content-Disposition
	fileName := fmt.Sprintf("report_%s.%s", job.ID, job.Format)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))

	if job.Format == "csv" {
		c.Header("Content-Type", "text/csv")
	} else if job.Format == "json" {
		c.Header("Content-Type", "application/json")
	}

	// 提供文件下载
	c.File(job.FilePath)
}
