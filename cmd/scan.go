package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
)

var scanTimeout time.Duration

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
			return errors.Join(err, closeExporters())
		}
		logger.Info("Scanning once")
		ctx, timeoutCancel := context.WithTimeout(context.Background(), scanTimeout)
		defer timeoutCancel()
		ctx, sigIntCancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer sigIntCancel()
		scanErr := scn.Scan(ctx, 0)
		closeErr := errors.Join(scn.Close(), closeExporters())
		switch {
		case errors.Is(scanErr, context.DeadlineExceeded):
		case errors.Is(scanErr, context.Canceled):
		case scanErr == nil:
		default:
			return errors.Join(fmt.Errorf("failed to scan: %w", scanErr), closeErr)
		}
		logger.Info("Scan completed")
		logger.Info("Stopping ruuvitag-gollector")
		return closeErr
	},
}

func init() {
	scanCmd.Flags().DurationVar(&scanTimeout, "timeout", 30*time.Second, "timeout for scan")

	rootCmd.AddCommand(scanCmd)
}
