package config

import (
	"github.com/kelseyhightower/envconfig"
)

type ApiConfig struct {
	Port              int    `envconfig:"PORT" default:"8888"`
	EnableJWT         bool   `envconfig:"ENABLE_JWT" default:"true"`
	EnableStatic      bool   `envconfig:"ENABLE_STATIC" default:"true"`
	EnableSSL         bool   `envconfig:"ENABLE_SSL" default:"false"`
	EnableCookieToken bool   `envconfig:"ENABLE_COOKIE_TOKEN" default:"true"`
	RootPath          string `envconfig:"ROOT_PATH" default:"/api"`
	APITitle          string `envconfig:"API_TITLE" default:"Sample API"`
	APIVersion        string `envconfig:"API_VERSION" default:"1.0.0"`
	JWTSecret         string `envconfig:"JWT_SECRET" default:"BHqQTg99LmSk$Q,_xe*LM+!P*5PKnR~n"`
}

type DBConfig struct {
	Host     string `envconfig:"DB_HOST" default:"localhost"`
	Port     int    `envconfig:"DB_PORT" default:"5432"`
	User     string `envconfig:"DB_USER" default:"postgres"`
	Password string `envconfig:"DB_PASSWORD" default:"password"`
	Database string `envconfig:"DB_DATABASE" default:"sample"`
}

func (a *ApiConfig) Load() error {
	if err := envconfig.Process("", a); err != nil {
		return err
	}
	return nil
}

func (a *DBConfig) Load() error {
	if err := envconfig.Process("", a); err != nil {
		return err
	}
	return nil
}
