//go:build postgres

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	createExtensionStmt = `CREATE EXTENSION IF NOT EXISTS timescaledb`

	createSchemaTemplate = `CREATE TABLE %s (
	%s TIMESTAMPTZ NOT NULL,
	mac MACADDR NOT NULL,
	name TEXT,
	temperature DOUBLE PRECISION,
	humidity DOUBLE PRECISION,
	pressure DOUBLE PRECISION,
	acceleration_x INTEGER,
	acceleration_y INTEGER,
	acceleration_z INTEGER,
	movement_counter INTEGER,
	battery DOUBLE PRECISION,
	measurement_number INTEGER,
	dew_point DOUBLE PRECISION,
	battery_voltage DOUBLE PRECISION,
	tx_power INTEGER)`

	createHyperTableTemplate = `SELECT create_hypertable('%s', by_range('time'))`
)

var postgresSchemaCmd = &cobra.Command{
	Use:   "postgres-schema",
	Short: "Print PostgreSQL schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			table      = viper.GetString("postgres.table")
			dbType     = viper.GetString("postgres.type")
			timeColumn = viper.GetString("postgres.column.time")
		)
		switch dbType {
		case "postgres":
			stmt := fmt.Sprintf(createSchemaTemplate, table, timeColumn)
			cmd.Printf("%s;\n", stmt)
		case "timescaledb":
			cmd.Printf("%s;\n", createExtensionStmt)
			createStmt := fmt.Sprintf(createSchemaTemplate, table, timeColumn)
			cmd.Printf("%s;\n", createStmt)
			hyperTableStmt := fmt.Sprintf(createHyperTableTemplate, table)
			cmd.Printf("%s;\n", hyperTableStmt)
		default:
			return fmt.Errorf("unknown database type: %s", dbType)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(postgresSchemaCmd)
}
