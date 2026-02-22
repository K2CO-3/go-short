package redis

import (
	"context"
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
