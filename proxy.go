package bilibili_crawler

import (
	"errors"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// ProxyPool 代理池管理器
type ProxyPool struct {
	proxies       []*Proxy
	current       uint32
	mu            sync.RWMutex
	checkInterval time.Duration
	stopChan      chan struct{}
}

// Proxy 代理信息
type Proxy struct {
	URL          string
	Available    bool
	LastChecked  time.Time
	FailCount    int
	SuccessCount int
	mu           sync.Mutex
}

// NewProxyPool 创建新的代理池
func NewProxyPool(proxyURLs []string, checkInterval time.Duration) *ProxyPool {
	pool := &ProxyPool{
		proxies:       make([]*Proxy, 0, len(proxyURLs)),
		checkInterval: checkInterval,
		stopChan:      make(chan struct{}),
	}

	for _, proxyURL := range proxyURLs {
		pool.proxies = append(pool.proxies, &Proxy{
			URL:       proxyURL,
			Available: true,
		})
	}

	// 启动代理健康检查
	if checkInterval > 0 {
		go pool.healthCheck()
	}

	return pool
}

// GetNext 获取下一个可用的代理（负载均衡）
func (p *ProxyPool) GetNext() (*Proxy, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.proxies) == 0 {
		return nil, errors.New("no proxies available")
	}

	// 轮询算法
	startIdx := atomic.AddUint32(&p.current, 1)
	for i := 0; i < len(p.proxies); i++ {
		idx := (startIdx + uint32(i)) % uint32(len(p.proxies))
		proxy := p.proxies[idx]

		proxy.mu.Lock()
		available := proxy.Available
		proxy.mu.Unlock()

		if available {
			return proxy, nil
		}
	}

	return nil, errors.New("all proxies are unavailable")
}

// MarkSuccess 标记代理成功
func (proxy *Proxy) MarkSuccess() {
	proxy.mu.Lock()
	defer proxy.mu.Unlock()
	proxy.SuccessCount++
	proxy.FailCount = 0
	proxy.Available = true
}

// MarkFailed 标记代理失败
func (proxy *Proxy) MarkFailed() {
	proxy.mu.Lock()
	defer proxy.mu.Unlock()
	proxy.FailCount++
	// 失败3次后标记为不可用
	if proxy.FailCount >= 3 {
		proxy.Available = false
	}
}

// healthCheck 定期检查代理健康状态
func (p *ProxyPool) healthCheck() {
	ticker := time.NewTicker(p.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.checkAllProxies()
		case <-p.stopChan:
			return
		}
	}
}

// checkAllProxies 检查所有代理
func (p *ProxyPool) checkAllProxies() {
	p.mu.RLock()
	proxies := make([]*Proxy, len(p.proxies))
	copy(proxies, p.proxies)
	p.mu.RUnlock()

	var wg sync.WaitGroup
	for _, proxy := range proxies {
		wg.Add(1)
		go func(px *Proxy) {
			defer wg.Done()
			p.checkProxy(px)
		}(proxy)
	}
	wg.Wait()
}

// checkProxy 检查单个代理
func (p *ProxyPool) checkProxy(proxy *Proxy) {
	proxyURL, err := url.Parse(proxy.URL)
	if err != nil {
		proxy.MarkFailed()
		return
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 10 * time.Second,
	}

	// 测试代理是否可用
	resp, err := client.Get("https://www.bilibili.com/")
	if err != nil {
		proxy.MarkFailed()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		proxy.MarkSuccess()
		proxy.mu.Lock()
		proxy.LastChecked = time.Now()
		proxy.mu.Unlock()
	} else {
		proxy.MarkFailed()
	}
}

// Stop 停止代理池
func (p *ProxyPool) Stop() {
	close(p.stopChan)
}

// GetStats 获取代理池统计信息
func (p *ProxyPool) GetStats() map[string]any {
	p.mu.RLock()
	defer p.mu.RUnlock()

	available := 0
	totalSuccess := 0
	totalFail := 0

	for _, proxy := range p.proxies {
		proxy.mu.Lock()
		if proxy.Available {
			available++
		}
		totalSuccess += proxy.SuccessCount
		totalFail += proxy.FailCount
		proxy.mu.Unlock()
	}

	return map[string]any{
		"total":         len(p.proxies),
		"available":     available,
		"total_success": totalSuccess,
		"total_fail":    totalFail,
	}
}
