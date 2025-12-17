package crontab

import (
	"github.com/google/wire"
	"github.com/seanbit/kratos/webkit/transport/crontab"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewJobTest,
	NewJobRegister,
)

func NewJobRegister(test *JobTest) crontab.JobRegister {
	return &JobRegister{
		jobs: []crontab.Job{
			NewJobWrap("test", "0 * * * * *", test),
		},
	}
}
