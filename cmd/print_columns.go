package cmd

import (
	"github.com/spf13/cobra"

	"github.com/niktheblak/ruuvitag-gollector/pkg/psql"
)

var printColumnsCmd = &cobra.Command{
	Use:          "columns",
	Short:        "Print default RuuviTag column names",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, c := range psql.DefaultColumnNames {
			cmd.Printf("%s\n", c)
		}
		return nil
	},
}

func init() {
	printCmd.AddCommand(printColumnsCmd)
}
