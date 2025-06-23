// internal/engine/limit_streamer.go
package engine

import "log"

// LimitStreamer 实现了 Streamer 接口，用于限制通过的数据行数
type LimitStreamer struct {
	Limit int
}

// Process 方法传递指定数量的行，然后会耗尽输入 channel 的剩余数据
func (s *LimitStreamer) Process(in <-chan Row, errc chan<- error) <-chan Row {
	out := make(chan Row)

	go func() {
		defer close(out)

		if s.Limit <= 0 {
			// 如果 limit 为 0 或负数，不产生任何数据，但仍需耗尽上游
			for range in {
			}
			return
		}

		count := 0
		for row := range in {
			if count < s.Limit {
				out <- row
				count++
			} else {
				// 已经达到 limit，但我们必须继续从 channel `in` 中读取并丢弃，
				// 直到它关闭，否则上游的 goroutine 会永远阻塞。
				log.Printf("LimitStreamer: Reached limit of %d. Draining remaining input.", s.Limit)
				// Drain the rest of the input channel
				for range in {
				}
				break // Draining is complete, exit the outer loop
			}
		}
	}()

	return out
}
