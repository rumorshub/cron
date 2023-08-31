package cron

import (
	"context"

	redislock "github.com/go-co-op/gocron-redis-lock"
	"github.com/redis/go-redis/v9"
	"github.com/roadrunner-server/errors"
)

const LockerPluginName = "cron.locker"

type LockerPlugin struct {
	client redis.UniversalClient
}

func (p *LockerPlugin) Init(cfg Configurer, maker RedisMaker, scheduler Scheduler) error {
	const op = errors.Op("cron.locker_plugin_init")

	if !cfg.Has(LockerPluginName) {
		return errors.E(op, errors.Disabled)
	}

	var lc LockerConfig
	if err := cfg.UnmarshalKey(LockerPluginName, &lc); err != nil {
		return errors.E(op, err)
	}

	client, err := maker.MakeRedis(PluginName)
	if err != nil {
		return errors.E(op, err)
	}

	locker, err := redislock.NewRedisLocker(client, lc.Options()...)
	if err != nil {
		_ = client.Close()

		return errors.E(op, err)
	}

	scheduler.WithDistributedLocker(locker)

	p.client = client

	return nil
}

func (p *LockerPlugin) Serve() chan error {
	return make(chan error, 1)
}

func (p *LockerPlugin) Stop(_ context.Context) error {
	if err := p.client.Close(); err != nil {
		return errors.E(errors.Op("cron.locker_plugin_stop"), err)
	}
	return nil
}

func (p *LockerPlugin) Name() string {
	return LockerPluginName
}
