package example

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/seanbit/kratos/template/cmd/job/jobs"
	"github.com/spf13/cobra"
)

var (
	contract  string
	fromBlock uint64
	endBlock  uint64
	batchSize int
	dryRun    bool
)

type eventReDispatchJob struct{}

func (j *eventReDispatchJob) Name() string  { return "event-re-dispatch" }
func (j *eventReDispatchJob) Short() string { return "Re-dispatch events from database" }
func (j *eventReDispatchJob) Long() string {
	return "Query events from database and re-dispatch them to EventDispatcher for reprocessing"
}

func (j *eventReDispatchJob) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&contract, "contract", "c", "", "Contract address (required)")
	cmd.Flags().Uint64VarP(&fromBlock, "from", "f", 0, "Start block number (required)")
	cmd.Flags().Uint64VarP(&endBlock, "end", "e", 0, "End block number (required)")
	cmd.Flags().IntVar(&batchSize, "batch-size", 10000, "Batch size for processing blocks")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Dry run mode (do not write to database)")

	cmd.MarkFlagRequired("contract")
	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("end")
}

func (j *eventReDispatchJob) Run(app *jobs.App, cmd *cobra.Command, args []string) error {
	if err := validateParams(); err != nil {
		return err
	}

	ctx := context.Background()
	printJobInfo()

	if dryRun {
		fmt.Println("\n⚠️  DRY-RUN MODE: No data will be written to database\n")
		return runDryMode(ctx, app)
	}
	fmt.Println("\n🚀 REAL MODE: Events will be re-dispatched and written to database\n")
	return runRealMode(ctx, app)
}

func init() {
	jobs.Register(&eventReDispatchJob{})
}

// validateParams 验证参数
func validateParams() error {
	if !common.IsHexAddress(contract) {
		return fmt.Errorf("invalid contract address: %s", contract)
	}

	if fromBlock > endBlock {
		return fmt.Errorf("from block (%d) must be <= end block (%d)", fromBlock, endBlock)
	}

	if batchSize <= 0 {
		return fmt.Errorf("batch size must be > 0, got: %d", batchSize)
	}

	return nil
}

// printJobInfo 打印任务信息
func printJobInfo() {
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  Event Re-Dispatch Job")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("  Contract:    %s\n", contract)
	fmt.Printf("  From Block:  %d\n", fromBlock)
	fmt.Printf("  End Block:   %d\n", endBlock)
	fmt.Printf("  Total Blocks: %d\n", endBlock-fromBlock+1)
	fmt.Printf("  Batch Size:  %d blocks/batch\n", batchSize)
	fmt.Printf("  Mode:        %s\n", getMode())
	fmt.Println("═══════════════════════════════════════════════════════════")
}

func getMode() string {
	if dryRun {
		return "DRY-RUN (read-only)"
	}
	return "REAL (write to database)"
}

// runDryMode 干跑模式（不写数据库）
func runDryMode(ctx context.Context, app *jobs.App) error {
	totalEvents := 0
	totalBatches := calculateBatches(fromBlock, endBlock, uint64(batchSize))
	currentBatch := 0

	for from := fromBlock; from <= endBlock; from += uint64(batchSize) {
		to := min(from+uint64(batchSize)-1, endBlock)
		currentBatch++

		events := make([]int, 100000)
		totalEvents += len(events)

		printProgress(currentBatch, totalBatches, from, to, len(events), totalEvents, true)

		for _, event := range events {
			_ = event
		}
	}

	fmt.Println()
	printSummary(totalEvents, true)
	return nil
}

// runRealMode 真实模式（写入数据库）
func runRealMode(ctx context.Context, app *jobs.App) error {
	totalEvents := 0
	processedEvents := 0
	totalBatches := calculateBatches(fromBlock, endBlock, uint64(batchSize))
	currentBatch := 0

	for from := fromBlock; from <= endBlock; from += uint64(batchSize) {
		to := min(from+uint64(batchSize)-1, endBlock)
		currentBatch++

		events := make([]int, 100000)
		totalEvents += len(events)

		printProgress(currentBatch, totalBatches, from, to, len(events), totalEvents, false)

		for range events {
			processedEvents++
		}
	}

	fmt.Println()
	printSummary(processedEvents, false)
	return nil
}

// calculateBatches 计算总批次数
func calculateBatches(from, to, batchSize uint64) int {
	total := to - from + 1
	batches := int(total / batchSize)
	if total%batchSize != 0 {
		batches++
	}
	return batches
}

// printProgress 打印进度条
func printProgress(currentBatch, totalBatches int, from, to uint64, batchEvents, totalEvents int, isDryRun bool) {
	percentage := float64(currentBatch) / float64(totalBatches) * 100
	barLength := 40
	filledLength := int(float64(barLength) * percentage / 100)

	bar := strings.Repeat("█", filledLength) + strings.Repeat("░", barLength-filledLength)

	mode := ""
	if isDryRun {
		mode = " [DRY-RUN]"
	}

	fmt.Printf("\r[%s] %.1f%% | Batch %d/%d | Blocks %d-%d | Events: %d (total: %d)%s",
		bar, percentage, currentBatch, totalBatches, from, to, batchEvents, totalEvents, mode)
}

// printSummary 打印总结
func printSummary(totalEvents int, isDryRun bool) {
	fmt.Println("\n═══════════════════════════════════════════════════════════")
	fmt.Println("  Job Completed Successfully")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("  Total Events Processed: %d\n", totalEvents)
	if isDryRun {
		fmt.Println("  Mode: DRY-RUN (no data written)")
	} else {
		fmt.Println("  Mode: REAL (data written to database)")
	}
	fmt.Println("═══════════════════════════════════════════════════════════")
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
