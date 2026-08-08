package dewpoint

import (
	"errors"
	"fmt"
	"math"

	"github.com/niktheblak/ruuvitag-gollector/pkg/temperature"
)

// Temperature constants
const (
	MinTemperature = 173.0
	MaxTemperature = 647.096
)

// Water saturation vapor pressure coefficients
const (
	N1  = 0.11670521452767e4
	N6  = 0.14915108613530e2
	N2  = -0.72421316703206e6
	N7  = -0.48232657361591e4
	N3  = -0.17073846940092e2
	N8  = 0.40511340542057e6
	N4  = 0.12020824702470e5
	N9  = -0.23855557567849
	N5  = -0.32325550322333e7
	N10 = 0.65017534844798e3
)

// Ice saturation vapor pressure coefficients
const (
	K0 = -5.8666426e3
	K1 = 2.232870244e1
	K2 = 1.39387003e-2
	K3 = -3.4262402e-5
	K4 = 2.7040955e-8
	K5 = 6.7063522e-1
)

var (
	ErrInvalidHumidity    = errors.New("invalid humidity")
	ErrInvalidTemperature = errors.New("invalid temperature")
)

// Calculate calculates dew point from the given temperature and relative humidity (percent)
func Calculate(temp float64, unit temperature.Unit, humidity float64) (float64, error) {
	if math.IsNaN(humidity) || math.IsInf(humidity, 0) || humidity <= 0 || humidity > 100 {
		return 0, fmt.Errorf("%w: %v", ErrInvalidHumidity, humidity)
	}
	if math.IsNaN(temp) || math.IsInf(temp, 0) {
		return 0, fmt.Errorf("%w: %v", ErrInvalidTemperature, temp)
	}
	tempInK := temperature.Convert(temp, unit, temperature.Kelvin)
	if tempInK < MinTemperature || tempInK > MaxTemperature {
		return 0, fmt.Errorf("%w: temperature %f %v out of range", ErrInvalidTemperature, temp, unit)
	}
	dpInK, err := Solve(pvs, humidity/100.0*pvs(tempInK), tempInK)
	return temperature.Convert(dpInK, temperature.Kelvin, unit), err
}

func pvs(tempInK float64) float64 {
	if tempInK < temperature.CelsiusOffset {
		return pvsIce(tempInK)
	}
	return pvsWater(tempInK)
}

// Saturation pressure over liquid water:
// IAPWS-IF97, Region 4 saturation-pressure equation.
func pvsWater(tempInK float64) float64 {
	th := tempInK + N9/(tempInK-N10)
	a := (th+N1)*th + N2
	b := (N3*th+N4)*th + N5
	c := (N6*th+N7)*th + N8

	p := 2 * c / (-b + math.Sqrt(b*b-4*a*c))
	p *= p
	p *= p
	return p * 1e6
}

// Saturation pressure over ice:
// Hardy ITS-90 formulation for saturation vapour pressure over ice.
func pvsIce(tempInK float64) float64 {
	lnP := K0/tempInK + K1 + (K2+(K3+(K4*tempInK))*tempInK)*tempInK + K5*math.Log(tempInK)
	return math.Exp(lnP)
}
