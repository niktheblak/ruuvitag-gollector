package cmd

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	"tinygo.org/x/bluetooth"

	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
)

var discoverTimeout time.Duration

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

func discover(timeout time.Duration) (ruuviTags []string, err error) {
	adapter := bluetooth.DefaultAdapter
	if err := adapter.Enable(); err != nil {
		return nil, err
	}
	scn := &scanner.BluetoothScanner{adapter}
	d := scanner.NewDiscover(scn, logger)
	ctx, timeoutCancel := context.WithTimeout(context.Background(), timeout)
	defer timeoutCancel()
	ctx, sigIntCancel := signal.NotifyContext(ctx, os.Interrupt)
	defer sigIntCancel()
	ruuviTags, err = d.Discover(ctx)
	return
}
