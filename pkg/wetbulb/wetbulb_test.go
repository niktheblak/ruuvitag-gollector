package wetbulb

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
		pressure    float64
		want        float64
	}{
		{
			name:        "existing use case",
			temperature: 29,
			unit:        temperature.Celsius,
			humidity:    85,
			want:        26.8904,
		},
		{
			name:        "upper temperature bound",
			temperature: 80,
			unit:        temperature.Celsius,
			humidity:    50,
			want:        64.5294,
		},
		{
			name:        "lower atmospheric pressure",
			temperature: 80,
			unit:        temperature.Celsius,
			humidity:    50,
			pressure:    700,
			want:        64.2333,
		},
		{
			name:        "sub-freezing wet bulb",
			temperature: 0,
			unit:        temperature.Celsius,
			humidity:    50,
			want:        -2.8963,
		},
		{
			name:        "zero relative humidity",
			temperature: 25,
			unit:        temperature.Celsius,
			humidity:    0,
			want:        8.2617,
		},
		{
			name:        "saturated air",
			temperature: 80,
			unit:        temperature.Celsius,
			humidity:    100,
			want:        80,
		},
		{
			name:        "saturated over ice below freezing",
			temperature: -20,
			unit:        temperature.Celsius,
			humidity:    100,
			want:        -20.2967,
		},
		{
			name:        "fahrenheit input and output",
			temperature: 176,
			unit:        temperature.Fahrenheit,
			humidity:    50,
			want:        148.1529,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Calculate(test.temperature, test.unit, test.humidity, test.pressure)
			require.NoError(t, err)
			assert.InDelta(t, test.want, got, 0.002)
		})
	}
}

func TestCalculateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		temp     float64
		humidity float64
		pressure float64
		wantErr  error
	}{
		{name: "humidity below minimum", temp: 20, humidity: -1, wantErr: ErrInvalidHumidity},
		{name: "humidity above maximum", temp: 20, humidity: 101, wantErr: ErrInvalidHumidity},
		{name: "humidity NaN", temp: 20, humidity: math.NaN(), wantErr: ErrInvalidHumidity},
		{name: "temperature below minimum", temp: -20.1, humidity: 50, wantErr: ErrInvalidTemperature},
		{name: "temperature above maximum", temp: 80.1, humidity: 50, wantErr: ErrInvalidTemperature},
		{name: "temperature positive infinity", temp: math.Inf(1), humidity: 50, wantErr: ErrInvalidTemperature},
		{name: "pressure below minimum", temp: 20, humidity: 50, pressure: 499.99, wantErr: ErrInvalidPressure},
		{name: "pressure above maximum", temp: 20, humidity: 50, pressure: 1155.36, wantErr: ErrInvalidPressure},
		{name: "pressure NaN", temp: 20, humidity: 50, pressure: math.NaN(), wantErr: ErrInvalidPressure},
		{name: "pressure positive infinity", temp: 20, humidity: 50, pressure: math.Inf(1), wantErr: ErrInvalidPressure},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Calculate(test.temp, temperature.Celsius, test.humidity, test.pressure)
			assert.ErrorIs(t, err, test.wantErr)
		})
	}
}

func TestCalculateUsesStandardPressureWhenUnavailable(t *testing.T) {
	want, err := Calculate(80, temperature.Celsius, 50, StandardAtmosphericPressureHPA)
	require.NoError(t, err)

	for _, pressure := range []float64{0, -1} {
		got, err := Calculate(80, temperature.Celsius, 50, pressure)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	}
}

func TestCalculateAcrossSupportedRange(t *testing.T) {
	for _, pressure := range []float64{MinPressureHPA, 700, StandardAtmosphericPressureHPA, MaxPressureHPA} {
		for tempC := MinTemperatureC; tempC <= MaxTemperatureC; tempC += 5 {
			previousWetBulb := math.Inf(-1)
			for humidity := MinHumidity; humidity <= MaxHumidity; humidity += 5 {
				wetBulbC, err := Calculate(tempC, temperature.Celsius, humidity, pressure)
				require.NoError(t, err, "temperature=%v humidity=%v pressure=%v", tempC, humidity, pressure)
				assert.False(t, math.IsNaN(wetBulbC), "temperature=%v humidity=%v pressure=%v", tempC, humidity, pressure)
				assert.LessOrEqual(t, wetBulbC, tempC, "temperature=%v humidity=%v pressure=%v", tempC, humidity, pressure)
				assert.GreaterOrEqual(t, wetBulbC, previousWetBulb,
					"wet-bulb temperature should increase with humidity at temperature=%v pressure=%v", tempC, pressure)
				previousWetBulb = wetBulbC
			}
		}
	}
}
