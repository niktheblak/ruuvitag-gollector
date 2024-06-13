package psql

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemovePassword(t *testing.T) {
	assert.Equal(
		t,
		"host=localhost port=5432 username=postgres password=[redacted] sslmode=disable",
		RemovePassword("host=localhost port=5432 username=postgres password=t4st_p4s!$ sslmode=disable"),
	)
}

func TestCreatePsqlInfoString(t *testing.T) {
	vpr := viper.New()
	vpr.Set("postgres.host", "localhost")
	vpr.Set("postgres.port", "5432")
	vpr.Set("postgres.username", "test_user")
	vpr.Set("postgres.password", "t4st_p4s!$")
	vpr.Set("postgres.database", "test_database")
	vpr.Set("postgres.table", "test_table")
	psqlInfo, err := CreatePsqlInfoString(vpr, "postgres")
	require.NoError(t, err)
	assert.Equal(
		t,
		"host=localhost port=5432 user=test_user password=t4st_p4s!$ dbname=test_database sslmode=disable",
		psqlInfo,
	)

	vpr.Set("postgres.table", "")
	_, err = CreatePsqlInfoString(vpr, "postgres")
	assert.ErrorContains(t, err, "table name must be specified")
}

func TestTrimQuery(t *testing.T) {
	trimmed := TrimQuery(`
		SELECT * FROM ruuvitag
		WHERE name = "Living Room"
		ORDER BY time DESC
		LIMIT 1`)
	assert.Equal(t, `SELECT * FROM ruuvitag WHERE name = "Living Room" ORDER BY time DESC LIMIT 1`, trimmed)
}

func TestAddPsqlFlags(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	vpr := viper.New()
	vpr.Set("postgres.host", "localhost")
	vpr.Set("postgres.port", "5432")
	vpr.Set("postgres.username", "test_user")
	vpr.Set("postgres.password", "t4st_p4s!$")
	vpr.Set("postgres.database", "test_database")
	vpr.Set("postgres.table", "test_table")
	AddPsqlFlags(fs, vpr, "postgres")
	f := fs.Lookup("postgres.host")
	require.NotNil(t, f)
	f = fs.Lookup("postgres.port")
	require.NotNil(t, f)
	f = fs.Lookup("postgres.username")
	require.NotNil(t, f)
	f = fs.Lookup("postgres.password")
	require.NotNil(t, f)
	f = fs.Lookup("postgres.database")
	require.NotNil(t, f)
	f = fs.Lookup("postgres.table")
	require.NotNil(t, f)
}
