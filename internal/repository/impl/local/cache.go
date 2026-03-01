package local

import (
	"sync"
	"time"
)

// cacheItem 缓存项，包含值和过期时间
type cacheItem struct {
	value      string
	expiresAt  time.Time
}

// LocalCache 本地内存缓存，线程安全
type LocalCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
	// 默认过期时间
	defaultTTL time.Duration
	// 清理过期项的间隔
	cleanupInterval time.Duration
	stopCleanup     chan bool
}

// NewLocalCache 创建本地缓存实例
// defaultTTL: 默认过期时间，例如 5 * time.Minute
// cleanupInterval: 清理过期项的间隔，例如 10 * time.Minute，设置为0则禁用自动清理
func NewLocalCache(defaultTTL time.Duration, cleanupInterval time.Duration) *LocalCache {
	cache := &LocalCache{
		items:           make(map[string]*cacheItem),
		defaultTTL:      defaultTTL,
		cleanupInterval: cleanupInterval,
		stopCleanup:     make(chan bool),
	}

	// 如果设置了清理间隔，启动后台清理协程
	if cleanupInterval > 0 {
		go cache.startCleanup()
	}

	return cache
}

// Get 获取缓存值，如果不存在或已过期则返回 false
func (c *LocalCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return "", false
	}

	// 检查是否过期
	if time.Now().After(item.expiresAt) {
		// 懒删除：过期项在读取时删除
		delete(c.items, key)
		return "", false
	}

	return item.value, true
}

// Set 设置缓存值，使用默认TTL
func (c *LocalCache) Set(key string, value string) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL 设置缓存值，指定TTL
func (c *LocalCache) SetWithTTL(key string, value string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

// Delete 删除指定key
func (c *LocalCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear 清空所有缓存
func (c *LocalCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*cacheItem)
}

// Size 返回当前缓存项数量
func (c *LocalCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// startCleanup 启动后台清理协程，定期删除过期项
func (c *LocalCache) startCleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanup 清理过期项
func (c *LocalCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.expiresAt) {
			delete(c.items, key)
		}
	}
}

// StopCleanup 停止后台清理协程
func (c *LocalCache) StopCleanup() {
	close(c.stopCleanup)
}
