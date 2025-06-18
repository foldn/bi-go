package service

import (
	"bytes"
	stdcsv "encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/expr-lang/expr"
	"github.com/foldn/bi-go/internal/config"
	"github.com/foldn/bi-go/internal/models"
	"github.com/foldn/bi-go/internal/repository"
	"github.com/google/uuid"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type JobService interface {
	SubmitJob(input models.ProcessJobInput) (*models.Job, interface{}, error)
	GetJobStatus(jobUUID string) (*models.JobStatusOutput, error)
	GetJobResult(jobUUID string) (*models.PaginatedJobResultOutput, error)
	StartWorkers()
}

type jobService struct {
	jobRepo         repository.JobRepository
	dsService       DataSourceService // For getting DB connections and validating schema
	cfg             config.JobConfig
	workerQueue     chan models.Job // In-memory queue
	resultsBaseDir  string
	activeWorkers   int
	maxWorkers      int
	workerWaitGroup sync.WaitGroup
}

func NewJobService(jobRepo repository.JobRepository, dsService DataSourceService, cfg config.JobConfig) JobService {

	resultsDir := cfg.ResultsBasePath
	if resultsDir == "" {
		resultsDir = "./job_results_default" // Default if not configured
	}
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		log.Printf("Warning: could not create job results directory %s: %v", resultsDir, err)
		// Potentially fallback or panic if essential
	}

	maxWorkers := cfg.NumWorkers
	if maxWorkers <= 0 {
		maxWorkers = 3 // Default number of workers
	}

	return &jobService{
		jobRepo:        jobRepo,
		dsService:      dsService,
		cfg:            cfg,
		workerQueue:    make(chan models.Job, 100), // Buffered channel
		resultsBaseDir: resultsDir,
		maxWorkers:     maxWorkers,
	}

}

func (s *jobService) StartWorkers() {
	slog.Info(fmt.Sprintf("Starting %d job workers...", s.maxWorkers))
	for i := 0; i < s.maxWorkers; i++ {
		s.workerWaitGroup.Add(1)
		go func(workerID int) {
			defer s.workerWaitGroup.Done()
			slog.Info(fmt.Sprintf("Worker %d started", workerID))
			for job := range s.workerQueue { // Process jobs from the queue
				slog.Info(fmt.Sprintf("Worker %d: Processing job %s", workerID, job.UUID))
				s.processJobInBackground(job)
			}
			slog.Info(fmt.Sprintf("Worker %d stopped", workerID))
		}(i)
	}
}

// (Call this on graceful shutdown if needed)
// func (s *jobService) StopWorkers() {
// 	close(s.workerQueue)
// 	s.workerWaitGroup.Wait()
// 	log.Println("All job workers stopped.")
// }

func (s *jobService) SubmitJob(input models.ProcessJobInput) (*models.Job, interface{}, error) {
	// 1. Validate DataSource
	ds, err := s.dsService.GetDataSourceByID(input.DataSourceID) // dsService has internal GORM error mapping
	if err != nil {
		return nil, nil, fmt.Errorf("invalid datasource_id: %w", err)
	}

	// 2. (Optional but recommended) Validate EntityName against schema
	// This requires GetDataSourceSchema to be robust from Phase 1
	// _, err = s.dsService.GetDataSourceEntitySchema(input.DataSourceID, input.EntityName)
	// if err != nil {
	// 	return nil, nil, fmt.Errorf("entity '%s' not found or schema error for datasource_id %d: %w", input.EntityName, input.DataSourceID, err)
	// }

	jobUUID := uuid.New().String()
	operationsJSON, err := json.Marshal(input.Operations)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal operations: %w", err)
	}

	// Set defaults if not provided
	if input.OutputFormat == "" {
		input.OutputFormat = models.OutputFormatJSON
	}
	if input.ExecutionMode == "" {
		input.ExecutionMode = models.ExecutionModeAsync
	}

	job := &models.Job{
		UUID:          jobUUID,
		DataSourceID:  input.DataSourceID,
		DataSource:    *ds, // Assign the fetched DataSource object
		EntityName:    input.EntityName,
		Operations:    string(operationsJSON),
		Status:        models.JobStatusPending,
		OutputFormat:  input.OutputFormat,
		ExecutionMode: input.ExecutionMode,
	}

	if err := s.jobRepo.CreateJob(job); err != nil {
		return nil, nil, fmt.Errorf("failed to create job record: %w", err)
	}
	log.Printf("Job %s created for datasource %d, entity %s. Mode: %s", job.UUID, job.DataSourceID, job.EntityName, job.ExecutionMode)

	if job.ExecutionMode == models.ExecutionModeSync {
		log.Printf("Executing job %s in sync mode.", job.UUID)
		// For sync mode, execute directly. Result (data or file path) handled by pipeline.
		syncResultData, resultFilePath, execErr := s.executeDataProcessingPipeline(job)
		now := time.Now()
		job.StartedAt = &now // Should be set at the beginning of execution

		if execErr != nil {
			job.Status = models.JobStatusFailed
			job.ResultMessage = execErr.Error()
			job.CompletedAt = &now
			s.jobRepo.UpdateJob(job) // Persist failure
			return job, nil, fmt.Errorf("sync job execution failed: %w", execErr)
		}

		job.Status = models.JobStatusCompleted
		job.ResultPath = resultFilePath
		job.CompletedAt = &now
		s.jobRepo.UpdateJob(job)        // Persist success
		return job, syncResultData, nil // syncResultData might be the actual data for JSON, or nil if CSV (path used)

	} else { // Async mode
		log.Printf("Dispatching job %s to async worker queue.", job.UUID)
		s.workerQueue <- *job // Send a copy to the queue
		return job, nil, nil  // Return job details immediately
	}
}

func (s *jobService) GetJobStatus(jobUUID string) (*models.JobStatusOutput, error) {
	job, err := s.jobRepo.GetJobByUUID(jobUUID)
	if err != nil {
		return nil, err // Handles gorm.ErrRecordNotFound from repo
	}
	return &models.JobStatusOutput{
		UUID:          job.UUID,
		Status:        job.Status,
		ResultMessage: job.ResultMessage,
		ResultPath:    job.ResultPath,
		CreatedAt:     job.CreatedAt,
		StartedAt:     job.StartedAt,
		CompletedAt:   job.CompletedAt,
		OutputFormat:  job.OutputFormat,
	}, nil
}

func (s *jobService) GetJobResult(jobUUID string) (*models.PaginatedJobResultOutput, error) {
	job, err := s.jobRepo.GetJobByUUID(jobUUID)
	if err != nil {
		return nil, err
	}

	if job.Status != models.JobStatusCompleted {
		return nil, fmt.Errorf("job %s is not completed. Current status: %s", jobUUID, job.Status)
	}

	output := &models.PaginatedJobResultOutput{
		FilePath:     job.ResultPath,
		IsFile:       job.ResultPath != "",
		OutputFormat: job.OutputFormat,
	}

	// For Phase 2, we primarily assume results are files for async jobs.
	// If we wanted to return small JSON results directly:
	if job.OutputFormat == models.OutputFormatJSON && job.ResultPath != "" {
		// Optionally, if the JSON file is small, read and return its content.
		// For large files, client should download via path. This is a policy decision.
		// For now, we always indicate it's a file if ResultPath is set.
		// content, err := os.ReadFile(job.ResultPath)
		// if err == nil {
		//   var jsonData interface{}
		//   if json.Unmarshal(content, &jsonData) == nil {
		//      output.Data = jsonData
		//   }
		// }
	}

	return output, nil
}

func (s *jobService) processJobInBackground(job models.Job) {
	now := time.Now()
	job.StartedAt = &now
	job.Status = models.JobStatusRunning
	if err := s.jobRepo.UpdateJob(&job); err != nil {
		log.Printf("Error updating job %s status to RUNNING: %v", job.UUID, err)
		// Potentially requeue or mark as terminally failed if update fails
		return
	}

	log.Printf("Starting background execution for job %s", job.UUID)
	_, resultFilePath, execErr := s.executeDataProcessingPipeline(&job) // We don't need syncData for async

	completedTime := time.Now()
	job.CompletedAt = &completedTime

	if execErr != nil {
		log.Printf("Job %s failed: %v", job.UUID, execErr)
		job.Status = models.JobStatusFailed
		job.ResultMessage = execErr.Error()
	} else {
		log.Printf("Job %s completed successfully. Result at: %s", job.UUID, resultFilePath)
		job.Status = models.JobStatusCompleted
		job.ResultPath = resultFilePath
		job.ResultMessage = "Successfully processed."
	}

	if err := s.jobRepo.UpdateJob(&job); err != nil {
		log.Printf("Error updating job %s final status: %v", job.UUID, err)
	}
}

// fetchDataForEntity fetches data for a given datasource and entity, returning it as []map[string]interface{}.
// This method centralizes the data loading logic used by both the initial pipeline and the 'join' operation.
func (s *jobService) fetchDataForEntity(dataSourceID uint, entityName string) ([]map[string]interface{}, error) {
	dsConfig, err := s.dsService.GetDataSourceByID(dataSourceID)
	if err != nil {
		return nil, fmt.Errorf("datasource with ID %d not found: %w", dataSourceID, err)
	}

	log.Printf("Fetching data for entity '%s' from datasource %d (type: %s)", entityName, dataSourceID, dsConfig.Type)

	if dsConfig.Type == models.CSV {
		// --- CSV Data Fetching ---
		file, err := os.Open(dsConfig.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open CSV file %s: %w", dsConfig.FilePath, err)
		}
		defer file.Close()
		reader := stdcsv.NewReader(file)
		header, err := reader.Read()
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV header from %s: %w", dsConfig.FilePath, err)
		}

		allRecords, err := reader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("failed to read all CSV records from %s: %w", dsConfig.FilePath, err)
		}

		var dataset []map[string]interface{}
		for _, record := range allRecords {
			if len(record) != len(header) {
				log.Printf("Warning: CSV record field count (%d) does not match header field count (%d). Skipping record: %v", len(record), len(header), record)
				continue
			}
			rowData := make(map[string]interface{})
			for i, colName := range header {
				rowData[colName] = record[i]
			}
			dataset = append(dataset, rowData)
		}
		return dataset, nil

	} else if dsConfig.IsDatabase() { // Assume a helper `IsDatabase()` on the model, or list types explicitly
		// --- Database Data Fetching ---
		db, dsConfigWithDetails, err := s.dsService.GetDBConnection(dataSourceID)
		if err != nil {
			return nil, err
		}
		defer db.Close()

		query := fmt.Sprintf("SELECT * FROM %s", dsConfigWithDetails.QuoteIdentifier(entityName))
		rows, err := db.Query(query)
		if err != nil {
			return nil, fmt.Errorf("failed to query entity %s: %w", entityName, err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for entity %s: %w", entityName, err)
		}

		var dataset []map[string]interface{}
		for rows.Next() {
			rowValues := make([]interface{}, len(columns))
			rowPointers := make([]interface{}, len(columns))
			for i := range rowValues {
				rowPointers[i] = &rowValues[i]
			}

			if err := rows.Scan(rowPointers...); err != nil {
				return nil, fmt.Errorf("failed to scan row for entity %s: %w", entityName, err)
			}

			rowData := make(map[string]interface{})
			for i, colName := range columns {
				val := rowValues[i]
				if b, ok := val.([]byte); ok {
					rowData[colName] = string(b)
				} else {
					rowData[colName] = val
				}
			}
			dataset = append(dataset, rowData)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error during row iteration for entity %s: %w", entityName, err)
		}
		return dataset, nil
	}

	return nil, fmt.Errorf("unsupported data source type for fetching data: %s", dsConfig.Type)
}

// executeDataProcessingPipeline now uses the refactored fetchDataForEntity.
func (s *jobService) executeDataProcessingPipeline(job *models.Job) (interface{}, string, error) {
	log.Printf("Executing pipeline for job %s on %s", job.UUID, job.EntityName)

	var operations []models.OperationStep
	if err := json.Unmarshal([]byte(job.Operations), &operations); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal operations for job %s: %w", job.UUID, err)
	}

	// 1. Initial Data Fetching using the refactored method
	currentData, err := s.fetchDataForEntity(job.DataSourceID, job.EntityName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch initial data for job %s: %w", job.UUID, err)
	}
	log.Printf("Job %s: Fetched initial dataset with %d rows.", job.UUID, len(currentData))

	// 2. In-Memory Transformation Loop
	for i, op := range operations {
		log.Printf("Job %s: Applying operation %d: %s", job.UUID, i+1, op.Type)
		transformedData, err := s.applyOperation(op, currentData)
		if err != nil {
			return nil, "", fmt.Errorf("error applying operation %s for job %s: %w", op.Type, job.UUID, err)
		}
		currentData = transformedData
		log.Printf("Job %s: After operation %s, dataset size: %d", job.UUID, op.Type, len(currentData))
	}

	// 3. Output Generation (this logic can be refactored into a helper if not already)
	return s.generateOutput(job, currentData)
}

// --- 增强 applyOperation ---

// applyOperation dispatches to specific transformation functions, now including advanced ops.
func (s *jobService) applyOperation(op models.OperationStep, data []map[string]interface{}) ([]map[string]interface{}, error) {
	if data == nil && op.Type != "join" { // Join is special as it can create the initial dataset
		return []map[string]interface{}{}, nil
	}
	switch op.Type {
	// --- Existing cases ---
	case "select_columns":
		if op.Columns == nil || len(op.Columns) == 0 {
			return nil, errors.New("select_columns: 'columns' field is required and cannot be empty")
		}
		return s.applySelect(op.Columns, data), nil
	case "filter_rows":
		if op.Condition == nil {
			return nil, errors.New("filter_rows: 'condition' field is required")
		}
		return s.applyFilter(*op.Condition, data)
	case "aggregate":
		if op.Aggregation == nil {
			return nil, errors.New("aggregate: 'aggregation' field is required")
		}
		return s.applyAggregate(*op.Aggregation, data)
	// --- New cases for Phase 3 ---
	case "limit_rows":
		if op.Limit == nil {
			return nil, errors.New("limit_rows: 'limit' field is required")
		}
		return s.applyLimit(*op.Limit, data), nil
	case "sort_rows":
		if op.Sort == nil || len(op.Sort) == 0 {
			return nil, errors.New("sort_rows: 'sort' field is required and cannot be empty")
		}
		return s.applySort(op.Sort, data)
	case "calculate_field":
		if op.Calculation == nil {
			return nil, errors.New("calculate_field: 'calculation' field is required")
		}
		return s.applyCalculateField(*op.Calculation, data)
	case "join":
		if op.Join == nil {
			return nil, errors.New("join: 'join' field is required")
		}
		return s.applyJoin(*op.Join, data) // data is the "left" dataset
	default:
		return nil, fmt.Errorf("unknown operation type: %s", op.Type)
	}
}

// --- 新的转换函数实现 ---

// applyLimit limits the number of rows in the dataset.
func (s *jobService) applyLimit(limit int, data []map[string]interface{}) []map[string]interface{} {
	if limit <= 0 {
		return []map[string]interface{}{} // Return empty if limit is zero or negative
	}
	if limit > len(data) {
		return data // Return all if limit is greater than dataset size
	}
	return data[:limit]
}

// applySort sorts the dataset based on specified columns and orders.
func (s *jobService) applySort(sortParams []models.SortParam, data []map[string]interface{}) ([]map[string]interface{}, error) {
	sort.Slice(data, func(i, j int) bool {
		for _, p := range sortParams {
			valI := data[i][p.Column]
			valJ := data[j][p.Column]

			// Default to ascending order if not specified
			order := strings.ToUpper(p.Order)
			if order == "" {
				order = "ASC"
			}

			// Handle nil values (nil is considered "less than" any other value)
			if valI == nil && valJ != nil {
				return order == "ASC"
			}
			if valI != nil && valJ == nil {
				return order == "DESC"
			}
			if valI == nil && valJ == nil {
				continue
			}

			// Compare based on type
			numI, iIsNum := convertToFloat64(valI)
			numJ, jIsNum := convertToFloat64(valJ)

			// If both are numbers, compare numerically
			if iIsNum && jIsNum {
				if numI != numJ {
					if order == "ASC" {
						return numI < numJ
					}
					return numI > numJ
				}
				continue // If equal, move to next sort criterion
			}

			// Otherwise, compare as strings
			strI := fmt.Sprintf("%v", valI)
			strJ := fmt.Sprintf("%v", valJ)
			if strI != strJ {
				if order == "ASC" {
					return strI < strJ
				}
				return strI > strJ
			}
		}
		return false // All sort criteria are equal
	})
	return data, nil
}

// applyCalculateField adds a new column based on an expression.
func (s *jobService) applyCalculateField(calc models.CalcParams, data []map[string]interface{}) ([]map[string]interface{}, error) {
	// If there's no data, there's nothing to calculate.
	if len(data) == 0 {
		return data, nil
	}

	// --- CORRECTED LOGIC ---
	// Compile the expression once for performance.
	// To do this correctly, we provide the first row of our data as an environment "schema".
	// This tells the compiler what variables to expect and what their types are.
	// This assumes all rows in the dataset have a consistent structure.
	program, err := expr.Compile(calc.Expression, expr.Env(data[0]))
	if err != nil {
		// This is a syntax error in the expression itself, it should fail the entire operation.
		return nil, fmt.Errorf("failed to compile expression '%s': %w", calc.Expression, err)
	}

	for _, row := range data {
		// Run the compiled program with the current row as the environment.
		// The keys in the map (column names) become available as variables in the expression.
		result, err := expr.Run(program, row)
		if err != nil {
			// An error during run could be due to a runtime type mismatch (e.g., 'text' + 1),
			// or a variable that exists in the first row but not this one.
			// Policy: We set the result to nil for this row and log the error, allowing the pipeline to continue.
			log.Printf("Failed to run expression on row: %v. Error: %v. Setting result to nil.", row, err)
			row[calc.NewColumnName] = nil
		} else {
			row[calc.NewColumnName] = result
		}
	}
	return data, nil
}

// applyJoin performs an in-memory hash join between the current dataset (left) and a new one (right).
func (s *jobService) applyJoin(params models.JoinParams, leftData []map[string]interface{}) ([]map[string]interface{}, error) {
	if len(params.On) != 1 {
		return nil, errors.New("join currently supports only a single 'on' condition") // Multi-condition join is more complex
	}
	joinCondition := params.On[0]

	// 1. Fetch the right-side data
	rightData, err := s.fetchDataForEntity(params.RightDataSourceID, params.RightEntityName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch right side of join ('%s'): %w", params.RightEntityName, err)
	}

	// 2. Build a hash map from the right-side data for efficient lookups.
	// The map key is the value of the join column. The map value is a slice of rows to handle non-unique keys.
	rightDataMap := make(map[interface{}][]map[string]interface{})
	for _, rightRow := range rightData {
		keyVal, ok := rightRow[joinCondition.RightColumn]
		if !ok {
			continue
		} // Skip rows where join key is missing
		rightDataMap[keyVal] = append(rightDataMap[keyVal], rightRow)
	}

	var joinedResult []map[string]interface{}

	// 3. Iterate through the left-side data and probe the hash map.
	for _, leftRow := range leftData {
		leftKeyVal, ok := leftRow[joinCondition.LeftColumn]
		if !ok {
			if params.JoinType == "left" {
				// For left join, if left key is missing, row still included with null right side.
				mergedRow := make(map[string]interface{})
				for k, v := range leftRow {
					mergedRow[k] = v
				}
				// Add nil for all columns from the right side
				if len(rightData) > 0 {
					for rightColName := range rightData[0] {
						mergedRow[params.RightEntityName+"_"+rightColName] = nil
					}
				}
				joinedResult = append(joinedResult, mergedRow)
			}
			continue
		}

		matchingRightRows, found := rightDataMap[leftKeyVal]

		if found {
			// Inner join or successful left join match
			for _, rightRow := range matchingRightRows {
				mergedRow := make(map[string]interface{})
				// Copy left row data
				for k, v := range leftRow {
					mergedRow[k] = v
				}
				// Copy and prefix right row data to avoid name collisions
				for rightColName, rightVal := range rightRow {
					mergedRow[params.RightEntityName+"_"+rightColName] = rightVal
				}
				joinedResult = append(joinedResult, mergedRow)
			}
		} else if params.JoinType == "left" {
			// Left join with no match
			mergedRow := make(map[string]interface{})
			for k, v := range leftRow {
				mergedRow[k] = v
			}
			// Add nil for all columns from the right side
			if len(rightData) > 0 {
				for rightColName := range rightData[0] {
					mergedRow[params.RightEntityName+"_"+rightColName] = nil
				}
			}
			joinedResult = append(joinedResult, mergedRow)
		}
	}

	// 'right' join is not implemented directly, but can be simulated by swapping left/right and doing a left join.
	if params.JoinType == "right" {
		return nil, errors.New("'right' join is not yet implemented; please swap data sources and use 'left' join")
	}

	return joinedResult, nil
}

func (s *jobService) generateOutput(job *models.Job, data []map[string]interface{}) (interface{}, string, error) {
	fileName := fmt.Sprintf("%s.%s", job.UUID, strings.ToLower(string(job.OutputFormat)))
	filePath := filepath.Join(s.resultsBaseDir, fileName)

	var outputDataBytes []byte
	var err error
	var syncReturnData interface{} = data // Default for JSON sync

	switch job.OutputFormat {
	case models.OutputFormatJSON:
		outputDataBytes, err = json.MarshalIndent(data, "", "  ")
		if err != nil {
			return nil, "", fmt.Errorf("failed to marshal result to JSON for job %s: %w", job.UUID, err)
		}
	case models.OutputFormatCSV:
		syncReturnData = nil // For CSV, sync return data is nil; path is used
		var buffer bytes.Buffer
		csvWriter := stdcsv.NewWriter(&buffer)

		if len(data) > 0 {
			var header []string
			// Determine header order consistently.
			// Option 1: Use keys from the first row (order might vary)
			// for key := range data[0] {
			// 	header = append(header, key)
			// }
			// Option 2: If select_columns was the last op, use its column order
			// This needs more sophisticated tracking of column order through ops.
			// For now, simple key iteration from first row.
			// To ensure consistent order, you might sort keys or get them from an "effective schema"

			// A more robust way to get headers, trying to maintain some order:
			if job.Operations != "" {
				var ops []models.OperationStep
				if json.Unmarshal([]byte(job.Operations), &ops) == nil && len(ops) > 0 {
					lastOp := ops[len(ops)-1]
					if lastOp.Type == "select_columns" && len(lastOp.Columns) > 0 {
						header = lastOp.Columns
					} else if lastOp.Type == "aggregate" && len(lastOp.Aggregation.GroupBy) > 0 {
						// For aggregate, header order can be groupBy columns + alias columns
						header = append(header, lastOp.Aggregation.GroupBy...)
						for _, aggFunc := range lastOp.Aggregation.AggFunctions {
							header = append(header, aggFunc.Alias)
						}
					}
				}
			}
			// Fallback if specific order not determined
			if len(header) == 0 && len(data) > 0 {
				log.Printf("Job %s: CSV header order fallback to keys of first data row.", job.UUID)
				for key := range data[0] {
					header = append(header, key)
				}
				// Potentially sort header alphabetically for some consistency
				// sort.Strings(header)
			}

			if len(header) > 0 { // Ensure we have a header before writing
				if err := csvWriter.Write(header); err != nil {
					return nil, "", fmt.Errorf("failed to write CSV header for job %s: %w", job.UUID, err)
				}
				// Write rows
				for _, rowMap := range data {
					var record []string
					for _, hCol := range header {
						val := rowMap[hCol]
						record = append(record, fmt.Sprintf("%v", val))
					}
					if err := csvWriter.Write(record); err != nil {
						return nil, "", fmt.Errorf("failed to write CSV row for job %s: %w", job.UUID, err)
					}
				}
			} else if len(data) > 0 { // Data exists but no header could be determined (should not happen if data[0] has keys)
				log.Printf("Job %s: CSV data exists but no header determined. Skipping CSV write.", job.UUID)
			}
		}
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			return nil, "", fmt.Errorf("error flushing CSV writer for job %s: %w", job.UUID, err)
		}
		outputDataBytes = buffer.Bytes()

	default:
		return nil, "", fmt.Errorf("unsupported output format %s for job %s", job.OutputFormat, job.UUID)
	}

	if err := os.WriteFile(filePath, outputDataBytes, 0644); err != nil {
		return nil, "", fmt.Errorf("failed to write result file %s for job %s: %w", filePath, job.UUID, err)
	}
	log.Printf("Job %s: Result saved to %s", job.UUID, filePath)

	return syncReturnData, filePath, nil
}

func (s *jobService) applySelect(columns []string, data []map[string]interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	for _, row := range data {
		newRow := make(map[string]interface{})
		for _, colName := range columns {
			if val, ok := row[colName]; ok {
				newRow[colName] = val
			} else {
				newRow[colName] = nil // Or skip, or error, based on desired behavior
			}
		}
		result = append(result, newRow)
	}
	return result
}

func (s *jobService) applyFilter(condition models.Condition, data []map[string]interface{}) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	for _, row := range data {
		val, ok := row[condition.Column]
		if !ok {
			// Column not found in this row, policy: treat as not matching or error?
			// For now, let's say it doesn't match if column is absent, unless operator handles nil.
			if condition.Operator == "!=" || condition.Operator == "NOT_CONTAINS" { // or specific 'IS_NULL' / 'IS_NOT_NULL' operators
				// If checking for inequality to a value, and column doesn't exist, it could be considered "not equal".
				// This logic can get complex. For simplicity, skip if column not found unless operator is specifically about nullity.
				continue
			}
			continue
		}

		match, err := s.evaluateCondition(val, condition.Operator, condition.Value)
		if err != nil {
			// Log error or decide if one bad row evaluation fails the whole filter.
			// For now, let's log and skip the row.
			log.Printf("Error evaluating filter condition for column %s: %v. Skipping row.", condition.Column, err)
			continue
		}
		if match {
			result = append(result, row)
		}
	}
	return result, nil
}

// evaluateCondition is a helper for applyFilter
func (s *jobService) evaluateCondition(rowValue interface{}, operator string, conditionValue interface{}) (bool, error) {
	// This is a complex part. Needs robust type handling and operator logic.
	// TODO: Implement proper type conversion and comparison for various operators and types.
	// Example for basic equality on strings and numbers:
	rowStr := fmt.Sprintf("%v", rowValue)
	condStr := fmt.Sprintf("%v", conditionValue)

	switch operator {
	case "=":
		// Attempt numeric comparison if possible, otherwise string
		rowFloat, rOk := convertToFloat64(rowValue)
		condFloat, cOk := convertToFloat64(conditionValue)
		if rOk && cOk {
			return rowFloat == condFloat, nil
		}
		return rowStr == condStr, nil
	case "!=":
		rowFloat, rOk := convertToFloat64(rowValue)
		condFloat, cOk := convertToFloat64(conditionValue)
		if rOk && cOk {
			return rowFloat != condFloat, nil
		}
		return rowStr != condStr, nil
	case ">":
		rowFloat, rOk := convertToFloat64(rowValue)
		condFloat, cOk := convertToFloat64(conditionValue)
		if rOk && cOk {
			return rowFloat > condFloat, nil
		}
		return false, fmt.Errorf("cannot compare non-numeric types with %s", operator)
	case "<":
		rowFloat, rOk := convertToFloat64(rowValue)
		condFloat, cOk := convertToFloat64(conditionValue)
		if rOk && cOk {
			return rowFloat < condFloat, nil
		}
		return false, fmt.Errorf("cannot compare non-numeric types with %s", operator)
	case ">=":
		rowFloat, rOk := convertToFloat64(rowValue)
		condFloat, cOk := convertToFloat64(conditionValue)
		if rOk && cOk {
			return rowFloat >= condFloat, nil
		}
		return false, fmt.Errorf("cannot compare non-numeric types with %s", operator)
	case "<=":
		rowFloat, rOk := convertToFloat64(rowValue)
		condFloat, cOk := convertToFloat64(conditionValue)
		if rOk && cOk {
			return rowFloat <= condFloat, nil
		}
		return false, fmt.Errorf("cannot compare non-numeric types with %s", operator)
	case "CONTAINS":
		return strings.Contains(strings.ToLower(rowStr), strings.ToLower(condStr)), nil
	case "NOT_CONTAINS":
		return !strings.Contains(strings.ToLower(rowStr), strings.ToLower(condStr)), nil
	case "STARTS_WITH":
		return strings.HasPrefix(strings.ToLower(rowStr), strings.ToLower(condStr)), nil
	case "ENDS_WITH":
		return strings.HasSuffix(strings.ToLower(rowStr), strings.ToLower(condStr)), nil

		// TODO: Add support for IN, NOT IN, IS NULL, IS NOT NULL, etc.
	default:
		return false, fmt.Errorf("unsupported filter operator: %s", operator)
	}
}

// Helper to convert interface{} to float64 for comparisons
func convertToFloat64(val interface{}) (float64, bool) {
	if val == nil {
		return 0, false
	}
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(v.Uint()), true
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	case reflect.String:
		f, err := strconv.ParseFloat(v.String(), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func (s *jobService) applyAggregate(agg models.Aggregation, data []map[string]interface{}) ([]map[string]interface{}, error) {
	// This is a complex operation.
	// TODO: Implement robust aggregation logic.
	// 1. Create a map to store intermediate aggregation results: map[groupKeyString]map[alias]aggregatorState
	//    aggregatorState might need to store sum, count for AVG, min, max etc.
	// 2. Iterate through `data`:
	//    a. Construct `groupKeyString` from `agg.GroupBy` columns.
	//    b. For each `aggFunc` in `agg.AggFunctions`:
	//       i. Get the value from `row[aggFunc.Column]`.
	//       ii. Update the `aggregatorState` for that group and alias.
	// 3. Convert the intermediate map into the final `[]map[string]interface{}` format.
	// This requires careful handling of types for SUM/AVG, initialization of aggregators.

	if len(data) == 0 {
		return []map[string]interface{}{}, nil
	}

	groupedResults := make(map[string]map[string]interface{})
	// intermediateAggs will store things like {groupKey: {alias: {sum: X, count: Y}}}
	intermediateAggs := make(map[string]map[string]map[string]float64)

	for _, row := range data {
		groupKeyParts := make([]string, len(agg.GroupBy))
		for i, groupCol := range agg.GroupBy {
			groupKeyParts[i] = fmt.Sprintf("%v", row[groupCol])
		}
		groupKey := strings.Join(groupKeyParts, "||") // Simple concatenation for group key

		if _, ok := groupedResults[groupKey]; !ok {
			groupedResults[groupKey] = make(map[string]interface{})
			for _, colName := range agg.GroupBy { // Add group by columns to the result
				groupedResults[groupKey][colName] = row[colName]
			}
			intermediateAggs[groupKey] = make(map[string]map[string]float64)
		}

		for _, f := range agg.AggFunctions {
			if _, ok := intermediateAggs[groupKey][f.Alias]; !ok {
				intermediateAggs[groupKey][f.Alias] = map[string]float64{"_count": 0, "_sum": 0, "_min": 0, "_max": 0, "_is_min_set": 0, "_is_max_set": 0}
			}

			val, valOk := row[f.Column]
			numVal, numOk := convertToFloat64(val)

			currentAggs := intermediateAggs[groupKey][f.Alias]
			currentAggs["_count"]++ // Always increment count for any function if row matches group

			if valOk && numOk { // Only process numeric for SUM, AVG, MIN, MAX
				currentAggs["_sum"] += numVal
				if currentAggs["_is_min_set"] == 0 || numVal < currentAggs["_min"] {
					currentAggs["_min"] = numVal
					currentAggs["_is_min_set"] = 1
				}
				if currentAggs["_is_max_set"] == 0 || numVal > currentAggs["_max"] {
					currentAggs["_max"] = numVal
					currentAggs["_is_max_set"] = 1
				}
			}
			intermediateAggs[groupKey][f.Alias] = currentAggs
		}
	}

	var finalResults []map[string]interface{}
	for groupKey, groupRow := range groupedResults {
		for _, f := range agg.AggFunctions {
			aggState := intermediateAggs[groupKey][f.Alias]
			switch strings.ToUpper(f.Function) {
			case "COUNT":
				groupRow[f.Alias] = aggState["_count"]
			case "SUM":
				groupRow[f.Alias] = aggState["_sum"]
			case "AVG":
				if aggState["_count"] > 0 {
					groupRow[f.Alias] = aggState["_sum"] / aggState["_count"]
				} else {
					groupRow[f.Alias] = 0 // Or nil, or NaN
				}
			case "MIN":
				if aggState["_is_min_set"] == 1 {
					groupRow[f.Alias] = aggState["_min"]
				} else {
					groupRow[f.Alias] = nil // Or appropriate default if no numeric values were found
				}
			case "MAX":
				if aggState["_is_max_set"] == 1 {
					groupRow[f.Alias] = aggState["_max"]
				} else {
					groupRow[f.Alias] = nil
				}
			default:
				return nil, fmt.Errorf("unsupported aggregation function: %s", f.Function)
			}
		}
		finalResults = append(finalResults, groupRow)
	}

	return finalResults, nil
}
