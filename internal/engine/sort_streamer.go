package engine

import (
	"fmt"
	"github.com/foldn/bi-go/internal/models"
	"github.com/foldn/bi-go/pkg/utils"
	"sort"
	"strings"
)

type SortStreamer struct {
	SortParams []models.SortParam
}

// Process 方法消费整个输入流，在内存中排序，然后产生输出流
func (s *SortStreamer) Process(in <-chan Row, errc chan<- error) <-chan Row {
	out := make(chan Row)
	go func() {
		defer close(out)

		// 1. 读取所有数据到内存切片中
		var allRows []Row
		for row := range in {
			allRows = append(allRows, row)
		}

		// 如果没有数据，直接返回
		if len(allRows) == 0 {
			return
		}

		// 2. 使用 sort.Slice 和一个自定义的比较函数对切片进行排序
		sort.Slice(allRows, func(i, j int) bool {
			// less 函数：如果 allRows[i]应该排在 allRows[j] 前面，则返回 true

			for _, p := range s.SortParams {
				valI := allRows[i][p.Column]
				valJ := allRows[j][p.Column]

				order := strings.ToUpper(p.Order)
				if order == "" {
					order = "ASC" // 默认为升序
				}

				// 处理 nil 值：我们约定 nil 总是最小的
				if valI == nil && valJ != nil {
					return order == "ASC" // 在升序中，nil (i) 在前
				}
				if valI != nil && valJ == nil {
					return order == "DESC" // 在降序中，nil (j) 在前，所以 i 在后
				}
				if valI == nil && valJ == nil {
					continue // 两者都为 nil，无法比较，继续下一个排序条件
				}

				// 尝试数值比较
				numI, iIsNum := utils.ConvertToFloat64(valI)
				numJ, jIsNum := utils.ConvertToFloat64(valJ)

				if iIsNum && jIsNum {
					if numI != numJ {
						if order == "ASC" {
							return numI < numJ
						}
						return numI > numJ
					}
					// 如果数值相等，继续下一个排序条件
					continue
				}

				// 如果不是数字或数字相等，则进行字符串比较
				strI := fmt.Sprintf("%v", valI)
				strJ := fmt.Sprintf("%v", valJ)

				if strI != strJ {
					if order == "ASC" {
						return strI < strJ
					}
					return strI > strJ
				}

				// 如果当前排序键的值相等，继续判断下一个排序键
			}

			// 如果所有排序键的值都相等，保持原始顺序
			return false
		})

		// 3. 将排序后的结果逐条发送到输出 channel
		for _, row := range allRows {
			out <- row
		}
	}()
	return out
}
