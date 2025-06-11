package cmd

import (
	"context"
	"errors"
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

func discover(timeout time.Duration) (ruuviTags []string, err error) {
	var d *scanner.Discover
	d, err = scanner.NewDiscover(device, &scanner.GoBLEScanner{}, &scanner.GoBLEDeviceCreator{}, logger)
	if err != nil {
		return
	}
	defer func() {
		closeErr := d.Close()
		if closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()
	ctx, timeoutCancel := context.WithTimeout(context.Background(), timeout)
	defer timeoutCancel()
	ctx, sigIntCancel := signal.NotifyContext(ctx, os.Interrupt)
	defer sigIntCancel()
	ruuviTags, err = d.Discover(ctx)
	return
}
