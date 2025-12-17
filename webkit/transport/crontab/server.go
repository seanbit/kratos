package crontab

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

var _ transport.Server = (*Executor)(nil)

type JobInfo interface {
	Name() string
	Spec() string
}

type Job interface {
	cron.Job
	JobInfo
}

type JobRegister interface {
	Jobs() []Job
}

type Executor struct {
	*cron.Cron
	jobs []Job
}

func NewServer(register JobRegister) *Executor {
	crontab := &Executor{
		Cron: cron.New(cron.WithSeconds()),
		jobs: register.Jobs(),
	}
	return crontab
}

func (e *Executor) Start(ctx context.Context) error {
	for _, job := range e.jobs {
		entryId, err := e.AddJob(job.Spec(), job)
		if err != nil {
			return errors.Wrapf(err, "crontab.AddJob(%s)", job.Name())
		}
		log.Infof("crontab.AddJob(%s) success:%d", job.Name(), entryId)
	}
	return nil
}

func (e *Executor) Stop(ctx context.Context) error {
	e.Cron.Stop()
	return nil
}
