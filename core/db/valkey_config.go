package db

type ValkeyConfig struct {
	Host string `envconfig:"DB_HOST" default:"localhost"`
	Port string `envconfig:"DB_PORT" default:"15432"`
}
