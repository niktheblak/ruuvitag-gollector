//go:build postgres

package cmd

import (
	"context"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	pexp "github.com/niktheblak/ruuvitag-gollector/pkg/exporter/postgres"
	"github.com/niktheblak/ruuvitag-gollector/pkg/psql"
)

var postgresSchemaCmd = &cobra.Command{
	Use:   "postgres-schema",
	Short: "Create PostgreSQL schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		psqlInfo, err := CreatePsqlInfoString("postgres")
		if err != nil {
			return err
		}
		table := viper.GetString("postgres.table")
		logger.LogAttrs(ctx, slog.LevelInfo, "Connecting to PostgreSQL", slog.String("conn_str", psql.RemovePassword(psqlInfo)), slog.String("table", table))
		exp, err := pexp.New(ctx, psqlInfo, table, logger)
		if err != nil {
			return err
		}
		defer exp.Close()
		creator, ok := exp.(exporter.SchemaCreator)
		if !ok {
			return fmt.Errorf("exporter does not support schema creation")
		}
		return creator.InitSchema(ctx)
	},
}

func init() {
	rootCmd.AddCommand(postgresSchemaCmd)
}
