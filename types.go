package bilibili_crawler

import (
	"time"
)

// VideoInfo 存储视频的所有信息
type VideoInfo struct {
	Title       string    `json:"title"`        // 视频标题
	URL         string    `json:"url"`          // 视频链接
	Author      string    `json:"author"`       // UP主名称
	AuthorID    string    `json:"author_id"`    // UP主ID
	Views       int       `json:"views"`        // 播放量
	Danmaku     int       `json:"danmaku"`      // 弹幕数
	Likes       int       `json:"likes"`        // 点赞数
	Coins       int       `json:"coins"`        // 投币数
	Favorites   int       `json:"favorites"`    // 收藏数
	Shares      int       `json:"shares"`       // 转发数
	PublishDate string    `json:"publish_date"` // 发布时间
	Duration    int       `json:"duration"`     // 视频时长（秒）
	VideoDesc   string    `json:"video_desc"`   // 视频简介
	AuthorDesc  string    `json:"author_desc"`  // 作者简介
	Tags        string    `json:"tags"`         // 标签
	Aid         string    `json:"aid"`          // 视频AID
	BV          string    `json:"bv"`           // BV号
	CrawledAt   time.Time `json:"crawled_at"`   // 爬取时间
}

// CrawlResult 爬取结果
type CrawlResult struct {
	VideoID string     `json:"video_id"`
	Info    *VideoInfo `json:"info"`
	Error   error      `json:"error,omitempty"`
}

// Config 配置选项
type Config struct {
	// 并发控制
	MaxWorkers     int           `json:"max_workers"`     // 最大并发数
	RequestDelay   time.Duration `json:"request_delay"`   // 请求间隔
	RequestTimeout time.Duration `json:"request_timeout"` // 请求超时

	// 代理配置
	ProxyURLs   []string `json:"proxy_urls"`   // 代理列表
	UseProxy    bool     `json:"use_proxy"`    // 是否使用代理
	ProxyRotate bool     `json:"proxy_rotate"` // 是否轮换代理

	// 重试配置
	MaxRetries int           `json:"max_retries"` // 最大重试次数
	RetryDelay time.Duration `json:"retry_delay"` // 重试间隔

	// 用户代理池
	UserAgents []string `json:"user_agents"` // User-Agent列表

	// 调试模式
	Debug bool `json:"debug"` // 调试模式
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		MaxWorkers:     5,
		RequestDelay:   1 * time.Second,
		RequestTimeout: 30 * time.Second,
		UseProxy:       false,
		ProxyRotate:    true,
		MaxRetries:     3,
		RetryDelay:     2 * time.Second,
		UserAgents:     defaultUserAgents(),
		Debug:          false,
	}
}

// defaultUserAgents 返回默认的User-Agent列表
func defaultUserAgents() []string {
	return []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:120.0) Gecko/20100101 Firefox/120.0",
	}
}
