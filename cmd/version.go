package cmd

import (
	"runtime/debug"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:          "version",
	Short:        "Print program version information and exit",
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return
		}
		for _, kv := range info.Settings {
			switch kv.Key {
			case "vcs.revision":
				cmd.Printf("Revision: %s\n", kv.Value)
			case "vcs.time":
				cmd.Printf("Time: %s\n", kv.Value)
			case "vcs.modified":
				cmd.Printf("Modified: %v\n", kv.Value == "true")
			case "GOARCH":
				cmd.Printf("Architecture: %s\n", kv.Value)
			case "GOOS":
				cmd.Printf("OS: %s\n", kv.Value)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
