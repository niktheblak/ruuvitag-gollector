package wetbulb

import (
	"errors"
	"fmt"
	"math"

	"github.com/niktheblak/ruuvitag-gollector/pkg/temperature"
)

var ErrInvalidHumidity = errors.New("invalid humidity")

// Calculate returns wet bulb temperature in °C
func Calculate(temp float64, unit temperature.Unit, humidity float64) (float64, error) {
	if humidity < 0 || humidity > 100 {
		return 0, fmt.Errorf("%w: %v", ErrInvalidHumidity, humidity)
	}
	tempC := temperature.Convert(temp, unit, temperature.Celsius)
	t := tempC
	r := humidity
	tw := t*math.Atan(0.151977*math.Sqrt(r+8.313659)) +
		math.Atan(t+r) -
		math.Atan(r-1.676331) +
		0.00391838*math.Pow(r, 1.5)*math.Atan(0.023101*r) -
		4.686035
	return tw, nil
}
