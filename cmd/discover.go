package cmd

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
)

var (
	discoverTimeout time.Duration
)

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover all nearby RuuviTags",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Discovering nearby RuuviTags")
		addrs, err := discover(discoverTimeout)
		if err != nil {
			return err
		}
		logger.Debug("Discovered RuuviTags", "addrs", addrs)
		for _, addr := range addrs {
			cmd.Println(addr)
		}
		return nil
	},
}

func init() {
	discoverCmd.Flags().DurationVar(&discoverTimeout, "timeout", 30*time.Second, "timeout for discovery")

	rootCmd.AddCommand(discoverCmd)
}

func discover(timeout time.Duration) ([]string, error) {
	d, err := scanner.NewDiscover(device, &scanner.GoBLEScanner{}, &scanner.GoBLEDeviceCreator{}, logger)
	if err != nil {
		return nil, err
	}
	defer d.Close()
	ctx, timeoutCancel := context.WithTimeout(context.Background(), timeout)
	defer timeoutCancel()
	ctx, sigIntCancel := signal.NotifyContext(ctx, os.Interrupt)
	defer sigIntCancel()
	return d.Discover(ctx)
}
