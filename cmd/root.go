package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

var ErrNotEnabled = errors.New("this exporter is not included in the build")

var (
	cfgFile     string
	logger      *slog.Logger
	peripherals map[string]string
	exporters   []exporter.Exporter
	device      string
)

var rootCmd = &cobra.Command{
	Use:          "ruuvitag-gollector",
	Short:        "Collects measurements from RuuviTag sensors",
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	logger = slog.Default()

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentFlags().StringToString("ruuvitags", nil, "RuuviTag addresses and names to use")
	rootCmd.PersistentFlags().StringToString("columns", nil, "RuuviTag fields to use and their column names")
	rootCmd.PersistentFlags().String("device", "", "HCL device to use")
	rootCmd.PersistentFlags().String("log.level", "info", "Log level")
	rootCmd.PersistentFlags().String("log.format", "text", "Log level")

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "text")
	viper.SetDefault("device", "default")
}

func initConfig() {
	cobra.CheckErr(viper.BindPFlags(rootCmd.PersistentFlags()))
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.ruuvitag-gollector")
		viper.AddConfigPath("/etc/ruuvitag-gollector/")
	}
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := viper.ReadInConfig(); err != nil {
		// configuration file does not exist; only use CLI args and env
	}
	logLevelCfg := viper.GetString("log.level")
	var logLevel = new(slog.LevelVar)
	if err := logLevel.UnmarshalText([]byte(logLevelCfg)); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid log level: %s\n", err)
		os.Exit(1)
	}
	logFormat := viper.GetString("log.format")
	var logHandler slog.Handler
	switch logFormat {
	case "text":
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	case "json":
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	default:
		fmt.Fprintf(os.Stderr, "Invalid log format: %s\n", logFormat)
		os.Exit(1)
	}
	logger = slog.New(logHandler)
}
