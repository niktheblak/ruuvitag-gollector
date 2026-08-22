package dewpoint

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/niktheblak/ruuvitag-gollector/pkg/temperature"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name        string
		temperature float64
		unit        temperature.Unit
		humidity    float64
		want        float64
	}{
		{
			name:        "dew point over water",
			temperature: 20,
			unit:        temperature.Celsius,
			humidity:    50,
			want:        9.2728,
		},
		{
			name:        "second dew point reference",
			temperature: 30,
			unit:        temperature.Celsius,
			humidity:    60,
			want:        21.3877,
		},
		{
			name:        "frost point over ice",
			temperature: -20,
			unit:        temperature.Celsius,
			humidity:    50,
			want:        -27.0206,
		},
		{
			name:        "low humidity converges",
			temperature: 80,
			unit:        temperature.Celsius,
			humidity:    0.01,
			want:        -48.4812,
		},
		{
			name:        "fahrenheit input and output",
			temperature: 68,
			unit:        temperature.Fahrenheit,
			humidity:    50,
			want:        48.6910,
		},
		{
			name:        "kelvin input and output",
			temperature: 293.15,
			unit:        temperature.Kelvin,
			humidity:    50,
			want:        282.4228,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Calculate(test.temperature, test.unit, test.humidity)
			require.NoError(t, err)
			assert.InDelta(t, test.want, got, 0.002)
		})
	}
}

func TestCalculateAtSaturation(t *testing.T) {
	tests := []struct {
		temperature float64
		unit        temperature.Unit
	}{
		{temperature: -20, unit: temperature.Celsius},
		{temperature: 0, unit: temperature.Celsius},
		{temperature: 80, unit: temperature.Celsius},
		{temperature: 68, unit: temperature.Fahrenheit},
		{temperature: 293.15, unit: temperature.Kelvin},
	}

	for _, test := range tests {
		got, err := Calculate(test.temperature, test.unit, 100)
		require.NoError(t, err)
		assert.InDelta(t, test.temperature, got, 1e-12)
	}
}

func TestCalculateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		temperature float64
		unit        temperature.Unit
		humidity    float64
		wantErr     error
	}{
		{name: "temperature below minimum", temperature: -100.01, unit: temperature.Celsius, humidity: 100, wantErr: ErrInvalidTemperature},
		{name: "temperature above maximum", temperature: maxTemperatureK + 1, unit: temperature.Kelvin, humidity: 50, wantErr: ErrInvalidTemperature},
		{name: "temperature NaN", temperature: math.NaN(), unit: temperature.Celsius, humidity: 50, wantErr: ErrInvalidTemperature},
		{name: "temperature infinity", temperature: math.Inf(1), unit: temperature.Celsius, humidity: 50, wantErr: ErrInvalidTemperature},
		{name: "humidity zero", temperature: 20, unit: temperature.Celsius, humidity: 0, wantErr: ErrInvalidHumidity},
		{name: "humidity negative", temperature: 20, unit: temperature.Celsius, humidity: -1, wantErr: ErrInvalidHumidity},
		{name: "humidity above maximum", temperature: 20, unit: temperature.Celsius, humidity: 101, wantErr: ErrInvalidHumidity},
		{name: "humidity NaN", temperature: 20, unit: temperature.Celsius, humidity: math.NaN(), wantErr: ErrInvalidHumidity},
		{name: "invalid unit", temperature: 20, unit: temperature.Unit(99), humidity: 50, wantErr: ErrInvalidUnit},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Calculate(test.temperature, test.unit, test.humidity)
			assert.ErrorIs(t, err, test.wantErr)
			assert.Zero(t, got)
		})
	}
}

func TestCalculateRejectsResultBelowFormulaRange(t *testing.T) {
	got, err := Calculate(-40, temperature.Celsius, 0.01)
	assert.ErrorIs(t, err, ErrDewPointOutOfRange)
	assert.Zero(t, got)
}

func TestCalculateAcrossRuuviTemperatureRange(t *testing.T) {
	minimumPressure := saturationVaporPressure(minTemperatureK)
	humidities := []float64{0.01, 0.1, 1, 5, 25, 50, 75, 99, 100}

	for tempC := -40.0; tempC <= 80; tempC += 5 {
		for _, humidity := range humidities {
			tempK := temperature.Convert(tempC, temperature.Celsius, temperature.Kelvin)
			targetPressure := humidity / 100 * saturationVaporPressure(tempK)
			got, err := Calculate(tempC, temperature.Celsius, humidity)
			if targetPressure < minimumPressure {
				assert.ErrorIs(t, err, ErrDewPointOutOfRange,
					"temperature=%v humidity=%v", tempC, humidity)
				assert.Zero(t, got)
				continue
			}

			require.NoError(t, err, "temperature=%v humidity=%v", tempC, humidity)
			assert.False(t, math.IsNaN(got), "temperature=%v humidity=%v", tempC, humidity)
			assert.GreaterOrEqual(t, got, -100.0, "temperature=%v humidity=%v", tempC, humidity)
			assert.LessOrEqual(t, got, tempC, "temperature=%v humidity=%v", tempC, humidity)
		}
	}
}

func TestSaturationVaporPressureWaterIAPWSVerificationValues(t *testing.T) {
	// IAPWS-IF97 Region 4, table 35.
	tests := []struct {
		temperatureK float64
		pressurePa   float64
	}{
		{temperatureK: 300, pressurePa: 0.353658941e-2 * 1e6},
		{temperatureK: 500, pressurePa: 0.263889776e1 * 1e6},
		{temperatureK: 600, pressurePa: 0.123443146e2 * 1e6},
	}

	for _, test := range tests {
		got := saturationVaporPressureWater(test.temperatureK)
		assert.InEpsilon(t, test.pressurePa, got, 1e-8)
	}
}

func TestSaturationVaporPressureIceHardyVerificationValues(t *testing.T) {
	tests := []struct {
		temperatureK float64
		pressurePa   float64
	}{
		{temperatureK: 173.15, pressurePa: 0.0014018723},
		{temperatureK: 233.15, pressurePa: 12.836848},
		{temperatureK: 273.16, pressurePa: 611.6571},
	}

	for _, test := range tests {
		got := saturationVaporPressureIce(test.temperatureK)
		assert.InEpsilon(t, test.pressurePa, got, 1e-6)
	}
}

func TestSaturationVaporPressureIsContinuousAtTriplePoint(t *testing.T) {
	icePressure := saturationVaporPressureIce(triplePointK)
	waterPressure := saturationVaporPressureWater(triplePointK)
	assert.InDelta(t, icePressure, waterPressure, 0.001)
}
