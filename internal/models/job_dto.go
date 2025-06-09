package models

type OperationStep struct {
	Type        string       `json:"type" binding:"required,oneof=select_columns filter_rows aggregate"`
	Columns     []string     `json:"columns,omitempty"`     // For select_columns
	Condition   *Condition   `json:"condition,omitempty"`   // For filter_rows
	Aggregation *Aggregation `json:"aggregation,omitempty"` // For aggregate
}

// Condition defines a filter condition
type Condition struct {
	Column   string      `json:"column" binding:"required"`
	Operator string      `json:"operator" binding:"required,oneof=> < >= <= = != CONTAINS NOT_CONTAINS STARTS_WITH ENDS_WITH"` // Add more as needed
	Value    interface{} `json:"value"`                                                                                        // Value can be string, number, bool, []interface{} for IN/NOT IN
}

// Aggregation defines aggregation parameters
type Aggregation struct {
	GroupBy      []string  `json:"groupBy" binding:"required"`
	AggFunctions []AggFunc `json:"aggFunctions" binding:"required,min=1"`
}

// AggFunc defines a single aggregation function
type AggFunc struct {
	Column   string `json:"column" binding:"required"`
	Function string `json:"function" binding:"required,oneof=SUM COUNT AVG MIN MAX"` // Add more as needed
	Alias    string `json:"alias" binding:"required"`
}
