package infra

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/newmo-oss/ergo"
	"github.com/redis/go-redis/v9"
)

// NewRedisClient はRedisクライアントを初期化します
// 本番では環境変数からADDR等を取得するように変更してください
func NewRedisClient(addr string, password string, db int) (*redis.Client, error) {
	poolSize := 10
	if env := os.Getenv("REDIS_POOL_SIZE"); env != "" {
		if v, err := strconv.Atoi(env); err == nil && v > 0 {
			poolSize = v
		}
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     poolSize,
	})

	// 接続テスト
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, ergo.New("failed to connect to redis", slog.String("error", err.Error()))
	}

	return rdb, nil
}
