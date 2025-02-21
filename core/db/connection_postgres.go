package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func NewConnection() (*pgx.Conn, error) {
	config := DBConfig{}
	if err := config.Load(); err != nil {
		return nil, err
	}

	databaseUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database)

	conn, err := pgx.Connect(context.Background(), databaseUrl)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
