package storage

import (
	"errors"
	"github.com/foldn/bi-go/internal/models"
	"sync"
)

// 内存存储实现
var (
	dataSources = make(map[string]*models.DataSource)
	reports     = make(map[string]*models.Report)
	reportJobs  = make(map[string]*models.ReportJob)
	dsLock      = &sync.RWMutex{}
	reportLock  = &sync.RWMutex{}
	jobLock     = &sync.RWMutex{}
)

// 数据源相关操作

// SaveDataSource 保存数据源
func SaveDataSource(ds *models.DataSource) error {
	if ds == nil || ds.ID == "" {
		return errors.New("无效的数据源")
	}

	dsLock.Lock()
	defer dsLock.Unlock()

	dataSources[ds.ID] = ds
	return nil
}

// GetDataSource 获取数据源
func GetDataSource(id string) (*models.DataSource, error) {
	dsLock.RLock()
	defer dsLock.RUnlock()

	ds, exists := dataSources[id]
	if !exists {
		return nil, errors.New("数据源不存在")
	}

	return ds, nil
}

// ListDataSources 列出所有数据源
func ListDataSources() []*models.DataSource {
	dsLock.RLock()
	defer dsLock.RUnlock()

	result := make([]*models.DataSource, 0, len(dataSources))
	for _, ds := range dataSources {
		result = append(result, ds)
	}

	return result
}

// DeleteDataSource 删除数据源
func DeleteDataSource(id string) error {
	dsLock.Lock()
	defer dsLock.Unlock()

	_, exists := dataSources[id]
	if !exists {
		return errors.New("数据源不存在")
	}

	delete(dataSources, id)
	return nil
}

// 报表相关操作

// SaveReport 保存报表
func SaveReport(report *models.Report) error {
	if report == nil || report.ID == "" {
		return errors.New("无效的报表")
	}

	reportLock.Lock()
	defer reportLock.Unlock()

	reports[report.ID] = report
	return nil
}

// GetReport 获取报表
func GetReport(id string) (*models.Report, error) {
	reportLock.RLock()
	defer reportLock.RUnlock()

	report, exists := reports[id]
	if !exists {
		return nil, errors.New("报表不存在")
	}

	return report, nil
}

// ListReports 列出所有报表
func ListReports() []*models.Report {
	reportLock.RLock()
	defer reportLock.RUnlock()

	result := make([]*models.Report, 0, len(reports))
	for _, report := range reports {
		result = append(result, report)
	}

	return result
}

// DeleteReport 删除报表
func DeleteReport(id string) error {
	reportLock.Lock()
	defer reportLock.Unlock()

	_, exists := reports[id]
	if !exists {
		return errors.New("报表不存在")
	}

	delete(reports, id)
	return nil
}

// 报表任务相关操作

// SaveReportJob 保存报表任务
func SaveReportJob(job *models.ReportJob) error {
	if job == nil || job.ID == "" {
		return errors.New("无效的报表任务")
	}

	jobLock.Lock()
	defer jobLock.Unlock()

	reportJobs[job.ID] = job
	return nil
}

// GetReportJob 获取报表任务
func GetReportJob(id string) (*models.ReportJob, error) {
	jobLock.RLock()
	defer jobLock.RUnlock()

	job, exists := reportJobs[id]
	if !exists {
		return nil, errors.New("报表任务不存在")
	}

	return job, nil
}

// ListReportJobs 列出特定报表的所有任务
func ListReportJobs(reportID string) []*models.ReportJob {
	jobLock.RLock()
	defer jobLock.RUnlock()

	result := make([]*models.ReportJob, 0)
	for _, job := range reportJobs {
		if job.ReportID == reportID {
			result = append(result, job)
		}
	}

	return result
}
