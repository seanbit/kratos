package crontab

import (
	"github.com/google/wire"
	"github.com/robfig/cron/v3"
	"github.com/seanbit/kratos/webkit/transport/crontab"

	"github.com/seanbit/kratos/template/internal/conf"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewJobTest,
	NewJobRegister,
	NewCrontabExecutor,
)

func NewJobRegister(cfg *conf.CronJob, test *JobTest) crontab.JobRegister {
	allJobs := map[string]cron.Job{
		"test": test,
	}

	var jobs []crontab.Job
	for _, entry := range cfg.GetJobs() {
		if entry.GetDisabled() {
			continue
		}
		if job, ok := allJobs[entry.GetName()]; ok {
			jobs = append(jobs, NewJobWrap(entry.GetName(), entry.GetSpec(), job))
		}
	}

	return &JobRegister{jobs: jobs}
}

func NewCrontabExecutor(register crontab.JobRegister) *crontab.Executor {
	return crontab.NewServer(register)
}
