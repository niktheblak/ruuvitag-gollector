package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Collect measurements from specified RuuviTags continuously",
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := createExporters(); err != nil {
			return err
		}
		interval := viper.GetDuration("interval")
		cfg := scanner.DefaultConfig()
		cfg.DeviceName = device
		cfg.Peripherals = peripherals
		cfg.Exporters = exporters
		cfg.Logger = logger
		var scn scanner.Scanner
		var err error
		if interval == 0 {
			scn, err = scanner.NewContinuous(cfg)
		} else {
			scn, err = scanner.NewInterval(cfg)
		}
		if err != nil {
			return errors.Join(err, closeExporters())
		}
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		err = scn.Scan(ctx, interval)
		return errors.Join(err, scn.Close(), closeExporters())
	},
}

func init() {
	daemonCmd.Flags().Duration("interval", 60*time.Second, "Wait time between RuuviTag device scans, 0 to scan continuously")

	cobra.CheckErr(viper.BindPFlags(daemonCmd.Flags()))

	rootCmd.AddCommand(daemonCmd)
}
