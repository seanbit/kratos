package crontab

import (
	"github.com/robfig/cron/v3"
	"github.com/seanbit/kratos/webkit/transport/crontab"
)

type JobRegister struct {
	jobs []crontab.Job
}

func (register *JobRegister) Jobs() []crontab.Job {
	return register.jobs
}

type JobWrap struct {
	cron.Job
	name string
	spec string
}

func NewJobWrap(name, spec string, job cron.Job) *JobWrap {
	return &JobWrap{
		Job:  job,
		name: name,
		spec: spec,
	}
}

func (job *JobWrap) Name() string {
	return job.name
}

func (job *JobWrap) Spec() string {
	return job.spec
}
