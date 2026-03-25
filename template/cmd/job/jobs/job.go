package jobs

import "github.com/spf13/cobra"

// Job 定义一个 job 命令的基本信息
type Job interface {
	Name() string
	Short() string
	Long() string
}

// JobRunner 带执行逻辑的 job
type JobRunner interface {
	Job
	Run(app *App, cmd *cobra.Command, args []string) error
}

// FlagsRegistrar 需要注册 flag 的 job 实现此接口
type FlagsRegistrar interface {
	RegisterFlags(cmd *cobra.Command)
}

var registry []JobRunner

// Register 注册一个 job（在各 job 包的 init() 中调用）
func Register(j JobRunner) {
	registry = append(registry, j)
}

// AllJobs 返回所有已注册的 jobs
func AllJobs() []JobRunner {
	return registry
}
