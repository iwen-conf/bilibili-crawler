package bilibili_crawler

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Client 爬虫客户端
type Client struct {
	config       *Config
	httpClient   *http.Client
	proxyPool    *ProxyPool
	userAgentIdx int
	mu           sync.Mutex
	rateLimiter  chan struct{}
}

// NewClient 创建新的爬虫客户端
func NewClient(config *Config) *Client {
	if config == nil {
		config = DefaultConfig()
	}

	client := &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
		rateLimiter: make(chan struct{}, config.MaxWorkers),
	}

	// 初始化代理池
	if config.UseProxy && len(config.ProxyURLs) > 0 {
		client.proxyPool = NewProxyPool(config.ProxyURLs, 5*time.Minute)
	}

	return client
}

// FetchVideo 获取单个视频信息
func (c *Client) FetchVideo(videoID string) (*VideoInfo, error) {
	return c.FetchVideoWithContext(context.Background(), videoID)
}

// FetchVideoWithContext 带上下文获取视频信息
func (c *Client) FetchVideoWithContext(ctx context.Context, videoID string) (*VideoInfo, error) {
	// 限流
	select {
	case c.rateLimiter <- struct{}{}:
		defer func() { <-c.rateLimiter }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// 构建URL
	videoURL := c.buildVideoURL(videoID)

	// 重试机制
	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(c.config.RetryDelay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		info, err := c.fetchWithRetry(ctx, videoURL)
		if err == nil {
			return info, nil
		}
		lastErr = err

		if c.config.Debug {
			fmt.Printf("Attempt %d failed for %s: %v\n", attempt+1, videoID, err)
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.config.MaxRetries+1, lastErr)
}

// fetchWithRetry 执行实际的抓取
func (c *Client) fetchWithRetry(ctx context.Context, videoURL string) (*VideoInfo, error) {
	// 创建请求
	req, err := c.createRequest(ctx, videoURL)
	if err != nil {
		return nil, err
	}

	// 设置代理
	client := c.httpClient
	var proxy *Proxy
	if c.config.UseProxy && c.proxyPool != nil {
		proxy, err = c.proxyPool.GetNext()
		if err == nil && proxy != nil {
			proxyURL, _ := url.Parse(proxy.URL)
			client = &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				},
				Timeout: c.config.RequestTimeout,
			}
		}
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		if proxy != nil {
			proxy.MarkFailed()
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		if proxy != nil {
			proxy.MarkFailed()
		}
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	// 标记代理成功
	if proxy != nil {
		proxy.MarkSuccess()
	}

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 解析视频信息
	info, err := parseVideoInfo(string(bodyBytes), videoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse video info: %w", err)
	}

	info.CrawledAt = time.Now()
	return info, nil
}

// createRequest 创建HTTP请求
func (c *Client) createRequest(ctx context.Context, videoURL string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", videoURL, nil)
	if err != nil {
		return nil, err
	}

	// 设置User-Agent（随机选择）
	userAgent := c.getRandomUserAgent()
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Referer", "https://www.bilibili.com")
	req.Header.Set("Connection", "keep-alive")

	return req, nil
}

// getRandomUserAgent 随机获取一个User-Agent
func (c *Client) getRandomUserAgent() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.config.UserAgents) == 0 {
		return defaultUserAgents()[0]
	}

	return c.config.UserAgents[rand.Intn(len(c.config.UserAgents))]
}

// buildVideoURL 构建视频URL
func (c *Client) buildVideoURL(videoID string) string {
	// 支持多种输入格式
	if isURL(videoID) {
		return videoID
	}
	if isBVID(videoID) {
		return fmt.Sprintf("https://www.bilibili.com/video/%s", videoID)
	}
	if isAVID(videoID) {
		return fmt.Sprintf("https://www.bilibili.com/video/av%s", videoID)
	}
	// 默认当作BV号处理
	return fmt.Sprintf("https://www.bilibili.com/video/%s", videoID)
}

// Close 关闭客户端
func (c *Client) Close() {
	if c.proxyPool != nil {
		c.proxyPool.Stop()
	}
}
