package local

import (
	"container/list"
	"sync"
	"time"
)

// cacheItem 缓存项，包含值和过期时间
type cacheItem struct {
	value      string
	expiresAt  time.Time
}

// lruEntry list 节点，用于 LRU 顺序
type lruEntry struct {
	key string
	*cacheItem
}

// LocalCache 本地内存缓存，线程安全，支持 TTL + 简单 LRU 淘汰
type LocalCache struct {
	mu    sync.RWMutex
	items map[string]*list.Element
	lru   *list.List // Front=MRU, Back=LRU
	// 默认过期时间
	defaultTTL time.Duration
	// 清理过期项的间隔
	cleanupInterval time.Duration
	stopCleanup     chan bool
	// 最大容量，0 表示不限制
	maxItems int
}

// NewLocalCache 创建本地缓存实例
// defaultTTL: 默认过期时间，例如 5 * time.Minute
// cleanupInterval: 清理过期项的间隔，0 则禁用自动清理
// maxItems: 最大缓存数量，0 表示不限制；>0 时超出将 LRU 淘汰
func NewLocalCache(defaultTTL time.Duration, cleanupInterval time.Duration, maxItems int) *LocalCache {
	cache := &LocalCache{
		items:           make(map[string]*list.Element),
		lru:             list.New(),
		defaultTTL:      defaultTTL,
		cleanupInterval: cleanupInterval,
		stopCleanup:     make(chan bool),
		maxItems:        maxItems,
	}

	if cleanupInterval > 0 {
		go cache.startCleanup()
	}

	return cache
}

// Get 获取缓存值，如果不存在或已过期则返回 false；命中时移到 MRU
func (c *LocalCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.items[key]
	if !exists {
		return "", false
	}

	entry := elem.Value.(*lruEntry)
	if time.Now().After(entry.expiresAt) {
		c.removeElement(elem)
		return "", false
	}

	c.lru.MoveToFront(elem)
	return entry.value, true
}

// Set 设置缓存值，使用默认TTL
func (c *LocalCache) Set(key string, value string) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL 设置缓存值，指定TTL；超出容量时淘汰 LRU
func (c *LocalCache) SetWithTTL(key string, value string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := &cacheItem{value: value, expiresAt: time.Now().Add(ttl)}
	if elem, exists := c.items[key]; exists {
		entry := elem.Value.(*lruEntry)
		entry.cacheItem = item
		c.lru.MoveToFront(elem)
		return
	}

	// 超出容量则淘汰最久未用
	for c.maxItems > 0 && c.lru.Len() >= c.maxItems {
		oldest := c.lru.Back()
		if oldest != nil {
			c.removeElement(oldest)
		}
	}

	entry := &lruEntry{key: key, cacheItem: item}
	elem := c.lru.PushFront(entry)
	c.items[key] = elem
}

func (c *LocalCache) removeElement(elem *list.Element) {
	c.lru.Remove(elem)
	entry := elem.Value.(*lruEntry)
	delete(c.items, entry.key)
}

// Delete 删除指定key
func (c *LocalCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.removeElement(elem)
	}
}

// Clear 清空所有缓存
func (c *LocalCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*list.Element)
	c.lru.Init()
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
	var toRemove []*list.Element
	for _, elem := range c.items {
		entry := elem.Value.(*lruEntry)
		if now.After(entry.expiresAt) {
			toRemove = append(toRemove, elem)
		}
	}
	for _, elem := range toRemove {
		c.removeElement(elem)
	}
}

// StopCleanup 停止后台清理协程
func (c *LocalCache) StopCleanup() {
	close(c.stopCleanup)
}
