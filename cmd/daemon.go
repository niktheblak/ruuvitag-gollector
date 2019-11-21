package cmd

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scanContinuously bool
var scanInterval time.Duration

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Collect measurements from specified RuuviTags continuously",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Starting ruuvitag-gollector")
		runAsDaemon(scn, scanInterval)
	},
}

func init() {
	daemonCmd.Flags().BoolVar(&scanContinuously, "continuous", false, "Scan for measurements continuously")
	daemonCmd.Flags().DurationVar(&scanInterval, "interval", 60*time.Second, "Wait time between RuuviTag device scans")

	viper.BindPFlag("continuous", rootCmd.PersistentFlags().Lookup("continuous"))
	viper.BindPFlag("interval", rootCmd.PersistentFlags().Lookup("interval"))
}

func runAsDaemon(scn *scanner.Scanner, scanInterval time.Duration) {
	ctx := context.Background()
	if scanContinuously {
		scn.ScanContinuously(ctx)
	} else {
		scn.ScanWithInterval(ctx, scanInterval)
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
