// internal/engine/join_streamer.go
package engine

import (
	"errors"
	"fmt"
	"github.com/foldn/bi-go/internal/models"
	"log"
	"log/slog"
)

// JoinStreamer 实现了 Streamer 接口，用于对数据流进行连接操作
type JoinStreamer struct {
	JoinParams models.JoinParams
	sCtx       StreamerContext // 需要用它来获取 DataSourceService
}

// Process 方法执行流式哈希连接
func (s *JoinStreamer) Process(in <-chan Row, errc chan<- error) <-chan Row {
	out := make(chan Row)

	go func() {
		defer close(out)

		if len(s.JoinParams.On) != 1 {
			errc <- errors.New("join currently supports only a single 'on' condition")
			return
		}
		joinOn := s.JoinParams.On[0]

		// 1. 构建阶段 (Build Phase): 完全加载右表数据到内存哈希表中
		slog.Info(fmt.Sprintf("Join build phase: fetching right table '%s'", s.JoinParams.RightEntityName))
		rightData, err := s.sCtx.Fetcher.FetchDataForEntity("", s.JoinParams.RightDataSourceID, s.JoinParams.RightEntityName, nil)
		if err != nil {
			errc <- fmt.Errorf("failed to fetch right side of join: %w", err)
			return
		}

		rightDataMap := make(map[interface{}][]Row)
		for _, rightRow := range rightData {
			keyVal, ok := rightRow[joinOn.RightColumn]
			if !ok {
				continue
			}
			rightDataMap[keyVal] = append(rightDataMap[keyVal], rightRow)
		}
		log.Printf("Join build phase: completed. Hash map contains %d unique keys.", len(rightDataMap))

		// 2. 探测阶段 (Probe Phase): 流式处理左表数据
		for leftRow := range in {
			leftKeyVal, ok := leftRow[joinOn.LeftColumn]

			// 如果左表的连接键不存在，或在右表中找不到匹配项
			if !ok {
				if s.JoinParams.JoinType == "left" {
					out <- s.mergeRows(leftRow, nil) // 左连接，保留左行，右行置空
				}
				continue
			}

			matchingRightRows, found := rightDataMap[leftKeyVal]
			if !found {
				if s.JoinParams.JoinType == "left" {
					out <- s.mergeRows(leftRow, nil) // 左连接，保留左行，右行置空
				}
				continue
			}

			// 找到匹配项，为每一个匹配的右行生成一个合并结果
			for _, rightRow := range matchingRightRows {
				out <- s.mergeRows(leftRow, rightRow)
			}
		}
	}()

	return out
}

// mergeRows 是一个辅助函数，用于合并左右两行
func (s *JoinStreamer) mergeRows(leftRow, rightRow Row) Row {
	mergedRow := make(Row)
	// 复制左表所有列
	for k, v := range leftRow {
		mergedRow[k] = v
	}

	if rightRow != nil {
		// 复制右表所有列，并添加前缀以防列名冲突
		for k, v := range rightRow {
			mergedRow[s.JoinParams.RightEntityName+"_"+k] = v
		}
	} else {
		// 如果是左连接未匹配的情况，需要为右表的所有列添加 nil 值
		// 这需要我们知道右表的 schema，一种简化的方式是假设所有右表都有相同的列
		// 更健壮的方式是在构建阶段就保存好右表的列名列表
		// For now, this case is handled implicitly by not adding any right-side keys if rightRow is nil.
		// A more complete implementation might explicitly add nil values for all right-side columns.
	}

	return mergedRow
}
