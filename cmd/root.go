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

const (
	logLevelConfigKey  = "log.level"
	logFormatConfigKey = "log.format"
	deviceConfigKey    = "device"
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
		panic(err)
	}
}

func init() {
	logger = slog.Default()

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentFlags().StringToString("ruuvitags", nil, "RuuviTag addresses and names to use")
	rootCmd.PersistentFlags().StringToString("columns", nil, "RuuviTag fields to use and their column names")
	rootCmd.PersistentFlags().String(deviceConfigKey, "", "HCL device to use")
	rootCmd.PersistentFlags().String(logLevelConfigKey, "info", "Log level")
	rootCmd.PersistentFlags().String(logFormatConfigKey, "text", "Log level")

	viper.SetDefault(deviceConfigKey, "default")
	viper.SetDefault(logLevelConfigKey, "info")
	viper.SetDefault(logFormatConfigKey, "text")
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
		fmt.Println("No config file found, using only CLI args and env vars")
	}
	logLevelCfg := viper.GetString(logLevelConfigKey)
	var logLevel = new(slog.LevelVar)
	if err := logLevel.UnmarshalText([]byte(logLevelCfg)); err != nil {
		panic(fmt.Sprintf("invalid log level %s: %v", logLevelCfg, err))
	}
	logFormat := viper.GetString(logFormatConfigKey)
	var logHandler slog.Handler
	switch logFormat {
	case "text":
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	case "json":
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	default:
		panic(fmt.Sprintf("invalid log format: %s", logFormat))
	}
	logger = slog.New(logHandler)
}
