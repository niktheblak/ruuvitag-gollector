package cmd

import (
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
	"github.com/spf13/cobra"
)

var printColumnsCmd = &cobra.Command{
	Use:          "columns",
	Short:        "Print default RuuviTag column names",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, c := range sensor.DefaultColumns {
			cmd.Printf("%s\n", c)
		}
		return nil
	},
}

func init() {
	printCmd.AddCommand(printColumnsCmd)
}
