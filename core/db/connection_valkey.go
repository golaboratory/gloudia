package db

import (
	"context"
	"fmt"
	"time"

	"github.com/golaboratory/gloudia/core/config"
	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeycompat"
)

// ValkeyClient はValkeyサーバと通信するためのクライアントインターフェースをラップした構造体です。
//   - Client: valkey-goのクライアントインスタンス
type ValkeyClient struct {
	Client valkey.Client
}

// NewValkeyClient はValkeyの設定情報(ValkeyConfig)を元に新しいValkeyClientインスタンスを生成します。
// 戻り値:
//   - *ValkeyClient: 生成されたクライアントインスタンス
//   - error: インスタンス生成中に発生したエラー
func NewValkeyClient() (*ValkeyClient, error) {
	cfg, err := config.New[ValkeyConfig]()
	if err != nil {
		return nil, fmt.Errorf("failed to get valkey config: %w", err)
	}

	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create valkey client: %w", err)
	}

	return &ValkeyClient{Client: client}, nil
}

// Get は指定されたキーの値をValkeyサーバから取得します。
// 引数:
//   - ctx: コンテキスト
//   - key: 取得対象のキー
//
// 戻り値:
//   - string: 取得された値
//   - error: 処理中に発生したエラー
func (v *ValkeyClient) Get(ctx context.Context, key string) (string, error) {
	compat := valkeycompat.NewAdapter(v.Client)
	return compat.Get(ctx, key).Result()
}

// Set は指定されたキーに対して値を設定し、オプションで有効期限を指定します。
// 引数:
//   - ctx: コンテキスト
//   - key: 設定対象のキー
//   - value: 設定する値
//   - expiration: 値の有効期間
//
// 戻り値:
//   - bool: 設定が成功した場合はtrue
//   - error: 設定中に発生したエラー
func (v *ValkeyClient) Set(ctx context.Context, key, value string, expiration time.Duration) (bool, error) {
	compat := valkeycompat.NewAdapter(v.Client)
	_, err := compat.Set(ctx, key, value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set value: %w", err)
	}
	return true, nil
}

// Delete は指定されたキーの値をValkeyサーバから削除します。
// 引数:
//   - ctx: コンテキスト
//   - key: 削除対象のキー
//
// 戻り値:
//   - bool: 削除成功時はtrue
//   - error: 削除中に発生したエラー
func (v *ValkeyClient) Delete(ctx context.Context, key string) (bool, error) {
	compat := valkeycompat.NewAdapter(v.Client)
	_, err := compat.Del(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to delete value: %w", err)
	}
	return true, nil
}

// Close はValkeyClientの内部クライアント接続をクローズします。
func (v *ValkeyClient) Close() {
	v.Client.Close()
}
