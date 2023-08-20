package cron

import "github.com/go-co-op/gocron"

var (
	_ Task         = (*TaskWrapper)(nil)
	_ taskWithTags = (*TaskWrapper)(nil)
	_ taskWithName = (*TaskWrapper)(nil)
	_ Tasks        = (*TasksWrapper)(nil)
)

type Tasks interface {
	Tasks() []interface{}
}

type Task interface {
	Expression() string
	Execute(job gocron.Job) error
}

type taskWithTags interface {
	Tags() []string
}

type taskWithName interface {
	Name() string
}

type TaskWrapper struct {
	execute    func(job gocron.Job) error
	tags       []string
	name       string
	expression string
}

type TasksWrapper struct {
	tasks []interface{}
}

func NewTaskWrapper(name, expression string, execute func(job gocron.Job) error) *TaskWrapper {
	return &TaskWrapper{
		name:       name,
		expression: expression,
		execute:    execute,
	}
}

func (t *TaskWrapper) AddTags(tags ...string) *TaskWrapper {
	t.tags = append(t.tags, tags...)
	return t
}

func (t *TaskWrapper) SetTags(tags ...string) *TaskWrapper {
	t.tags = tags
	return t
}

func (t *TaskWrapper) Tags() []string {
	return t.tags
}

func (t *TaskWrapper) Name() string {
	return t.name
}

func (t *TaskWrapper) Expression() string {
	return t.expression
}

func (t *TaskWrapper) Execute(job gocron.Job) error {
	return t.execute(job)
}

func NewTasksWrapper(tasks ...interface{}) TasksWrapper {
	return TasksWrapper{
		tasks: tasks,
	}
}

func (t TasksWrapper) Tasks() []interface{} {
	return t.tasks
}
