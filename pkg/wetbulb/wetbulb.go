package wetbulb

import (
	"errors"
	"fmt"
	"math"

	"github.com/niktheblak/ruuvitag-gollector/pkg/temperature"
)

const (
	MaxTemperatureC = 50.0
	MinTemperatureC = -20.0
	MaxHumidity     = 99.0
	MinHumidity     = 5.0
)

var (
	ErrInvalidHumidity    = errors.New("invalid humidity")
	ErrInvalidTemperature = errors.New("invalid temperature")
)

// Calculate returns wet bulb temperature using the Roland Stull empirical wet-bulb approximation in the provided units.
func Calculate(temp float64, unit temperature.Unit, humidity float64) (float64, error) {
	if math.IsNaN(humidity) || math.IsInf(humidity, 0) || humidity < MinHumidity || humidity > MaxHumidity {
		return 0, fmt.Errorf("%w: %v", ErrInvalidHumidity, humidity)
	}
	if math.IsNaN(temp) || math.IsInf(temp, 0) {
		return 0, fmt.Errorf("%w: %v", ErrInvalidTemperature, temp)
	}
	tempC := temperature.Convert(temp, unit, temperature.Celsius)
	if tempC < MinTemperatureC || tempC > MaxTemperatureC {
		return 0, fmt.Errorf("%w: %v", ErrInvalidTemperature, temp)
	}
	t := tempC
	r := humidity
	tw := t*math.Atan(0.151977*math.Sqrt(r+8.313659)) +
		math.Atan(t+r) -
		math.Atan(r-1.676331) +
		0.00391838*math.Pow(r, 1.5)*math.Atan(0.023101*r) -
		4.686035
	return temperature.Convert(tw, temperature.Celsius, unit), nil
}
