// +build postgres

package cmd

import (
	"context"
	"fmt"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/postgres"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.PersistentFlags().Bool("postgres.enabled", false, "Store measurements to PostgreSQL")
	rootCmd.PersistentFlags().String("postgres.conn", "", "PostgreSQL connection string")
	rootCmd.PersistentFlags().String("postgres.table", "", "PostgreSQL table")
}

func addPostgresExporter(exporters *[]exporter.Exporter) error {
	ctx := context.Background()
	connStr := viper.GetString("postgres.conn")
	table := viper.GetString("postgres.table")
	exp, err := postgres.New(ctx, connStr, table)
	if err != nil {
		return fmt.Errorf("failed to create PostgreSQL exporter: %w", err)
	}
	*exporters = append(*exporters, exp)
	return nil
}
