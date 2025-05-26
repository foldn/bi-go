package models

type EntityInfo struct {
	Name string `json:"name"`
	Type string `json:"type"` // e.g., "TABLE", "VIEW", "FILE"
}

// ColumnInfo represents a column within an entity.
type ColumnInfo struct {
	Name            string `json:"name"`
	DatabaseType    string `json:"database_type"` // Original type from the DB
	ScanType        string `json:"scan_type"`     // Go type `reflect.TypeOf(value).Name()` from sql.ColumnType
	IsNullable      bool   `json:"is_nullable"`
	OrdinalPosition int    `json:"ordinal_position"`
	// Add more fields if needed: DefaultValue, MaxLength, Precision, Scale, IsPrimaryKey, etc.
}

// EntitySchema combines entity information with its columns.
type EntitySchema struct {
	EntityInfo
	Columns []ColumnInfo `json:"columns"`
}
