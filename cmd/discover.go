package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
)

var (
	timeout time.Duration
)

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover all nearby RuuviTags",
	RunE: func(cmd *cobra.Command, args []string) error {
		d, err := scanner.NewDiscover(device, &scanner.GoBLEScanner{}, &scanner.GoBLEDeviceCreator{}, logger)
		if err != nil {
			return err
		}
		defer d.Close()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		addrs, err := d.Discover(ctx)
		if err != nil {
			return err
		}
		for _, addr := range addrs {
			cmd.Println(addr)
		}
		return nil
	},
}

func init() {
	discoverCmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "timeout for discovery")

	rootCmd.AddCommand(discoverCmd)
}
