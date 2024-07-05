package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	initTimeout   time.Duration
	outputCfgFile string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Discover all nearby RuuviTags and create a configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Discovering nearby RuuviTags")
		addrs, err := discover(initTimeout)
		if err != nil {
			return err
		}
		logger.Debug("Discovered RuuviTags", "addrs", addrs)
		builder := new(strings.Builder)
		builder.WriteString("interval = \"0m\"\n")
		builder.WriteString("device = \"default\"\n\n")
		builder.WriteString("[ruuvitags]\n")
		for i, addr := range addrs {
			builder.WriteString(fmt.Sprintf("\"%s\" = \"RuuviTag %d\"\n", addr, i+1))
		}
		if outputCfgFile != "" {
			logger.Info("Writing config to file", "file", outputCfgFile)
			if err := os.WriteFile(outputCfgFile, []byte(builder.String()), 0644); err != nil {
				return err
			}
		} else {
			cmd.Println(builder.String())
		}
		return nil
	},
}

func init() {
	initCmd.Flags().DurationVar(&initTimeout, "timeout", 30*time.Second, "timeout for discovery")
	initCmd.Flags().StringVarP(&outputCfgFile, "output", "o", "", "write config to a file")

	rootCmd.AddCommand(initCmd)
}
