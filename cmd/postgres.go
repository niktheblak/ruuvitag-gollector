//go:build postgres

package cmd

import (
	"context"
	"log/slog"
	"time"

	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/postgres"
	"github.com/niktheblak/ruuvitag-gollector/pkg/psql"
)

func init() {
	rootCmd.PersistentFlags().Bool("postgres.enabled", false, "enable PostgreSQL exporter")
	psql.AddFlags(rootCmd.PersistentFlags(), viper.GetViper(), "postgres")
}

func addPostgresExporter(exporters *[]exporter.Exporter, columns map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	psqlInfo, err := psql.CreateConnString(viper.GetViper(), "postgres")
	if err != nil {
		return err
	}
	table := viper.GetString("postgres.table")
	logger.LogAttrs(ctx, slog.LevelInfo, "Connecting to PostgreSQL", slog.String("conn_str", psql.RemovePassword(psqlInfo)), slog.String("table", table))
	exp, err := postgres.New(ctx, postgres.Config{
		ConnString: psqlInfo,
		Table:      table,
		Columns:    columns,
		Logger:     logger,
	})
	if err != nil {
		return err
	}
	*exporters = append(*exporters, exp)
	return nil
}
