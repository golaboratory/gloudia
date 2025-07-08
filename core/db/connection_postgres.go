package db

import (
	"context"
	"fmt"

	"github.com/golaboratory/gloudia/core/config"
	"github.com/jackc/pgx/v5"
)

// NewPostgresConnection はPostgreSQLデータベースへの新しい接続を確立します。
// 戻り値:
//   - *pgx.Conn: データベース接続オブジェクト
//   - error: 接続時に発生したエラー。正常に接続できた場合はnilを返します。
func NewPostgresConnection() (*pgx.Conn, error) {
	dbConfig, err := config.New[DBConfig]()
	if err != nil {
		return nil, err
	}

	databaseUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Database)

	conn, err := pgx.Connect(
		context.Background(),
		databaseUrl)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
