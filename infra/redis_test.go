package infra

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisClient(t *testing.T) {
	// miniredisのセットアップ
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	t.Run("正常系: 接続成功", func(t *testing.T) {
		// miniredisのアドレスを使用して接続
		client, err := NewRedisClient(mr.Addr(), "", 0)

		assert.NoError(t, err)
		assert.NotNil(t, client)

		if client != nil {
			defer client.Close()
			// Pingが通るか確認
			pong, err := client.Ping(context.Background()).Result()
			assert.NoError(t, err)
			assert.Equal(t, "PONG", pong)
		}
	})

	t.Run("異常系: 接続失敗（無効なアドレス）", func(t *testing.T) {
		// 存在しないポートを指定
		// 注: ローカル環境によっては偶然ポートが開いている可能性もゼロではないが、
		// 通常は接続拒否されるはず
		client, err := NewRedisClient("localhost:99999", "", 0)

		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to connect to redis")
	})
}
