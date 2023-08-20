package cron

import (
	"runtime"
	"time"

	redislock "github.com/go-co-op/gocron-redis-lock"
	"github.com/go-redsync/redsync/v4"
)

type LockerConfig struct {
	// Tries can be used to set the number of times lock acquire is attempted.
	Tries *int `mapstructure:"tries" json:"tries,omitempty" bson:"tries,omitempty"`

	// DriftFactor can be used to set the clock drift factor.
	DriftFactor *float64 `mapstructure:"drift_factor" json:"drift_factor,omitempty" bson:"drift_factor,omitempty"`

	// TimeoutFactor can be used to set the timeout factor.
	TimeoutFactor *float64 `mapstructure:"timeout_factor" json:"timeout_factor,omitempty" bson:"timeout_factor,omitempty"`

	// Expiry can be used to set the expiry of a mutex to the given value.
	Expiry *time.Duration `mapstructure:"expiry" json:"expiry,omitempty" bson:"expiry,omitempty"`

	// RetryDelay can be used to set the amount of time to wait between retries.
	RetryDelay *time.Duration `mapstructure:"retry_delay" json:"retry_delay,omitempty" bson:"retry_delay,omitempty"`
}

func (c *LockerConfig) InitDefaults() {
	if c.Tries == nil {
		c.Tries = toPtr(32)
	}
	if c.DriftFactor == nil {
		c.DriftFactor = toPtr(0.01)
	}
	if c.TimeoutFactor == nil {
		c.TimeoutFactor = toPtr(0.05)
	}
	if c.Expiry == nil {
		c.Expiry = toPtr(8 * time.Second)
	}
}

func (c *LockerConfig) Options() []redsync.Option {
	c.InitDefaults()

	opts := []redsync.Option{
		redislock.WithTries(*c.Tries),
		redislock.WithDriftFactor(*c.DriftFactor),
		redislock.WithTimeoutFactor(*c.TimeoutFactor),
		redislock.WithExpiry(*c.Expiry),
	}

	if c.RetryDelay != nil {
		opts = append(opts, redislock.WithRetryDelay(*c.RetryDelay))
	}

	return opts
}

type Config struct {
	MaxConcurrentJobs  int  `mapstructure:"max_concurrent_jobs" json:"max_concurrent_jobs,omitempty" bson:"max_concurrent_jobs,omitempty"`
	TagsUnique         bool `mapstructure:"tags_unique" json:"tags_unique,omitempty" bson:"tags_unique,omitempty"`
	WaitForScheduleAll bool `mapstructure:"wait_for_schedule_all" json:"wait_for_schedule_all,omitempty" bson:"wait_for_schedule_all,omitempty"`
	WaitForSchedule    bool `mapstructure:"wait_for_schedule" json:"wait_for_schedule,omitempty" bson:"wait_for_schedule,omitempty"`
}

func (c *Config) InitDefaults() {
	if c.MaxConcurrentJobs == 0 {
		c.MaxConcurrentJobs = runtime.NumCPU()
	}
}

func toPtr[T any](val T) *T {
	return &val
}
