package cron

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
)

var _ Scheduler = (*scheduler)(nil)

var ErrExpressionEmpty = errors.New("expression is empty")

type Scheduler interface {
	WithDistributedLocker(locker gocron.Locker)

	StartAsync()
	Stop()
	IsRunning() bool

	Jobs() []*gocron.Job
	GetAllTags() []string

	Remove(jobFun interface{})
	RemoveByReference(job *gocron.Job)
	RemoveByTag(tag string) error
	RemoveByTags(tags ...string) error
	RemoveByTagsAny(tags ...string) error

	FindJobsByTag(tags ...string) ([]*gocron.Job, error)
	TaskPresent(jobFun interface{}) bool

	Clear()

	RegisterEventListeners(eventListeners ...gocron.EventListener)

	UpdateJob(job *gocron.Job) error
	AddTask(expression string, jobFun interface{}, name string, tags ...string) (*gocron.Job, error)
}

type scheduler struct {
	*gocron.Scheduler
	waitForSchedule bool
	log             *slog.Logger
	taskLog         *slog.Logger
}

func newScheduler(cfg Config, log *slog.Logger) *scheduler {
	cfg.InitDefaults()

	s := gocron.NewScheduler(time.UTC)
	s.SetMaxConcurrentJobs(cfg.MaxConcurrentJobs, gocron.RescheduleMode)

	if cfg.TagsUnique {
		s.TagsUnique()
	}

	if cfg.WaitForScheduleAll {
		s.WaitForScheduleAll()
	}

	return &scheduler{
		Scheduler:       s,
		waitForSchedule: cfg.WaitForSchedule,
		log:             log,
		taskLog:         log.WithGroup("task"),
	}
}

func (s *scheduler) UpdateJob(job *gocron.Job) error {
	_, err := s.Job(job).Update()
	return err
}

func (s *scheduler) AddTask(expression string, jobFun interface{}, name string, tags ...string) (*gocron.Job, error) {
	if expression == "" {
		return nil, ErrExpressionEmpty
	}

	switch strings.Count(expression, " ") {
	case 0:
		s.Every(expression)
	case 4:
		s.Cron(expression)
	case 5:
		s.CronWithSeconds(expression)
	default:
		return nil, fmt.Errorf("invalid task expression `%s`", expression)
	}

	if name != "" {
		s.Name(name)
	}

	if len(tags) > 0 {
		s.Tag(tags...)
	}

	s.log.Info("task add", "expression", expression, "name", name, "tags", tags)

	job, err := s.DoWithJobDetails(jobFun)
	if err != nil {
		return nil, err
	}

	job.RegisterEventListeners(gocron.BeforeJobRuns(func(jobName string) {
		s.taskLog.WithGroup(jobName).Info("run starting")
	}), gocron.AfterJobRuns(func(jobName string) {
		s.taskLog.WithGroup(jobName).Info("run ending")
	}), gocron.WhenJobReturnsError(func(jobName string, err error) {
		s.taskLog.WithGroup(jobName).Error("task returns error", "error", err)
	}))

	return job, nil
}

func (s *scheduler) addTask(task Task) (*gocron.Job, error) {
	var (
		name string
		tags []string
	)

	if t, ok := task.(taskWithName); ok {
		name = t.Name()
	}

	if t, ok := task.(taskWithTags); ok {
		tags = t.Tags()
	}

	return s.AddTask(task.Expression(), task.Execute, name, tags...)
}
