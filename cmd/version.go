package cmd

import (
	"github.com/carlmjohnson/versioninfo"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:          "version",
	Short:        "Print program version information and exit",
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(versioninfo.Short())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
