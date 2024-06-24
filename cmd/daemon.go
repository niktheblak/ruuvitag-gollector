package cmd

import (
	"context"
	"errors"
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
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if err = start(); err != nil {
			return
		}
		defer func() {
			err = errors.Join(err, stop())
		}()
		interval := viper.GetDuration("interval")
		if interval > 0 {
			scn := scanner.NewInterval(logger, peripherals)
			scn.Exporters = exporters
			err = runWithInterval(scn, interval)
		} else {
			scn := scanner.NewContinuous(logger, peripherals)
			scn.Exporters = exporters
			err = runContinuously(scn)
		}
		return
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
