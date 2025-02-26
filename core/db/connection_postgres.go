package db

import (
	"context"
	"fmt"

	"github.com/golaboratory/gloudia/core/config"
	"github.com/jackc/pgx/v5"
)

func NewPostgresConnection() (*pgx.Conn, error) {
	config, err := config.New[DBConfig]()
	if err != nil {
		return nil, err
	}

	databaseUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database)

	conn, err := pgx.Connect(
		context.Background(),
		databaseUrl)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
