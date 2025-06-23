package engine

import (
	"fmt"
	"github.com/expr-lang/expr"
	"github.com/foldn/bi-go/internal/models"
)

// Row 代表在数据流中传递的单行数据
type Row map[string]interface{}

// Streamer 是处理数据流的单个操作单元的接口
type Streamer interface {
	// Process 方法接收一个输入channel，返回一个输出channel。
	// 它在一个新的 goroutine 中异步处理数据。
	// errc 用于从 goroutine 中传递错误。
	Process(in <-chan Row, errc chan<- error) <-chan Row
}

// StreamerContext 包含了执行 Streamer 所需的上下文信息
type StreamerContext struct {
	//// 用于 'join' 操作，需要能够获取其他数据源的数据
	Fetcher DataFetcher
}

// NewStreamer 是一个工厂函数，根据操作定义创建具体的 Streamer 实例
func NewStreamer(op models.OperationStep, sCtx StreamerContext) (Streamer, error) {
	switch op.Type {
	case "select_columns":
		return &SelectStreamer{Columns: op.Columns}, nil
	case "filter_rows":
		return &FilterStreamer{Condition: *op.Condition}, nil
	case "calculate_field":
		// 'expr' 库的 program 编译一次后可复用，是线程安全的
		program, err := expr.Compile(op.Calculation.Expression, expr.Env(map[string]interface{}{}))
		if err != nil {
			return nil, fmt.Errorf("failed to compile expression: %w", err)
		}
		return &CalculateFieldStreamer{
			NewColumnName: op.Calculation.NewColumnName,
			Program:       program,
		}, nil
	case "sort_rows":
		return &SortStreamer{SortParams: op.Sort}, nil
	case "aggregate":
		return &AggregateStreamer{Aggregation: *op.Aggregation}, nil
	case "join":
		return &JoinStreamer{
			JoinParams: *op.Join,
			sCtx:       sCtx,
		}, nil
	case "limit_rows":
		return &LimitStreamer{Limit: *op.Limit}, nil
	default:
		return nil, fmt.Errorf("unknown operation type for streaming: %s", op.Type)
	}
}
