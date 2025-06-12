package cmd

import (
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
	"github.com/spf13/cobra"
)

var format string

var printColumnsCmd = &cobra.Command{
	Use:          "columns",
	Short:        "Print default RuuviTag column names",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch format {
		case "toml":
			cmd.Println("[columns]")
			for _, c := range sensor.DefaultColumns {
				cmd.Printf("\"%s\" = \"%s\"\n", c, sensor.DefaultColumnMap[c])
			}
		case "yaml":
			cmd.Println("columns:")
			for _, c := range sensor.DefaultColumns {
				cmd.Printf("  %s: %s\n", c, sensor.DefaultColumnMap[c])
			}
		default:
			for _, c := range sensor.DefaultColumns {
				cmd.Printf("%s\n", c)
			}
		}
		return nil
	},
}

func init() {
	printColumnsCmd.Flags().StringVarP(&format, "format", "f", "", "write columns in specified format")

	printCmd.AddCommand(printColumnsCmd)
}
