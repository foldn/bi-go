package models

type EntityInfo struct {
	Name string `json:"name"`
	Type string `json:"type"` // e.g., "TABLE", "VIEW", "FILE"
}

// ColumnInfo represents a column within an entity.
type ColumnInfo struct {
	Name            string `json:"name"`
	DatabaseType    string `json:"databaseType"` // Original type from the DB
	ScanType        string `json:"scanType"`     // Go type `reflect.TypeOf(value).Name()` from sql.ColumnType
	IsNullable      bool   `json:"isNullable"`
	OrdinalPosition int    `json:"ordinalPosition"`
	// Add more fields if needed: DefaultValue, MaxLength, Precision, Scale, IsPrimaryKey, etc.
}

// EntitySchema combines entity information with its columns.
type EntitySchema struct {
	EntityInfo
	Columns []ColumnInfo `json:"columns"`
}
