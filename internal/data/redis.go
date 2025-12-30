package data

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ydcloud-dy/opshub/internal/conf"
)

// Redis Redis客户端
type Redis struct {
	client *redis.Client
}

// NewRedis 创建 Redis 客户端
func NewRedis(cfg *conf.Config) (*Redis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.GetRedisAddr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConn,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis连接失败: %w", err)
	}

	return &Redis{
		client: rdb,
	}, nil
}

// Get 获取客户端
func (r *Redis) Get() *redis.Client {
	return r.client
}

// Close 关闭连接
func (r *Redis) Close() error {
	return r.client.Close()
}
