//go:build timescaledb

package cmd

import (
	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/timescaledb"
)

func init() {
	rootCmd.PersistentFlags().Bool("timescaledb.enabled", false, "enable TimescaleDB exporter")
	rootCmd.PersistentFlags().String("timescaledb.host", "", "database host or IP")
	rootCmd.PersistentFlags().Int("timescaledb.port", 0, "database port")
	rootCmd.PersistentFlags().String("timescaledb.username", "", "database username")
	rootCmd.PersistentFlags().String("timescaledb.password", "", "database password")
	rootCmd.PersistentFlags().String("timescaledb.database", "", "database name")
	rootCmd.PersistentFlags().String("timescaledb.table", "", "table name")
	rootCmd.PersistentFlags().String("timescaledb.sslmode", "", "SSL mode")
	rootCmd.PersistentFlags().String("timescaledb.sslcert", "", "path to SSL certificate file")
	rootCmd.PersistentFlags().String("timescaledb.sslkey", "", "path to SSL key file")

	viper.SetDefault("timescaledb.port", "5432")
	viper.SetDefault("timescaledb.sslmode", "disable")
}

func addTimescaleDBExporter(exporters *[]exporter.Exporter) error {
	psqlInfo, err := CreatePsqlInfoString("timescaledb")
	if err != nil {
		return err
	}
	logger.Info("Connecting to TimescaleDB", "connstr", SanitizePassword(psqlInfo))
	exp, err := timescaledb.New(psqlInfo, viper.GetString("timescaledb.table"))
	if err != nil {
		return err
	}
	*exporters = append(*exporters, exp)
	return nil
}
