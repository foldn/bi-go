package models

type OperationStep struct {
	Type        string       `json:"type" binding:"required,oneof=select_columns filter_rows aggregate join calculate_field sort_rows limit_rows"`
	Columns     []string     `json:"columns,omitempty"`     // For select_columns
	Condition   *Condition   `json:"condition,omitempty"`   // For filter_rows
	Aggregation *Aggregation `json:"aggregation,omitempty"` // For aggregate
	Join        *JoinParams  `json:"join,omitempty"`        // For join
	Calculation *CalcParams  `json:"calculation,omitempty"` // For calculate_field
	Sort        []SortParam  `json:"sort,omitempty"`        // For sort_rows
	Limit       *int         `json:"limit,omitempty"`       // For limit_rows
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

// JoinParams defines the parameters for a join operation.
type JoinParams struct {
	RightDataSourceID uint     `json:"rightDatasourceId" binding:"required"`
	RightEntityName   string   `json:"rightEntityName" binding:"required"`
	JoinType          string   `json:"joinType" binding:"required,oneof=inner left"` // Starting with inner and left join
	On                []JoinOn `json:"on" binding:"required,min=1"`
}

// JoinOn defines a single join condition.
type JoinOn struct {
	LeftColumn  string `json:"leftColumn" binding:"required"`
	RightColumn string `json:"rightColumn" binding:"required"`
}

// CalcParams defines the parameters for a calculate_field operation.
type CalcParams struct {
	NewColumnName string `json:"newColumnName" binding:"required"`
	Expression    string `json:"expression" binding:"required"`
}

// SortParam defines a single sort condition.
type SortParam struct {
	Column string `json:"column" binding:"required"`
	Order  string `json:"order,omitempty" binding:"omitempty,oneof=ASC DESC"`
}
