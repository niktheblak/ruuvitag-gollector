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
			scn.Exporters = exporters
			return runWithInterval(scn, interval)
		} else {
			scn := scanner.NewContinuous(logger, peripherals)
			scn.Exporters = exporters
			return runContinuously(scn)
		}
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		logger.Info("Stopping ruuvitag-gollector")
	},
}

func init() {
	daemonCmd.Flags().Duration("interval", 60*time.Second, "Wait time between RuuviTag device scans, 0 to scan continuously")

	cobra.CheckErr(viper.BindPFlags(daemonCmd.Flags()))

	rootCmd.AddCommand(daemonCmd)
}

func runWithInterval(scn *scanner.Scanner, scanInterval time.Duration) error {
	if err := scn.Init(device); err != nil {
		return err
	}
	ctx := context.Background()
	scn.Scan(ctx, scanInterval)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	select {
	case <-interrupt:
	case <-scn.Quit:
	}
	scn.Stop()
	return nil
}

func runContinuously(scn *scanner.ContinuousScanner) error {
	if err := scn.Init(device); err != nil {
		return err
	}
	ctx := context.Background()
	scn.Scan(ctx)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	select {
	case <-interrupt:
	case <-scn.Quit:
	}
	scn.Stop()
	return nil
}
