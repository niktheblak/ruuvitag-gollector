//go:build postgres

package cmd

import (
	"context"
	"log/slog"

	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/postgres"
	"github.com/niktheblak/ruuvitag-gollector/pkg/psql"
)

func init() {
	rootCmd.PersistentFlags().Bool("postgres.enabled", false, "enable PostgreSQL exporter")
	psql.AddPsqlFlags(rootCmd.PersistentFlags(), "postgres")
}

func addPostgresExporter(exporters *[]exporter.Exporter) error {
	ctx := context.Background()
	psqlInfo, err := psql.CreatePsqlInfoString("postgres")
	if err != nil {
		return err
	}
	table := viper.GetString("postgres.table")
	timeColumn := viper.GetString("postgres.column.time")
	logger.LogAttrs(ctx, slog.LevelInfo, "Connecting to PostgreSQL", slog.String("conn_str", psql.RemovePassword(psqlInfo)), slog.String("table", table))
	exp, err := postgres.New(ctx, psqlInfo, table, timeColumn, logger)
	if err != nil {
		return err
	}
	*exporters = append(*exporters, exp)
	return nil
}
