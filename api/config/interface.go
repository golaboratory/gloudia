package config

type Configure interface {
	Load() error
}
