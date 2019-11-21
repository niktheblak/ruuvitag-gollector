package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
	"github.com/spf13/cobra"
)

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect measurements from all specified RuuviTags once",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Starting ruuvitag-gollector")
		return runOnce(scn)
	},
}

func runOnce(scn *scanner.Scanner) error {
	logger.Info("Scanning once")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		cancel()
		scn.Stop()
	}()
	if err := scn.ScanOnce(ctx); err != nil {
		return fmt.Errorf("failed to scan: %w", err)
	}
	logger.Info("Stopping ruuvitag-gollector")
	return nil
}
