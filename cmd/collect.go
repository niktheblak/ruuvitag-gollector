package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
)

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect measurements from all specified RuuviTags once",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Starting ruuvitag-gollector")
		scn := scanner.NewOnce(logger, peripherals)
		scn.Exporters = exporters
		return runOnce(scn)
	},
}

func init() {
	rootCmd.AddCommand(collectCmd)
}

func runOnce(scn *scanner.OnceScanner) error {
	if err := scn.Init(device); err != nil {
		return err
	}
	logger.Info("Scanning once")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		cancel()
	}()
	if err := scn.Scan(ctx); err != nil {
		return fmt.Errorf("failed to scan: %w", err)
	}
	logger.Info("Stopping scanner")
	scn.Close()
	return nil
}
