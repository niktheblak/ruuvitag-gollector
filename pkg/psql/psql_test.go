package psql

import (
	"testing"
	"time"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
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

func TestRenderInsertQuery(t *testing.T) {
	columns := map[string]string{
		"time":               "ts",
		"mac":                "addr",
		"name":               "roomName",
		"temperature":        "temperature",
		"humidity":           "humidity",
		"pressure":           "pressure",
		"acceleration_x":     "accX",
		"acceleration_y":     "accY",
		"acceleration_z":     "accZ",
		"movement_counter":   "movementCounter",
		"measurement_number": "measurementNumber",
		"dew_point":          "dewPoint",
		"battery_voltage":    "batteryVoltage",
	}
	q, err := RenderInsertQuery("ruuvitag", columns)
	require.NoError(t, err)
	assert.Equal(t, `INSERT INTO ruuvitag(ts,addr,roomName,temperature,humidity,pressure,accX,accY,accZ,movementCounter,measurementNumber,dewPoint,batteryVoltage) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`, q)
}

func TestBuildQuery(t *testing.T) {
	columns := map[string]string{
		"time":               "ts",
		"mac":                "mac",
		"name":               "name",
		"temperature":        "temperature",
		"humidity":           "humidity",
		"pressure":           "pressure",
		"movement_counter":   "movementCounter",
		"measurement_number": "measurementNumber",
		"dew_point":          "dewPoint",
		"battery_voltage":    "batteryVoltage",
	}
	ts := time.Now()
	data := sensor.Data{
		Addr:              "ec-40-93-94-35-2a",
		Name:              "Test",
		Temperature:       22.5,
		Humidity:          46,
		DewPoint:          12.1,
		Pressure:          1002,
		BatteryVoltage:    1.45,
		TxPower:           0,
		AccelerationX:     0,
		AccelerationY:     0,
		AccelerationZ:     0,
		MovementCounter:   11,
		MeasurementNumber: 111,
		Timestamp:         ts,
	}
	args := BuildQuery(columns, data)
	require.Len(t, args, 10)
	assert.Equal(t, data.Timestamp, args[0])
	assert.Equal(t, data.Addr, args[1])
	assert.Equal(t, data.Name, args[2])
	assert.Equal(t, data.Temperature, args[3])
	assert.Equal(t, data.Humidity, args[4])
	assert.Equal(t, data.Pressure, args[5])
	assert.Equal(t, data.MovementCounter, args[6])
	assert.Equal(t, data.MeasurementNumber, args[7])
	assert.Equal(t, data.DewPoint, args[8])
	assert.Equal(t, data.BatteryVoltage, args[9])
}
