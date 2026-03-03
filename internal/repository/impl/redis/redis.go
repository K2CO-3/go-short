package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisRepoImpl struct {
	rdb *redis.Client
}

// NewRedisClient 初始化 Redis 连接
func NewRedisClient() (*redis.Client, error) {
	// 1. 读取配置
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	password := os.Getenv("REDIS_PASSWORD")

	// 2. 创建客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,  // 使用默认 DB 0
		PoolSize:     20, // 连接池大小
		MinIdleConns: 5,  // 最小空闲连接
	})

	// 3. Ping 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	log.Println("✅ Redis connected successfully")
	return rdb, nil
}

// NewRedisRepository 创建 RedisRepository 实例
func NewRedisRepository(rdb *redis.Client) *redisRepoImpl {
	return &redisRepoImpl{rdb: rdb}
}

// ==========================================
// Redis 缓存辅助操作
// ==========================================

// CacheLink 将长链接写入 Redis 缓存
func (d *redisRepoImpl) CacheLink(ctx context.Context, code string, originalURL string, duration time.Duration) error {
	return d.rdb.Set(ctx, "short:"+code, originalURL, duration).Err()
}

// GetLinkFromCache 从 Redis 获取长链接
func (d *redisRepoImpl) GetLinkFromCache(ctx context.Context, code string) (string, error) {
	return d.rdb.Get(ctx, "short:"+code).Result()
}

// DeleteLinkCache 删除 Redis 中的链接缓存
func (d *redisRepoImpl) DeleteLinkCache(ctx context.Context, code string) error {
	return d.rdb.Del(ctx, "short:"+code).Err()
}

// CacheInvalidateChannel Redirect 订阅此 channel 以删除本地缓存
const CacheInvalidateChannel = "cache_invalidate"

// PublishCacheInvalidate 发布缓存失效消息，Redirect 服务订阅后删除本地缓存
func (d *redisRepoImpl) PublishCacheInvalidate(ctx context.Context, code string) error {
	return d.rdb.Publish(ctx, CacheInvalidateChannel, code).Err()
}

// doInvalidateLink 立即执行：删除 Redis 缓存并发布失效消息（供 worker 内部调用）
func (d *redisRepoImpl) doInvalidateLink(ctx context.Context, code string) error {
	if err := d.DeleteLinkCache(ctx, code); err != nil {
		return err
	}
	_ = d.PublishCacheInvalidate(ctx, code) // 本地缓存失效，失败不影响主流程
	return nil
}

// InvalidateLink 实现 CacheInvalidator：将任务入队，由 worker 异步执行（提高可靠性）
func (d *redisRepoImpl) InvalidateLink(ctx context.Context, code string) error {
	return d.EnqueueCacheInvalidate(ctx, code)
}

// ==========================================
// 延迟队列（增加缓存删除可靠性，支持失败重试）
// ==========================================

// 访问日志 Stream（取代原 List access_logs，支持 at-least-once 投递）
// 使用新 key 避免与旧 List 共存时 WRONGTYPE
const (
	AccessLogStream        = "access_logs_stream"
	AccessLogConsumerGroup = "access_logs_group"

	CacheInvalidateDelayQueue = "cache_invalidate_delay_queue"
	CacheInvalidateDelaySec   = 1 // 延迟 1 秒执行，确保 DB 事务已提交
	CacheInvalidateRetrySec   = 5 // 失败后 5 秒重试
	LockKeyPrefix             = "lock:short:"
	LockTTLSeconds            = 10
)

// EnqueueCacheInvalidate 将缓存失效任务推入延迟队列，由 worker 异步处理
func (d *redisRepoImpl) EnqueueCacheInvalidate(ctx context.Context, code string) error {
	executeAt := float64(time.Now().Unix()) + CacheInvalidateDelaySec
	return d.rdb.ZAdd(ctx, CacheInvalidateDelayQueue, redis.Z{Score: executeAt, Member: code}).Err()
}

// RunCacheInvalidateWorker 消费延迟队列，执行缓存失效（可阻塞运行）
func (d *redisRepoImpl) RunCacheInvalidateWorker(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now().Unix()
			maxScore := fmt.Sprintf("%d", now)
			codes, err := d.rdb.ZRangeByScore(ctx, CacheInvalidateDelayQueue, &redis.ZRangeBy{
				Min: "-inf", Max: maxScore, Count: 50,
			}).Result()
			if err != nil || len(codes) == 0 {
				continue
			}
			for _, code := range codes {
				if err := d.doInvalidateLink(ctx, code); err != nil {
					// 失败：重新入队延迟重试（ZAdd 更新同 member 的 score）
					retryAt := float64(time.Now().Unix() + CacheInvalidateRetrySec)
					_ = d.rdb.ZAdd(ctx, CacheInvalidateDelayQueue, redis.Z{Score: retryAt, Member: code}).Err()
				} else {
					_ = d.rdb.ZRem(ctx, CacheInvalidateDelayQueue, code).Err()
				}
			}
		}
	}
}

// ==========================================
// 分布式锁（防缓存击穿：热点 key 过期时，仅一个请求回源 DB）
// ==========================================

// tryLockToken 生成随机 token，用于安全释放锁
func tryLockToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// TryLock 尝试获取分布式锁，key 通常为 lock:short:{code}
// 返回 (token, true) 表示获取成功；token 用于 Unlock
func (d *redisRepoImpl) TryLock(ctx context.Context, key string, ttl time.Duration) (string, bool) {
	token, err := tryLockToken()
	if err != nil {
		return "", false
	}
	ok, err := d.rdb.SetNX(ctx, key, token, ttl).Result()
	if err != nil || !ok {
		return "", false
	}
	return token, true
}

// Unlock 释放锁，仅当 token 匹配时删除（防止误删他人锁）
var unlockScript = redis.NewScript(`
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
`)

func (d *redisRepoImpl) Unlock(ctx context.Context, key string, token string) error {
	return unlockScript.Run(ctx, d.rdb, []string{key}, token).Err()
}

// LockKeyForCode 返回短码对应的锁 key
func LockKeyForCode(code string) string {
	return LockKeyPrefix + code
}
