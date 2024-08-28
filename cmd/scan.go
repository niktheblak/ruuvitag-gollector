package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
)

var (
	scanTimeout time.Duration
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan measurements from all specified RuuviTags once",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Starting ruuvitag-gollector")
		if err := createExporters(); err != nil {
			return err
		}
		cfg := scanner.DefaultConfig()
		cfg.DeviceName = device
		cfg.Peripherals = peripherals
		cfg.Exporters = exporters
		cfg.Logger = logger
		scn, err := scanner.NewOnce(cfg)
		if err != nil {
			return err
		}
		logger.Info("Scanning once")
		ctx, timeoutCancel := context.WithTimeout(context.Background(), scanTimeout)
		defer timeoutCancel()
		ctx, sigIntCancel := signal.NotifyContext(ctx, os.Interrupt)
		defer sigIntCancel()
		err = scn.Scan(ctx, 0)
		switch {
		case errors.Is(err, context.DeadlineExceeded):
		case errors.Is(err, context.Canceled):
		case err == nil:
		default:
			return fmt.Errorf("failed to scan: %w", err)
		}
		logger.Info("Scan completed")
		err = scn.Close()
		logger.Info("Stopping ruuvitag-gollector")
		return errors.Join(err, closeExporters())
	},
}

func init() {
	scanCmd.Flags().DurationVar(&scanTimeout, "timeout", 30*time.Second, "timeout for scan")

	rootCmd.AddCommand(scanCmd)
}
