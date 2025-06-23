// internal/engine/calculate_field_streamer.go
package engine

import (
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"log"
)

// CalculateFieldStreamer 实现了 Streamer 接口，用于添加计算字段
type CalculateFieldStreamer struct {
	NewColumnName string
	Program       *vm.Program // Pre-compiled expression program
}

// Process 方法对流中的每一行应用表达式并添加新列
func (s *CalculateFieldStreamer) Process(in <-chan Row, errc chan<- error) <-chan Row {
	out := make(chan Row)
	go func() {
		defer close(out)
		for row := range in {
			// 使用预编译的程序在当前行上运行表达式
			result, err := expr.Run(s.Program, row)
			if err != nil {
				// 如果单行计算出错 (例如类型不匹配)，记录错误并将新列的值设为 nil
				// 这样可以保证数据流的健壮性，不会因个别脏数据而中断整个流程
				log.Printf(
					"Failed to run expression for column '%s' on row: %v. Error: %v. Setting result to nil.",
					s.NewColumnName,
					row,
					err,
				)
				row[s.NewColumnName] = nil
			} else {
				row[s.NewColumnName] = result
			}
			out <- row
		}
	}()
	return out
}
