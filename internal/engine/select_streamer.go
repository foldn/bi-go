// internal/engine/select_streamer.go
package engine

// SelectStreamer 实现了 Streamer 接口，用于选择数据流中的特定列
type SelectStreamer struct {
	Columns []string
}

// Process 方法对流中的每一行进行处理，只保留指定的列
func (s *SelectStreamer) Process(in <-chan Row, errc chan<- error) <-chan Row {
	out := make(chan Row)

	go func() {
		// 确保在 goroutine 退出时关闭输出 channel，以通知下游处理结束
		defer close(out)

		for row := range in {
			// 为每一行创建一个新的 map，只包含需要的列
			newRow := make(Row, len(s.Columns)) // 预分配 map 容量以提高性能

			for _, colName := range s.Columns {
				// 从输入行中查找指定的列
				if val, ok := row[colName]; ok {
					// 如果找到了，就将其添加到新行中
					newRow[colName] = val
				} else {
					// 如果在当前行中找不到指定的列，
					// 我们选择将该列的值设为 nil，以保证输出的数据流结构一致
					// (即所有 map 都有相同的键集合)
					newRow[colName] = nil
				}
			}

			// 将新的、更窄的数据行发送到输出 channel
			out <- newRow
		}
	}()

	return out
}
