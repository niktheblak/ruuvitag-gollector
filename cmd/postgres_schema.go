//go:build postgres

package cmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	_ "github.com/lib/pq"

	pexp "github.com/niktheblak/ruuvitag-gollector/pkg/exporter/postgres"
)

var postgresSchemaCmd = &cobra.Command{
	Use:   "postgres-schema",
	Short: "Create PostgreSQL schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn := viper.GetString("postgres.conn")
		table := viper.GetString("postgres.table")
		logger.Info("Creating schema", zap.String("conn", conn), zap.String("table", table))
		schema := fmt.Sprintf(pexp.SchemaTmpl, table)
		db, err := sql.Open("postgres", conn)
		if err != nil {
			return err
		}
		defer db.Close()
		_, err = db.ExecContext(cmd.Context(), schema)
		if err != nil {
			return err
		}
		_, err = db.ExecContext(cmd.Context(), fmt.Sprintf("CREATE INDEX idx_name ON %s(name)", table))
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(postgresSchemaCmd)
}
