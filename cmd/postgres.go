//go:build postgres

package cmd

import (
	"context"

	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/postgres"
)

func init() {
	rootCmd.PersistentFlags().Bool("postgres.enabled", false, "enable PostgreSQL exporter")
	rootCmd.PersistentFlags().String("postgres.host", "", "database host or IP")
	rootCmd.PersistentFlags().Int("postgres.port", 0, "database port")
	rootCmd.PersistentFlags().String("postgres.username", "", "database username")
	rootCmd.PersistentFlags().String("postgres.password", "", "database password")
	rootCmd.PersistentFlags().String("postgres.database", "", "database name")
	rootCmd.PersistentFlags().String("postgres.table", "", "table name")
	rootCmd.PersistentFlags().String("postgres.sslmode", "", "SSL mode")
	rootCmd.PersistentFlags().String("postgres.sslcert", "", "path to SSL certificate file")
	rootCmd.PersistentFlags().String("postgres.sslkey", "", "path to SSL key file")

	viper.SetDefault("postgres.port", "5432")
	viper.SetDefault("postgres.sslmode", "disable")
}

func addPostgresExporter(exporters *[]exporter.Exporter) error {
	ctx := context.Background()
	psqlInfo, err := CreatePsqlInfoString("postgres")
	if err != nil {
		return err
	}
	logger.Info("Connecting to PostgreSQL", "connstr", SanitizePassword(psqlInfo))
	exp, err := postgres.New(ctx, psqlInfo, viper.GetString("postgres.table"))
	if err != nil {
		return err
	}
	*exporters = append(*exporters, exp)
	return nil
}
