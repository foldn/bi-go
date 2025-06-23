// internal/engine/aggregate_streamer.go
package engine

import (
	"fmt"
	"github.com/foldn/bi-go/internal/models"
	"github.com/foldn/bi-go/pkg/utils"
	"strings"
)

// AggregateStreamer 实现了 Streamer 接口，用于对数据流进行聚合
type AggregateStreamer struct {
	Aggregation models.Aggregation
}

// Process 方法消费整个输入流，进行聚合计算，然后输出结果
func (s *AggregateStreamer) Process(in <-chan Row, errc chan<- error) <-chan Row {
	out := make(chan Row)
	go func() {
		defer close(out)

		// 1. 累积阶段：消费所有输入数据并进行中间聚合
		// groupedResults 存储分组键列和最终的聚合结果
		groupedResults := make(map[string]Row)
		// intermediateAggs 存储聚合过程中的状态，如 sum 和 count，用于计算 AVG 等
		intermediateAggs := make(map[string]map[string]map[string]float64)

		for row := range in {
			// 构造分组键 (group key)
			var groupKeyParts []string
			for _, groupCol := range s.Aggregation.GroupBy {
				groupKeyParts = append(groupKeyParts, fmt.Sprintf("%v", row[groupCol]))
			}
			groupKey := strings.Join(groupKeyParts, "||")

			// 如果是新的分组，则初始化
			if _, ok := groupedResults[groupKey]; !ok {
				groupedResults[groupKey] = make(Row)
				for _, colName := range s.Aggregation.GroupBy {
					groupedResults[groupKey][colName] = row[colName]
				}
				intermediateAggs[groupKey] = make(map[string]map[string]float64)
			}

			// 对每个聚合函数更新中间状态
			for _, f := range s.Aggregation.AggFunctions {
				if _, ok := intermediateAggs[groupKey][f.Alias]; !ok {
					intermediateAggs[groupKey][f.Alias] = map[string]float64{"_count": 0, "_sum": 0, "_min": 0, "_max": 0, "_is_min_set": 0}
				}

				val, valOk := row[f.Column]
				numVal, numOk := utils.ConvertToFloat64(val)

				aggState := intermediateAggs[groupKey][f.Alias]
				aggState["_count"]++

				if valOk && numOk {
					aggState["_sum"] += numVal
					if aggState["_is_min_set"] == 0 || numVal < aggState["_min"] {
						aggState["_min"] = numVal
						aggState["_is_min_set"] = 1
					}
					if numVal > aggState["_max"] { // max可以直接比较，不需要 is_set
						aggState["_max"] = numVal
					}
				}
			}
		}

		// 2. 完成与输出阶段：在所有数据都处理完毕后，计算最终值并发送
		for groupKey, groupRow := range groupedResults {
			for _, f := range s.Aggregation.AggFunctions {
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
						groupRow[f.Alias] = 0
					}
				case "MIN":
					if aggState["_is_min_set"] == 1 {
						groupRow[f.Alias] = aggState["_min"]
					} else {
						groupRow[f.Alias] = nil
					}
				case "MAX":
					// 如果 count > 0 至少有一个值，但可能都不是数字，所以max可能是0
					// 更严谨的判断是需要一个 is_max_set
					groupRow[f.Alias] = aggState["_max"]
				}
			}
			out <- groupRow
		}
	}()
	return out
}
