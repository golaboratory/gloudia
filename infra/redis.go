package infra

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/newmo-oss/ergo"
	"github.com/redis/go-redis/v9"
)

// センチネルエラー定義
var (
	// ErrRedisConnectionFailed はRedisへの接続に失敗した場合のエラー
	ErrRedisConnectionFailed = ergo.NewSentinel("failed to connect to redis")
)

// NewRedisClient はRedisクライアントを初期化します。
// プールサイズは環境変数 REDIS_POOL_SIZE から取得します（既定値 10）。
// DialTimeout 10秒 / ReadTimeout 30秒 / WriteTimeout 30秒 は固定です。
// 生成時に 5 秒タイムアウトの Ping による接続確認（ネットワークI/O）をブロッキングで実行し、
// 失敗した場合は ErrRedisConnectionFailed をラップしたエラーを返します。
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
		return nil, ergo.Wrap(ErrRedisConnectionFailed, err.Error())
	}

	return rdb, nil
}
