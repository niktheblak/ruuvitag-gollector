package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	output string
)

var printConfigCmd = &cobra.Command{
	Use:          "config",
	Short:        "Print active configuration",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if output != "" {
			return viper.WriteConfigAs(output)
		}
		keys := viper.AllKeys()
		sort.Strings(keys)
		for _, key := range keys {
			if viper.Get(key) != "" {
				fmt.Printf("%s = %v\n", key, viper.Get(key))
			}
		}
		return nil
	},
}

func init() {
	printConfigCmd.Flags().StringVarP(&output, "output", "o", "", "print config into a file")

	printCmd.AddCommand(printConfigCmd)
}
