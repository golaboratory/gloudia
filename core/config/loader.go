package config

import (
	"github.com/kelseyhightower/envconfig"
)

func New[T interface{}]() (T, error) {
	v := *new(T)
	if err := envconfig.Process("", &v); err != nil {
		return *new(T), err
	}
	return v, nil
}
