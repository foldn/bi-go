package engine

import (
	"fmt"
	"github.com/foldn/bi-go/internal/models"
	"github.com/foldn/bi-go/pkg/utils"
)

type FilterStreamer struct {
	Condition models.Condition
}

func (s *FilterStreamer) Process(in <-chan Row, errc chan<- error) <-chan Row {
	out := make(chan Row)
	go func() {
		defer close(out)
		for row := range in {
			// evaluateCondition 是我们在 service 中已有的辅助函数，可以移到 engine 或 util 包中
			match, err := utils.EvaluateCondition(row[s.Condition.Column], s.Condition.Operator, s.Condition.Value)
			if err != nil {
				// 发送错误并停止处理此流
				errc <- fmt.Errorf("error evaluating filter condition: %w", err)
				return
			}
			if match {
				out <- row
			}
		}
	}()
	return out
}
