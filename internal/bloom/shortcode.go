package bloom

import (
	"context"
	"sync"

	"github.com/bits-and-blooms/bloom/v3"
	"gorm.io/gorm"
)

// ShortCodeBloom 防 Redis 缓存穿透的布隆过滤器
// 仅用于 redirect 判定「一定不存在」的短码，直接 404，避免穿透到 Redis/DB
type ShortCodeBloom struct {
	mu     sync.RWMutex
	filter *bloom.BloomFilter
	ready  bool
}

// NewShortCodeBloom 创建，expectedItems 如 100 万，fpRate 如 0.01
func NewShortCodeBloom(expectedItems uint, fpRate float64) *ShortCodeBloom {
	return &ShortCodeBloom{filter: bloom.NewWithEstimates(expectedItems, fpRate)}
}

// Add 将短码加入
func (b *ShortCodeBloom) Add(code string) {
	if code == "" {
		return
	}
	b.mu.Lock()
	b.filter.AddString(code)
	b.mu.Unlock()
}

// DefinitelyNotExist 若为 true 则短码一定不存在，可直返 404
func (b *ShortCodeBloom) DefinitelyNotExist(code string) bool {
	b.mu.RLock()
	ready := b.ready
	notIn := !b.filter.TestString(code)
	b.mu.RUnlock()
	return ready && notIn
}

// LoadFromDB 从 DB 加载所有有效短码，加载完成后 SetReady
func (b *ShortCodeBloom) LoadFromDB(ctx context.Context, db *gorm.DB) (int, error) {
	var codes []string
	err := db.WithContext(ctx).Table("links").
		Select("short_code").
		Where("status = ?", true).
		Where("expires_at IS NULL OR expires_at > NOW()").
		Pluck("short_code", &codes).Error
	if err != nil {
		return 0, err
	}
	b.mu.Lock()
	for _, c := range codes {
		if c != "" {
			b.filter.AddString(c)
		}
	}
	b.ready = true
	b.mu.Unlock()
	return len(codes), nil
}
