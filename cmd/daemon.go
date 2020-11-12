package cmd

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Collect measurements from specified RuuviTags continuously",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Starting ruuvitag-gollector")
		interval := viper.GetDuration("interval")
		if interval > 0 {
			scn := scanner.NewInterval(logger, peripherals)
			if err := scn.Init(device); err != nil {
				return err
			}
			runWithInterval(scn, interval)
		} else {
			scn := scanner.NewContinuous(logger, peripherals)
			if err := scn.Init(device); err != nil {
				return err
			}
			runContinuously(scn)
		}
		return nil
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		logger.Info("Stopping ruuvitag-gollector")
	},
}

func init() {
	daemonCmd.Flags().Duration("interval", 60*time.Second, "Wait time between RuuviTag device scans, 0 to scan continuously")

	viper.BindPFlags(daemonCmd.Flags())

	rootCmd.AddCommand(daemonCmd)
}

func runWithInterval(scn *scanner.Scanner, scanInterval time.Duration) {
	ctx := context.Background()
	scn.Scan(ctx, scanInterval)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	select {
	case <-interrupt:
	case <-scn.Quit:
	}
	scn.Stop()
}

func runContinuously(scn *scanner.ContinuousScanner) {
	ctx := context.Background()
	scn.Scan(ctx)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	select {
	case <-interrupt:
	case <-scn.Quit:
	}
	scn.Stop()
}
