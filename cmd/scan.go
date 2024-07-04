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
		ctx, timeoutCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer timeoutCancel()
		ctx, sigIntCancel := signal.NotifyContext(ctx, os.Interrupt)
		defer sigIntCancel()
		if err := scn.Scan(ctx, 0); err != nil {
			return fmt.Errorf("failed to scan: %w", err)
		}
		err = scn.Close()
		logger.Info("Stopping ruuvitag-gollector")
		return errors.Join(err, closeExporters())
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
