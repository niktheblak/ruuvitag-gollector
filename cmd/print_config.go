package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var printConfigCmd = &cobra.Command{
	Use:          "print-config",
	Short:        "Print configuration and exit",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if viper.ConfigFileUsed() != "" {
			fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
		}
		keys := viper.AllKeys()
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Printf("%s = %v\n", key, viper.Get(key))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(printConfigCmd)
}
