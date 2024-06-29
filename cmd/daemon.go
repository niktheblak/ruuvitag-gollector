package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Collect measurements from specified RuuviTags continuously",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := createExporters(); err != nil {
			return err
		}
		interval := viper.GetDuration("interval")
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		var scn scanner.Scanner
		var err error
		if interval == 0 {
			scn, err = scanner.NewContinuous(device, peripherals, exporters, logger)
		} else {
			scn, err = scanner.NewInterval(device, peripherals, exporters, logger)
		}
		if err != nil {
			return err
		}
		err = scn.Scan(ctx, interval)
		return errors.Join(err, scn.Close(), closeExporters())
	},
}

func init() {
	daemonCmd.Flags().Duration("interval", 60*time.Second, "Wait time between RuuviTag device scans, 0 to scan continuously")

	cobra.CheckErr(viper.BindPFlags(daemonCmd.Flags()))

	rootCmd.AddCommand(daemonCmd)
}
