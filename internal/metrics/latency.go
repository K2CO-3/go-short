// Package metrics 提供业务层延迟与命中率统计。
// 需在 handler 中调用 RecordXxx 埋点，用于观察本地缓存、Redis、PostgreSQL 各层效果。
// 与 pprof（零埋点、代码级 CPU/内存分析）互补。
package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// LatencyStats 延迟统计
type LatencyStats struct {
	// 本地缓存统计
	LocalCacheHits   uint64        // 命中次数
	LocalCacheMisses uint64        // 未命中次数
	LocalCacheTotal  uint64        // 总查询次数
	LocalCacheSum    int64         // 总耗时（纳秒）
	LocalCacheMax    int64         // 最大耗时（纳秒）

	// Redis 统计
	RedisHits   uint64
	RedisMisses uint64
	RedisTotal  uint64
	RedisSum    int64
	RedisMax    int64

	// PostgreSQL 统计
	PostgresHits   uint64
	PostgresMisses uint64
	PostgresTotal  uint64
	PostgresSum    int64
	PostgresMax    int64

	mu sync.RWMutex
}

var globalStats = &LatencyStats{}

// RecordLocalCache 记录本地缓存操作
func RecordLocalCache(hit bool, duration time.Duration) {
	stats := globalStats
	atomic.AddUint64(&stats.LocalCacheTotal, 1)
	durationNs := int64(duration.Nanoseconds())

	if hit {
		atomic.AddUint64(&stats.LocalCacheHits, 1)
	} else {
		atomic.AddUint64(&stats.LocalCacheMisses, 1)
	}

	atomic.AddInt64(&stats.LocalCacheSum, durationNs)

	// 更新最大值（使用 CAS 循环）
	for {
		currentMax := atomic.LoadInt64(&stats.LocalCacheMax)
		if durationNs <= currentMax {
			break
		}
		if atomic.CompareAndSwapInt64(&stats.LocalCacheMax, currentMax, durationNs) {
			break
		}
	}
}

// RecordRedis 记录 Redis 操作
func RecordRedis(hit bool, duration time.Duration) {
	stats := globalStats
	atomic.AddUint64(&stats.RedisTotal, 1)
	durationNs := int64(duration.Nanoseconds())

	if hit {
		atomic.AddUint64(&stats.RedisHits, 1)
	} else {
		atomic.AddUint64(&stats.RedisMisses, 1)
	}

	atomic.AddInt64(&stats.RedisSum, durationNs)

	for {
		currentMax := atomic.LoadInt64(&stats.RedisMax)
		if durationNs <= currentMax {
			break
		}
		if atomic.CompareAndSwapInt64(&stats.RedisMax, currentMax, durationNs) {
			break
		}
	}
}

// RecordPostgres 记录 PostgreSQL 操作
func RecordPostgres(hit bool, duration time.Duration) {
	stats := globalStats
	atomic.AddUint64(&stats.PostgresTotal, 1)
	durationNs := int64(duration.Nanoseconds())

	if hit {
		atomic.AddUint64(&stats.PostgresHits, 1)
	} else {
		atomic.AddUint64(&stats.PostgresMisses, 1)
	}

	atomic.AddInt64(&stats.PostgresSum, durationNs)

	for {
		currentMax := atomic.LoadInt64(&stats.PostgresMax)
		if durationNs <= currentMax {
			break
		}
		if atomic.CompareAndSwapInt64(&stats.PostgresMax, currentMax, durationNs) {
			break
		}
	}
}

// GetStats 获取统计快照
func GetStats() LatencyStats {
	stats := globalStats
	return LatencyStats{
		LocalCacheHits:   atomic.LoadUint64(&stats.LocalCacheHits),
		LocalCacheMisses: atomic.LoadUint64(&stats.LocalCacheMisses),
		LocalCacheTotal:  atomic.LoadUint64(&stats.LocalCacheTotal),
		LocalCacheSum:    atomic.LoadInt64(&stats.LocalCacheSum),
		LocalCacheMax:    atomic.LoadInt64(&stats.LocalCacheMax),

		RedisHits:   atomic.LoadUint64(&stats.RedisHits),
		RedisMisses: atomic.LoadUint64(&stats.RedisMisses),
		RedisTotal:  atomic.LoadUint64(&stats.RedisTotal),
		RedisSum:    atomic.LoadInt64(&stats.RedisSum),
		RedisMax:    atomic.LoadInt64(&stats.RedisMax),

		PostgresHits:   atomic.LoadUint64(&stats.PostgresHits),
		PostgresMisses: atomic.LoadUint64(&stats.PostgresMisses),
		PostgresTotal:  atomic.LoadUint64(&stats.PostgresTotal),
		PostgresSum:    atomic.LoadInt64(&stats.PostgresSum),
		PostgresMax:    atomic.LoadInt64(&stats.PostgresMax),
	}
}

// ResetStats 重置统计（用于测试）
func ResetStats() {
	stats := globalStats
	atomic.StoreUint64(&stats.LocalCacheHits, 0)
	atomic.StoreUint64(&stats.LocalCacheMisses, 0)
	atomic.StoreUint64(&stats.LocalCacheTotal, 0)
	atomic.StoreInt64(&stats.LocalCacheSum, 0)
	atomic.StoreInt64(&stats.LocalCacheMax, 0)

	atomic.StoreUint64(&stats.RedisHits, 0)
	atomic.StoreUint64(&stats.RedisMisses, 0)
	atomic.StoreUint64(&stats.RedisTotal, 0)
	atomic.StoreInt64(&stats.RedisSum, 0)
	atomic.StoreInt64(&stats.RedisMax, 0)

	atomic.StoreUint64(&stats.PostgresHits, 0)
	atomic.StoreUint64(&stats.PostgresMisses, 0)
	atomic.StoreUint64(&stats.PostgresTotal, 0)
	atomic.StoreInt64(&stats.PostgresSum, 0)
	atomic.StoreInt64(&stats.PostgresMax, 0)
}

// calculateAvg 计算平均延迟（纳秒转毫秒）
func calculateAvg(sum int64, count uint64) float64 {
	if count == 0 {
		return 0
	}
	return float64(sum) / float64(count) / 1e6 // 转换为毫秒
}

// calculateMax 计算最大延迟（纳秒转毫秒）
func calculateMax(maxNs int64) float64 {
	return float64(maxNs) / 1e6 // 转换为毫秒
}

// FormatStats 格式化统计信息为可读字符串
func FormatStats() map[string]interface{} {
	stats := GetStats()

	result := make(map[string]interface{})

	// 本地缓存统计
	if stats.LocalCacheTotal > 0 {
		result["local_cache"] = map[string]interface{}{
			"hits":        stats.LocalCacheHits,
			"misses":      stats.LocalCacheMisses,
			"total":       stats.LocalCacheTotal,
			"hit_rate":    float64(stats.LocalCacheHits) / float64(stats.LocalCacheTotal) * 100,
			"avg_latency_ms": calculateAvg(stats.LocalCacheSum, stats.LocalCacheTotal),
			"max_latency_ms": calculateMax(stats.LocalCacheMax),
		}
	}

	// Redis 统计
	if stats.RedisTotal > 0 {
		result["redis"] = map[string]interface{}{
			"hits":        stats.RedisHits,
			"misses":      stats.RedisMisses,
			"total":       stats.RedisTotal,
			"hit_rate":    float64(stats.RedisHits) / float64(stats.RedisTotal) * 100,
			"avg_latency_ms": calculateAvg(stats.RedisSum, stats.RedisTotal),
			"max_latency_ms": calculateMax(stats.RedisMax),
		}
	}

	// PostgreSQL 统计
	if stats.PostgresTotal > 0 {
		result["postgres"] = map[string]interface{}{
			"hits":        stats.PostgresHits,
			"misses":      stats.PostgresMisses,
			"total":       stats.PostgresTotal,
			"hit_rate":    float64(stats.PostgresHits) / float64(stats.PostgresTotal) * 100,
			"avg_latency_ms": calculateAvg(stats.PostgresSum, stats.PostgresTotal),
			"max_latency_ms": calculateMax(stats.PostgresMax),
		}
	}

	return result
}
