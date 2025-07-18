package wetbulb

import (
	"errors"
	"math"
)

var (
	ErrInvalidHumidity = errors.New("invalid humidity")
)

// Calculate returns wet bulb temperature in °C
func Calculate(tempC, rh float64) (float64, error) {
	if rh < 0 || rh > 100 {
		return 0, ErrInvalidHumidity
	}
	t := tempC
	r := rh
	tw := t*math.Atan(0.151977*math.Sqrt(r+8.313659)) +
		math.Atan(t+r) -
		math.Atan(r-1.676331) +
		0.00391838*math.Pow(r, 1.5)*math.Atan(0.023101*r) -
		4.686035
	return tw, nil
}
