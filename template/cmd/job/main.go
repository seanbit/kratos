package main

import (
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/seanbit/kratos/webkit"
	"github.com/spf13/cobra"

	"github.com/seanbit/kratos/template/cmd/job/jobs"
	"github.com/seanbit/kratos/template/internal/global"

	// blank import: 触发各 job 包的 init() 完成自注册
	_ "github.com/seanbit/kratos/template/cmd/job/jobs/example"
)

var (
	configFile string
	secretFile string
	rootCmd    = &cobra.Command{
		Use:   "job",
		Short: "Job Runner",
		Long:  "Command-line tool for running various jobs",
	}
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("job version 1.0.0")
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "configs/config.yaml", "config file path")
	rootCmd.PersistentFlags().StringVar(&secretFile, "secret", "", "secret file name")

	// 遍历已注册的 jobs，自动生成子命令
	for _, job := range jobs.AllJobs() {
		cmd := &cobra.Command{
			Use:   job.Name(),
			Short: job.Short(),
			Long:  job.Long(),
			RunE:  getSubCommandRunE(job.Run),
		}
		// 如果 job 需要注册 flags
		if fr, ok := job.(jobs.FlagsRegistrar); ok {
			fr.RegisterFlags(cmd)
		}
		rootCmd.AddCommand(cmd)
	}
	rootCmd.AddCommand(versionCmd)
}

func getSubCommandRunE(fn func(*jobs.App, *cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cleanConfig := global.InitConfig("file", configFile, secretFile)
		defer cleanConfig()
		cfg := global.GetConfig()

		webkit.InitLogger(rootCmd.Use, versionCmd.Version, int(cfg.LogLevel))

		app, cleanup, err := initApp(log.DefaultLogger)
		if err != nil {
			return fmt.Errorf("failed to init app: %w", err)
		}
		defer cleanup()

		return fn(app, cmd, args)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
