# Bilibili 视频信息爬虫 (bilibili-crawler)

[![Go Report Card](https://goreportcard.com/badge/github.com/iwen-conf/bilibili-crawler)](https://goreportcard.com/report/github.com/iwen-conf/bilibili-crawler)
[![Go Version](https://img.shields.io/github/go-mod/go-version/iwen-conf/bilibili-crawler)](https://golang.org)
[![License](https://img.shields.io/github/license/iwen-conf/bilibili-crawler)](https://github.com/iwen-conf/bilibili-crawler/blob/master/LICENSE)

一个简单、高效且可配置的 Go 语言库，用于从 Bilibili (B站) 抓取单个视频的详细信息。

## ✨ 功能特性

- **全面的数据提取**: 获取视频的标题、作者、播放量、弹幕、点赞、投币、收藏、分享、发布日期、时长、简介和标签等。
- **多种ID支持**: 支持通过 BV 号、AV 号或完整的视频 URL 来抓取信息。
- **高度可配置**: 支持自定义并发数、请求超时、重试次数和请求延迟。
- **代理支持**: 内置代理池，支持使用 HTTP/HTTPS 代理，并可自动轮换和管理代理状态。
- **健壮性设计**: 包含随机 User-Agent 池、请求重试和错误处理机制，提高抓取成功率。
- **批量处理**: 提供了批量抓取多个视频的便捷方法。

## 🚀 安装

```bash
go get -u github.com/iwen-conf/bilibili-crawler
```

## 💡 快速上手

以下是一个简单的例子，展示如何获取一个视频的信息。

```go
package main

import (
	"fmt"
	bilibili "github.com/iwen-conf/bilibili-crawler"
)

func main() {
	// 创建一个默认配置的客户端
	client := bilibili.NewClient(nil)
	defer client.Close()

	// 视频ID可以是 BV号, AV号, 或者完整的URL
	videoID := "BV1GJ411x7h7"

	// 抓取视频信息
	info, err := client.FetchVideo(videoID)
	if err != nil {
		fmt.Println("抓取失败:", err)
		return
	}

	// 打印视频信息
	fmt.Printf("标题: %s\n", info.Title)
	fmt.Printf("UP主: %s (ID: %s)\n", info.Author, info.AuthorID)
	fmt.Printf("播放量: %d\n", info.Views)
	fmt.Printf("弹幕数: %d\n", info.Danmaku)
	fmt.Printf("发布日期: %s\n", info.PublishDate)
	fmt.Printf("视频链接: %s\n", info.URL)
}
```

## 📊 返回数据结构

成功抓取后，`FetchVideo` 函数会返回一个 `VideoInfo` 结构体，包含以下字段：

```go
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
```

## ⚙️ 高级配置

你可以通过创建一个 `Config` 结构体来自定义爬虫的行为。

```go
config := &bilibili.Config{
    MaxWorkers:     10,                    // 最大并发数为10
    RequestTimeout: 30 * time.Second,      // 请求超时30秒
    MaxRetries:     5,                     // 最大重试5次
    UseProxy:       true,                  // 启用代理
    ProxyURLs:      []string{"http://user:pass@host:port"}, // 代理列表
    UserAgents:     []string{"MyCustomUserAgent/1.0"}, // 自定义User-Agent
    Debug:          true,                  // 开启调试模式
}

client := bilibili.NewClient(config)
// ...
```

### 配置选项说明

- `MaxWorkers`: 最大并发请求数。
- `RequestDelay`: 每个请求之间的延迟（当前版本通过限流实现，此参数保留）。
- `RequestTimeout`: HTTP 请求的超时时间。
- `ProxyURLs`: 代理服务器地址列表。
- `UseProxy`: 是否启用代理。
- `MaxRetries`: 单个请求失败后的最大重试次数。
- `RetryDelay`: 每次重试之间的等待时间。
- `UserAgents`: 自定义 User-Agent 池，爬虫会从中随机选择。
- `Debug`: 是否打印详细的调试信息。

## 📄 许可证

本项目基于 [MIT License](LICENSE) 开源。
