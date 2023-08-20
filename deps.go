package cron

import (
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type Configurer interface {
	Has(name string) bool
	UnmarshalKey(name string, out interface{}) error
}

type RedisMaker interface {
	MakeRedis(name string) (redis.UniversalClient, error)
}

type Logger interface {
	NamedLogger(name string) *slog.Logger
}
