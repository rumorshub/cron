package cron

import (
	"context"
	"sync"

	"github.com/go-co-op/gocron"
	redislock "github.com/go-co-op/gocron-redis-lock"
	"github.com/roadrunner-server/endure/v2/dep"
	"github.com/roadrunner-server/errors"
)

const (
	PluginName = "cron"

	lockerKey = PluginName + ".locker"
)

type Plugin struct {
	mu sync.RWMutex

	scheduler *scheduler
	tasks     []Task
}

func (p *Plugin) Init(cfg Configurer, maker RedisMaker, logger Logger) error {
	const op = errors.Op("cron_plugin_init")

	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	var c Config
	if err := cfg.UnmarshalKey(PluginName, &c); err != nil {
		return errors.E(op, err)
	}

	var locker gocron.Locker
	if cfg.Has(lockerKey) {
		var lc LockerConfig
		if err := cfg.UnmarshalKey(lockerKey, &lc); err != nil {
			return errors.E(op, err)
		}

		client, err := maker.MakeRedis(PluginName)
		if err != nil {
			return errors.E(op, err)
		}

		locker, err = redislock.NewRedisLocker(client, lc.Options()...)
		if err != nil {
			return errors.E(op, err)
		}
	}

	p.scheduler = newScheduler(c, locker, logger.NamedLogger(PluginName).WithGroup("scheduler"))

	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)

	p.scheduler.StartAsync()

	go func() {
		p.mu.RLock()
		defer p.mu.RUnlock()

		for _, task := range p.tasks {
			if _, err := p.scheduler.addTask(task); err != nil {
				errCh <- err
				break
			}
		}
	}()

	return errCh
}

func (p *Plugin) Stop(_ context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	p.scheduler.Stop()

	return nil
}

func (p *Plugin) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(pp interface{}) {
			task := pp.(Task)

			p.mu.Lock()
			p.tasks = append(p.tasks, task)
			p.mu.Unlock()
		}, (*Task)(nil)),
		dep.Fits(func(pp interface{}) {
			t := pp.(Tasks)

			p.mu.Lock()
			for _, task := range t.Tasks() {
				p.tasks = append(p.tasks, task.(Task))
			}
			p.mu.Unlock()
		}, (*Tasks)(nil)),
	}
}

func (p *Plugin) Provides() []*dep.Out {
	return []*dep.Out{
		dep.Bind((*Scheduler)(nil), p.Scheduler),
	}
}

func (p *Plugin) Scheduler() Scheduler {
	return p.scheduler
}

func (p *Plugin) Name() string {
	return PluginName
}
