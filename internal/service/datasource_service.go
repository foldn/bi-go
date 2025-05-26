package service

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/foldn/bi-go/internal/models"
	"github.com/foldn/bi-go/internal/repository"
	"gorm.io/gorm"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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

func (s *dataSourceService) getDBConnection(dataSourceID uint) (*sql.DB, *models.DataSource, error) {
	ds, err := s.repo.GetByID(dataSourceID)
	if err != nil {
		return nil, nil, fmt.Errorf("datasource with ID %d not found: %w", dataSourceID, err)
	}

	var dsn string
	var driverName string

	switch ds.Type {
	case models.PostgreSQL:
		driverName = "postgres"
		// Example DSN: "postgres://user:password@host:port/dbname?sslmode=disable"
		// Ensure your ds.Password is handled securely if it contains special characters for URL
		password := url.QueryEscape(ds.Password)
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", ds.Username, password, ds.Host, ds.Port, ds.DBName)
		if ds.OtherParams != "" {
			dsn += "?" + ds.OtherParams // e.g., "sslmode=disable"
		}
	case models.MySQL:
		driverName = "mysql"
		// Example DSN: "user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", ds.Username, ds.Password, ds.Host, ds.Port, ds.DBName)
		if ds.OtherParams != "" {
			dsn += "?" + ds.OtherParams // e.g., "charset=utf8mb4&parseTime=True&loc=Local"
		} else { // Default recommended params for MySQL
			dsn += "?charset=utf8mb4&parseTime=True&loc=Local"
		}
	case models.SQLite:
		driverName = "sqlite3"
		dsn = ds.FilePath // For SQLite, DSN is usually the file path
		if ds.OtherParams != "" {
			dsn += "?" + ds.OtherParams
		}
	case models.ClickHouse: // Using clickhouse-go/v2 native TCP protocol
		driverName = "clickhouse"
		// Example DSN: "clickhouse://user:password@host:port/dbname?dial_timeout=200ms&compress=lz4"
		// HTTP protocol DSN: "http://user:password@host:port/dbname?dial_timeout=200ms..."
		// Default port for native is 9000, HTTP is 8123
		dsn = fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s", ds.Username, ds.Password, ds.Host, ds.Port, ds.DBName)
		if ds.OtherParams != "" {
			dsn += "?" + ds.OtherParams
		}

	default:
		return nil, ds, fmt.Errorf("unsupported database type for direct connection: %s", ds.Type)
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, ds, fmt.Errorf("failed to open connection to %s: %w", ds.Type, err)
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, ds, fmt.Errorf("failed to ping %s: %w", ds.Type, err)
	}
	return db, ds, nil
}

// GetDataSourceSchema retrieves the list of entities (tables/views/files) for a data source.
func (s *dataSourceService) GetDataSourceSchema(dataSourceID uint) (interface{}, error) {
	dsConfig, err := s.repo.GetByID(dataSourceID)
	if err != nil {
		return nil, fmt.Errorf("datasource with ID %d not found: %w", dataSourceID, err)
	}

	if dsConfig.Type == models.CSV {
		// For a CSV, the "schema" is essentially the file itself as an entity.
		// We could check if the file exists.
		if _, err := os.Stat(dsConfig.FilePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("CSV file not found at path: %s", dsConfig.FilePath)
		}
		return []models.EntityInfo{{Name: filepath.Base(dsConfig.FilePath), Type: "FILE"}}, nil
	}

	db, _, err := s.getDBConnection(dataSourceID) // dsConfig already fetched
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entities []models.EntityInfo
	var query string
	var rows *sql.Rows

	switch dsConfig.Type {
	case models.PostgreSQL:
		query = `SELECT table_name, table_type FROM information_schema.tables 
		         WHERE table_schema = 'public' OR table_schema = current_schema()
		         ORDER BY table_name;`
		rows, err = db.Query(query)
	case models.MySQL:
		query = `SELECT table_name, table_type FROM information_schema.tables 
		         WHERE table_schema = DATABASE() 
		         ORDER BY table_name;`
		rows, err = db.Query(query)
	case models.SQLite:
		query = `SELECT name, type FROM sqlite_master 
		         WHERE type IN ('table', 'view') AND name NOT LIKE 'sqlite_%' 
		         ORDER BY name;`
		rows, err = db.Query(query)
	case models.ClickHouse:
		// ClickHouse doesn't have a direct 'TABLE'/'VIEW' type in system.tables as clearly
		// It uses engine to distinguish. We'll list tables.
		// Views in ClickHouse are often just tables with specific engines like 'View' or 'MaterializedView'
		query = `SELECT name, engine AS type FROM system.tables 
		         WHERE database = currentDatabase()
		         ORDER BY name;`
		rows, err = db.Query(query)
	default:
		return nil, fmt.Errorf("schema retrieval not supported for type: %s", dsConfig.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query schema for %s: %w", dsConfig.Type, err)
	}
	defer rows.Close()

	for rows.Next() {
		var entity models.EntityInfo
		if err := rows.Scan(&entity.Name, &entity.Type); err != nil {
			log.Printf("Error scanning entity row for %s: %v", dsConfig.Type, err)
			continue // Or return error
		}
		// Normalize type for ClickHouse, as engine is more specific
		if dsConfig.Type == models.ClickHouse {
			if strings.Contains(strings.ToLower(entity.Type), "view") {
				entity.Type = "VIEW"
			} else {
				entity.Type = "TABLE" // Simplification
			}
		}
		entities = append(entities, entity)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schema rows for %s: %w", dsConfig.Type, err)
	}

	return entities, nil
}

// GetDataSourceEntitySchema retrieves column information for a specific entity.
func (s *dataSourceService) GetDataSourceEntitySchema(dataSourceID uint, entityName string) (interface{}, error) {
	dsConfig, err := s.repo.GetByID(dataSourceID)
	if err != nil {
		return nil, fmt.Errorf("datasource with ID %d not found: %w", dataSourceID, err)
	}

	if dsConfig.Type == models.CSV {
		// For CSV, read the header row
		file, err := os.Open(dsConfig.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open CSV file %s: %w", dsConfig.FilePath, err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		header, err := reader.Read()
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV header from %s: %w", dsConfig.FilePath, err)
		}

		var columns []models.ColumnInfo
		for i, colName := range header {
			columns = append(columns, models.ColumnInfo{
				Name:            colName,
				DatabaseType:    "TEXT",   // CSVs are typically schemaless, default to TEXT
				ScanType:        "string", // Assume string initially
				IsNullable:      true,     // Usually true for CSVs
				OrdinalPosition: i + 1,
			})
		}
		return models.EntitySchema{
			EntityInfo: models.EntityInfo{Name: filepath.Base(dsConfig.FilePath), Type: "FILE"},
			Columns:    columns,
		}, nil
	}

	db, _, err := s.getDBConnection(dataSourceID) // dsConfig already fetched
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var columns []models.ColumnInfo
	var query string
	var rows *sql.Rows

	// Using placeholder for entityName to prevent SQL injection.
	// Proper quoting might be needed if entity names can contain special chars,
	// but for standard identifiers, direct use in query is common with information_schema.

	switch dsConfig.Type {
	case models.PostgreSQL:
		query = `SELECT column_name, data_type, udt_name, is_nullable, ordinal_position
		         FROM information_schema.columns
		         WHERE table_schema = 'public' AND table_name = $1
		         ORDER BY ordinal_position;`
		rows, err = db.Query(query, entityName)
	case models.MySQL:
		query = `SELECT column_name, data_type, column_type, is_nullable, ordinal_position
		         FROM information_schema.columns
		         WHERE table_schema = DATABASE() AND table_name = ?
		         ORDER BY ordinal_position;`
		rows, err = db.Query(query, entityName)
	case models.SQLite:
		// PRAGMA table_info returns: cid, name, type, notnull, dflt_value, pk
		query = fmt.Sprintf("PRAGMA table_info(%s);", quoteIdentifierSQLite(entityName)) // Basic quoting
		rows, err = db.Query(query)
	case models.ClickHouse:
		query = `SELECT name, type, is_nullable_expr AS is_nullable, position
		         FROM system.columns
		         WHERE database = currentDatabase() AND table = ?
		         ORDER BY position;`
		rows, err = db.Query(query, entityName)
	default:
		return nil, fmt.Errorf("entity schema retrieval not supported for type: %s", dsConfig.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query entity schema for %s on table %s: %w", dsConfig.Type, entityName, err)
	}
	defer rows.Close()

	sqlColumnTypes, err := rows.ColumnTypes()
	if err != nil && dsConfig.Type != models.SQLite { // SQLite PRAGMA doesn't support ColumnTypes well
		log.Printf("Warning: Could not get rows.ColumnTypes() for %s: %v", dsConfig.Type, err)
	}

	idx := 0
	for rows.Next() {
		var col models.ColumnInfo
		var isNullableStr string  // For PG, MySQL
		var notNullInt int        // For SQLite (0 or 1)
		var isNullableUInt8 uint8 // For ClickHouse (0 or 1)

		// Scan logic needs to be dialect-specific due to different columns returned by schema queries
		switch dsConfig.Type {
		case models.PostgreSQL:
			// column_name, data_type, udt_name, is_nullable, ordinal_position
			var udtName string // PostgreSQL specific user defined type
			err = rows.Scan(&col.Name, &col.DatabaseType, &udtName, &isNullableStr, &col.OrdinalPosition)
			if udtName != "" && udtName != col.DatabaseType { // Prefer udt_name if available and different
				col.DatabaseType = udtName
			}
			col.IsNullable = (isNullableStr == "YES")
		case models.MySQL:
			// column_name, data_type, column_type, is_nullable, ordinal_position
			var columnType string // MySQL specific, often more detailed like VARCHAR(255)
			err = rows.Scan(&col.Name, &col.DatabaseType, &columnType, &isNullableStr, &col.OrdinalPosition)
			if columnType != "" { // Prefer the more detailed column_type
				col.DatabaseType = columnType
			}
			col.IsNullable = (isNullableStr == "YES")
		case models.SQLite:
			// cid, name, type, notnull, dflt_value, pk
			var cid int
			var dfltValue sql.NullString
			var pk int
			err = rows.Scan(&cid, &col.Name, &col.DatabaseType, &notNullInt, &dfltValue, &pk)
			col.IsNullable = (notNullInt == 0)
			col.OrdinalPosition = cid + 1 // cid is 0-indexed
		case models.ClickHouse:
			// name, type, is_nullable_expr, position
			err = rows.Scan(&col.Name, &col.DatabaseType, &isNullableUInt8, &col.OrdinalPosition)
			col.IsNullable = (isNullableUInt8 == 1)
		}

		if err != nil {
			log.Printf("Error scanning column row for %s, table %s: %v", dsConfig.Type, entityName, err)
			continue
		}

		// Get ScanType (Go reflect type) if possible
		if sqlColumnTypes != nil && idx < len(sqlColumnTypes) && sqlColumnTypes[idx] != nil {
			// For SQLite's PRAGMA, ColumnTypes() might not be reliable or available.
			// We try to get it for others.
			// The index for sqlColumnTypes should correspond to the SELECT order of the PRAGMA query's 'name' column if possible
			// However, simpler to just use a default or leave empty if not easily determinable for SQLite's PRAGMA.
			if dsConfig.Type == models.SQLite {
				// Manual mapping or heuristics for SQLite types to Go types might be needed here if ScanType is crucial.
				// For now, we can leave it or assign a common default like "string" or "interface{}"
				col.ScanType = getScanTypeFromSQLiteType(col.DatabaseType) // Example helper
			} else if st := sqlColumnTypes[idx].ScanType(); st != nil { // find the correct column index for name
				//nameColIdx := -1
				//for i, ct := range sqlColumnTypes {
				//	if ct.Name() == "name" || ct.Name() == "column_name" { // common name for the column name itself
				//		nameColIdx = i
				//		break
				//	}
				//}
				col.ScanType = col.DatabaseType // Simplified: This is not the Go ScanType.
			}
		} else {
			// Fallback for SQLite or if ColumnTypes not available
			if dsConfig.Type == models.SQLite {
				col.ScanType = getScanTypeFromSQLiteType(col.DatabaseType)
			} else {
				// For other DBs, if ColumnTypes() failed, we might leave it blank or use DatabaseType
				col.ScanType = "unknown" // Placeholder
			}
		}

		columns = append(columns, col)
		idx++
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating entity schema rows for %s: %w", dsConfig.Type, err)
	}

	entityInfo := models.EntityInfo{Name: entityName}
	// We might not easily get "TABLE" vs "VIEW" here without another query or more complex initial query.
	// For simplicity, we'll determine it based on the `GetDataSourceSchema` call if needed, or just leave it generic.
	entityList, err := s.GetDataSourceSchema(dataSourceID)
	if err == nil {
		if eiList, ok := entityList.([]models.EntityInfo); ok {
			for _, ei := range eiList {
				if ei.Name == entityName {
					entityInfo.Type = ei.Type
					break
				}
			}
		}
	}
	if entityInfo.Type == "" {
		entityInfo.Type = "TABLE" // Default if not found
	}

	return models.EntitySchema{
		EntityInfo: entityInfo,
		Columns:    columns,
	}, nil
}

// Helper for SQLite to map common types to Go-like scan types (simplified)
func getScanTypeFromSQLiteType(sqliteType string) string {
	upperType := strings.ToUpper(sqliteType)
	if strings.Contains(upperType, "INT") {
		return "int64"
	}
	if strings.Contains(upperType, "CHAR") || strings.Contains(upperType, "TEXT") || strings.Contains(upperType, "CLOB") {
		return "string"
	}
	if strings.Contains(upperType, "REAL") || strings.Contains(upperType, "FLOA") || strings.Contains(upperType, "DOUB") {
		return "float64"
	}
	if strings.Contains(upperType, "BLOB") {
		return "[]byte"
	}
	if strings.Contains(upperType, "BOOL") {
		return "bool"
	}
	if strings.Contains(upperType, "DATE") || strings.Contains(upperType, "TIME") {
		return "time.Time" // Assuming parseTime=True or similar handling
	}
	return "interface{}" // Default
}

// Basic identifier quoting for SQLite to prevent issues with names containing spaces or keywords
func quoteIdentifierSQLite(identifier string) string {
	// SQLite uses double quotes for identifiers, or backticks like MySQL, or square brackets like SQL Server.
	// Double quotes are SQL standard.
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}
