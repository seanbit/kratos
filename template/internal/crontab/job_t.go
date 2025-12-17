package crontab

import "github.com/go-kratos/kratos/v2/log"

type JobTest struct {
}

func NewJobTest() *JobTest {
	return &JobTest{}
}

func (job *JobTest) Run() {
	log.Debugf("job test run")
}
