package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"speakmate/internal/config"
)

// 基础设施层复用的哨兵错误。
var (
	// ErrRedisUnavailable 表示启用 Redis 时无法建立可用连接。
	ErrRedisUnavailable = errors.New("Redis 不可用")
)

// OpenClient 初始化 Redis client，并在启用 Redis 时执行 ping 健康检查。
func OpenClient(ctx context.Context, cfg config.RedisConfig) (*goredis.Client, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	timeout := time.Duration(cfg.ConnectTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	client := goredis.NewClient(&goredis.Options{
		Addr:            cfg.Addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		DialTimeout:     timeout,
		ReadTimeout:     timeout,
		WriteTimeout:    timeout,
		MaxRetries:      0,
		MinRetryBackoff: 0,
		MaxRetryBackoff: 0,
	})
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("%w：ping %s 失败：%v", ErrRedisUnavailable, cfg.Addr, err)
	}

	return client, nil
}
