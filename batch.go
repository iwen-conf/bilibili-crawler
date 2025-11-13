package bilibili_crawler

import (
	"context"
	"sync"
	"time"
)

// BatchOptions 批量处理选项
type BatchOptions struct {
	// 进度回调
	OnProgress func(completed, total int, current *CrawlResult)
	// 错误处理
	OnError func(videoID string, err error)
	// 成功回调
	OnSuccess func(info *VideoInfo)
}

// FetchBatch 批量获取视频信息
func (c *Client) FetchBatch(videoIDs []string, options *BatchOptions) []*CrawlResult {
	return c.FetchBatchWithContext(context.Background(), videoIDs, options)
}

// FetchBatchWithContext 带上下文批量获取视频信息
func (c *Client) FetchBatchWithContext(ctx context.Context, videoIDs []string, options *BatchOptions) []*CrawlResult {
	if options == nil {
		options = &BatchOptions{}
	}

	results := make([]*CrawlResult, len(videoIDs))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, c.config.MaxWorkers)

	// 创建结果通道
	resultChan := make(chan struct {
		index  int
		result *CrawlResult
	}, len(videoIDs))

	// 启动工作协程
	for i, videoID := range videoIDs {
		wg.Add(1)
		go func(idx int, id string) {
			defer wg.Done()

			// 获取信号量
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				resultChan <- struct {
					index  int
					result *CrawlResult
				}{
					index: idx,
					result: &CrawlResult{
						VideoID: id,
						Error:   ctx.Err(),
					},
				}
				return
			}

			// 添加延迟以避免过快请求
			if idx > 0 && c.config.RequestDelay > 0 {
				select {
				case <-time.After(c.config.RequestDelay):
				case <-ctx.Done():
					resultChan <- struct {
						index  int
						result *CrawlResult
					}{
						index: idx,
						result: &CrawlResult{
							VideoID: id,
							Error:   ctx.Err(),
						},
					}
					return
				}
			}

			// 执行爬取
			info, err := c.FetchVideoWithContext(ctx, id)
			result := &CrawlResult{
				VideoID: id,
				Info:    info,
				Error:   err,
			}

			// 发送结果
			resultChan <- struct {
				index  int
				result *CrawlResult
			}{
				index:  idx,
				result: result,
			}

			// 回调处理
			if err != nil {
				if options.OnError != nil {
					options.OnError(id, err)
				}
			} else if options.OnSuccess != nil {
				options.OnSuccess(info)
			}
		}(i, videoID)
	}

	// 启动结果收集协程
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	completed := 0
	for res := range resultChan {
		results[res.index] = res.result
		completed++

		// 进度回调
		if options.OnProgress != nil {
			options.OnProgress(completed, len(videoIDs), res.result)
		}
	}

	return results
}

// FetchBatchConcurrent 并发批量获取（返回通道）
func (c *Client) FetchBatchConcurrent(ctx context.Context, videoIDs []string) <-chan *CrawlResult {
	resultChan := make(chan *CrawlResult)

	go func() {
		defer close(resultChan)

		var wg sync.WaitGroup
		semaphore := make(chan struct{}, c.config.MaxWorkers)

		for i, videoID := range videoIDs {
			// 检查上下文
			select {
			case <-ctx.Done():
				resultChan <- &CrawlResult{
					VideoID: videoID,
					Error:   ctx.Err(),
				}
				continue
			default:
			}

			wg.Add(1)
			go func(idx int, id string) {
				defer wg.Done()

				// 获取信号量
				select {
				case semaphore <- struct{}{}:
					defer func() { <-semaphore }()
				case <-ctx.Done():
					resultChan <- &CrawlResult{
						VideoID: id,
						Error:   ctx.Err(),
					}
					return
				}

				// 添加延迟
				if idx > 0 && c.config.RequestDelay > 0 {
					select {
					case <-time.After(c.config.RequestDelay):
					case <-ctx.Done():
						resultChan <- &CrawlResult{
							VideoID: id,
							Error:   ctx.Err(),
						}
						return
					}
				}

				// 执行爬取
				info, err := c.FetchVideoWithContext(ctx, id)
				resultChan <- &CrawlResult{
					VideoID: id,
					Info:    info,
					Error:   err,
				}
			}(i, videoID)
		}

		wg.Wait()
	}()

	return resultChan
}

// BatchStats 批量处理统计
type BatchStats struct {
	Total     int
	Success   int
	Failed    int
	Duration  time.Duration
	StartTime time.Time
	EndTime   time.Time
}

// CalculateBatchStats 计算批量处理统计
func CalculateBatchStats(results []*CrawlResult, startTime, endTime time.Time) *BatchStats {
	stats := &BatchStats{
		Total:     len(results),
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}

	for _, result := range results {
		if result.Error == nil {
			stats.Success++
		} else {
			stats.Failed++
		}
	}

	return stats
}
