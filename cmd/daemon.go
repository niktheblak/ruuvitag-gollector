package cmd

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Collect measurements from specified RuuviTags continuously",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Starting ruuvitag-gollector")
		interval := viper.GetDuration("interval")
		runAsDaemon(scn, interval)
	},
}

func init() {
	daemonCmd.Flags().Duration("interval", 60*time.Second, "Wait time between RuuviTag device scans, 0 to scan continuously")

	viper.BindPFlags(daemonCmd.Flags())

	rootCmd.AddCommand(daemonCmd)
}

func runAsDaemon(scn *scanner.Scanner, scanInterval time.Duration) {
	if err := scn.Init(device); err != nil {
		logger.Fatal("Failed to initialize device", zap.Error(err))
	}
	ctx := context.Background()
	if scanInterval > 0 {
		scn.ScanWithInterval(ctx, scanInterval)
	} else {
		scn.ScanContinuously(ctx)
	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	select {
	case <-interrupt:
	case <-scn.Quit:
	}
	logger.Info("Stopping ruuvitag-gollector")
	scn.Stop()
}
