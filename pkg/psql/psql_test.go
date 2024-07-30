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

var testData = sensor.Data{
	Addr:              "ec-40-93-94-35-2a",
	Name:              "Test",
	Temperature:       22.5,
	Humidity:          46,
	DewPoint:          12.1,
	Pressure:          1002,
	BatteryVoltage:    1.45,
	TxPower:           -18,
	AccelerationX:     772,
	AccelerationY:     -724,
	AccelerationZ:     -4,
	MovementCounter:   33,
	MeasurementNumber: 55526,
	Timestamp:         time.Date(2024, time.July, 3, 10, 24, 23, 213, time.UTC),
}

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
	psqlInfo, err := CreateConnString(vpr, "postgres")
	require.NoError(t, err)
	assert.Equal(
		t,
		"host=localhost port=5432 user=test_user password=t4st_p4s!$ dbname=test_database sslmode=disable",
		psqlInfo,
	)

	vpr.Set("postgres.table", "")
	_, err = CreateConnString(vpr, "postgres")
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
	AddFlags(fs, vpr, "postgres")
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

func TestDefaultBuildInsertQuery(t *testing.T) {
	q, err := BuildInsertQuery("ruuvitag", sensor.DefaultColumnMap)
	require.NoError(t, err)
	assert.Equal(t, `INSERT INTO ruuvitag(time,mac,name,temperature,humidity,pressure,acceleration_x,acceleration_y,acceleration_z,movement_counter,measurement_number,dew_point,battery_voltage,tx_power) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`, q)
}

func TestBuildInsertQuery(t *testing.T) {
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
	q, err := BuildInsertQuery("ruuvitag", columns)
	require.NoError(t, err)
	assert.Equal(t, `INSERT INTO ruuvitag(ts,addr,roomName,temperature,humidity,pressure,accX,accY,accZ,movementCounter,measurementNumber,dewPoint,batteryVoltage) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`, q)
}

func TestDefaultBuildQueryArguments(t *testing.T) {
	args := BuildQueryArguments(sensor.DefaultColumnMap, testData)
	require.Len(t, args, 14)
	assert.Equal(t, testData.Timestamp, args[0])
	assert.Equal(t, testData.Addr, args[1])
	assert.Equal(t, testData.Name, args[2])
	assert.Equal(t, testData.Temperature, args[3])
	assert.Equal(t, testData.Humidity, args[4])
	assert.Equal(t, testData.Pressure, args[5])
	assert.Equal(t, testData.AccelerationX, args[6])
	assert.Equal(t, testData.AccelerationY, args[7])
	assert.Equal(t, testData.AccelerationZ, args[8])
	assert.Equal(t, testData.MovementCounter, args[9])
	assert.Equal(t, testData.MeasurementNumber, args[10])
	assert.Equal(t, testData.DewPoint, args[11])
	assert.Equal(t, testData.BatteryVoltage, args[12])
	assert.Equal(t, testData.TxPower, args[13])
}

func TestBuildQueryArguments(t *testing.T) {
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
	args := BuildQueryArguments(columns, testData)
	require.Len(t, args, 10)
	assert.Equal(t, testData.Timestamp, args[0])
	assert.Equal(t, testData.Addr, args[1])
	assert.Equal(t, testData.Name, args[2])
	assert.Equal(t, testData.Temperature, args[3])
	assert.Equal(t, testData.Humidity, args[4])
	assert.Equal(t, testData.Pressure, args[5])
	assert.Equal(t, testData.MovementCounter, args[6])
	assert.Equal(t, testData.MeasurementNumber, args[7])
	assert.Equal(t, testData.DewPoint, args[8])
	assert.Equal(t, testData.BatteryVoltage, args[9])
}
