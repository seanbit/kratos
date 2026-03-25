package main

import (
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/seanbit/kratos/template/cmd/job/jobs/example"
	"github.com/seanbit/kratos/webkit"
	"github.com/spf13/cobra"

	"evm-scan/cmd/job/jobs"
	"evm-scan/internal/global"
)

var (
	configFile string
	secretFile string
	rootCmd    = &cobra.Command{
		Use:   "job",
		Short: "EVM Scan Job Runner",
		Long:  "Command-line tool for running various EVM scan jobs",
	}
)

func init() {
	// 全局标志
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "configs/config.yaml", "config file path")
	rootCmd.PersistentFlags().StringVar(&secretFile, "secret", "", "secret file name")

	// 注册 event-re-dispatch 命令的参数
	example.RegisterEventReDispatchFlags(eventReDispatchCmd)

	// 添加子命令
	rootCmd.AddCommand(eventReDispatchCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(versionCmd)
}

func getSubCommandRunE(fn func(cmd *cobra.Command, app *jobs.App) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// 初始化配置
		cleanConfig := global.InitConfig("file", configFile, secretFile)
		defer cleanConfig()
		cfg := global.GetConfig()

		// 初始化logger
		webkit.InitLogger(rootCmd.Use, versionCmd.Version, int(cfg.LogLevel))

		// 初始化依赖注入
		app, cleanup, err := initApp(cfg.Server, cfg.Data, cfg.Blockchain, cfg.Scanner, log.DefaultLogger)
		if err != nil {
			return fmt.Errorf("failed to init app: %w", err)
		}
		defer cleanup()

		return fn(cmd, app)
	}
}

var eventReDispatchCmd = &cobra.Command{
	Use:   "event-re-dispatch",
	Short: "Re-dispatch events from database",
	Long:  "Query events from database and re-dispatch them to EventDispatcher for reprocessing",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 初始化配置
		cleanConfig := global.InitConfig("file", configFile, secretFile)
		defer cleanConfig()
		cfg := global.GetConfig()

		// 初始化logger
		webkit.InitLogger(rootCmd.Use, versionCmd.Version, int(cfg.LogLevel))

		// 初始化依赖注入
		app, cleanup, err := initApp(cfg.Server, cfg.Data, cfg.Blockchain, cfg.Scanner, log.DefaultLogger)
		if err != nil {
			return fmt.Errorf("failed to init app: %w", err)
		}
		defer cleanup()

		return example.RunEventReDispatch(cmd, app)
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a specific job",
	Long:  "Run a specific job (deprecated, use subcommands directly)",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("job version 1.0.0")
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
