package cmd

import (
	"github.com/spf13/cobra"
)

var printCmd = &cobra.Command{
	Use:          "print",
	Short:        "Print requested configuration values and exit",
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(printCmd)
}
