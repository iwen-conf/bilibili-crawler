package bilibili_crawler

import (
	"context"
	"testing"
	"time"
)

func TestClient_FetchVideo(t *testing.T) {
	client := NewClient(nil)
	defer client.Close()

	// 测试获取视频信息
	info, err := client.FetchVideo("BV1Ks1UB8Ek4")
	if err != nil {
		t.Fatalf("获取视频失败: %v", err)
	}

	// 验证基本字段
	if info.BV != "BV1Ks1UB8Ek4" {
		t.Errorf("BV号不匹配: got %s, want BV1Ks1UB8Ek4", info.BV)
	}

	if info.Title == "" {
		t.Error("标题为空")
	}

	if info.Author == "" {
		t.Error("作者为空")
	}

	t.Logf("成功获取视频: %s - %s", info.Title, info.Author)
}

func TestClient_FetchBatch(t *testing.T) {
	config := &Config{
		MaxWorkers:   3,
		RequestDelay: 500 * time.Millisecond,
	}

	client := NewClient(config)
	defer client.Close()

	videoIDs := []string{
		"BV1Ks1UB8Ek4",
		"BV1234567890", // 可能不存在的视频
	}

	results := client.FetchBatch(videoIDs, nil)

	if len(results) != len(videoIDs) {
		t.Errorf("结果数量不匹配: got %d, want %d", len(results), len(videoIDs))
	}

	successCount := 0
	for _, result := range results {
		if result.Error == nil && result.Info != nil {
			successCount++
			t.Logf("成功获取: %s", result.Info.Title)
		} else if result.Error != nil {
			t.Logf("失败: %s - %v", result.VideoID, result.Error)
		}
	}

	if successCount == 0 {
		t.Error("没有成功获取任何视频")
	}
}

func TestClient_Context(t *testing.T) {
	client := NewClient(nil)
	defer client.Close()

	// 创建一个会超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 故意延迟以触发超时
	time.Sleep(200 * time.Millisecond)

	_, err := client.FetchVideoWithContext(ctx, "BV1Ks1UB8Ek4")
	if err == nil {
		t.Error("预期超时错误，但没有发生")
	}
}

func TestProxyPool(t *testing.T) {
	proxyURLs := []string{
		"http://proxy1.example.com:8080",
		"http://proxy2.example.com:8080",
		"http://proxy3.example.com:8080",
	}

	pool := NewProxyPool(proxyURLs, 0) // 禁用健康检查
	defer pool.Stop()

	// 测试轮询
	seen := make(map[string]int)
	for i := 0; i < 9; i++ {
		proxy, err := pool.GetNext()
		if err != nil {
			t.Errorf("获取代理失败: %v", err)
			continue
		}
		seen[proxy.URL]++
	}

	// 每个代理应该被选中3次
	for url, count := range seen {
		if count != 3 {
			t.Errorf("代理 %s 被选中 %d 次，期望 3 次", url, count)
		}
	}
}

func TestFormatters(t *testing.T) {
	// 测试时间格式化
	duration := FormatDuration(3661)
	expected := "1:01:01"
	if duration != expected {
		t.Errorf("时间格式化错误: got %s, want %s", duration, expected)
	}

	// 测试数字格式化
	number := FormatNumber(12345)
	expected = "1.2万"
	if number != expected {
		t.Errorf("数字格式化错误: got %s, want %s", number, expected)
	}
}

// 基准测试
func BenchmarkFetchVideo(b *testing.B) {
	client := NewClient(nil)
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.FetchVideo("BV1Ks1UB8Ek4")
	}
}

func BenchmarkParseVideoInfo(b *testing.B) {
	// 使用一个简单的HTML样本
	html := `<title>测试视频_哔哩哔哩_bilibili</title>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseVideoInfo(html, "https://www.bilibili.com/video/BV1Ks1UB8Ek4")
	}
}
