package dewpoint

import (
	"errors"
	"fmt"
	"math"

	"github.com/niktheblak/ruuvitag-gollector/pkg/temperature"
)

const (
	minTemperatureK  = 173.15
	maxTemperatureK  = 647.096
	triplePointK     = 273.16
	solverToleranceK = 0.001
)

// Water saturation vapor pressure coefficients
const (
	n1  = 0.11670521452767e4
	n2  = -0.72421316703206e6
	n3  = -0.17073846940092e2
	n4  = 0.12020824702470e5
	n5  = -0.32325550322333e7
	n6  = 0.14915108613530e2
	n7  = -0.48232657361591e4
	n8  = 0.40511340542057e6
	n9  = -0.23855557567849
	n10 = 0.65017534844798e3
)

// Ice saturation vapor pressure coefficients
const (
	k0 = -5.8666426e3
	k1 = 2.232870244e1
	k2 = 1.39387003e-2
	k3 = -3.4262402e-5
	k4 = 2.7040955e-8
	k5 = 6.7063522e-1
)

var (
	ErrDewPointOutOfRange = errors.New("dew point out of range")
	ErrInvalidHumidity    = errors.New("invalid humidity")
	ErrInvalidTemperature = errors.New("invalid temperature")
	ErrInvalidUnit        = errors.New("invalid temperature unit")
)

// Calculate returns the equilibrium condensation temperature for the given air
// temperature and relative humidity in percent. It returns liquid-water dew
// point at and above the water triple point (273.16 K), and ice frost point
// below it. The result uses the same temperature unit as temp.
//
// Saturation pressure over liquid water follows IAPWS-IF97 Region 4. Saturation
// pressure over ice follows Hardy's ITS-90 formulation, whose published lower
// validity limit is -100 C. ErrDewPointOutOfRange is returned when the supplied
// conditions imply a result below that limit.
func Calculate(temp float64, unit temperature.Unit, humidity float64) (float64, error) {
	if math.IsNaN(humidity) || math.IsInf(humidity, 0) || humidity <= 0 || humidity > 100 {
		return 0, fmt.Errorf("%w: %v", ErrInvalidHumidity, humidity)
	}
	if !validUnit(unit) {
		return 0, fmt.Errorf("%w: %v", ErrInvalidUnit, unit)
	}
	if math.IsNaN(temp) || math.IsInf(temp, 0) {
		return 0, fmt.Errorf("%w: %v", ErrInvalidTemperature, temp)
	}
	tempInK := temperature.Convert(temp, unit, temperature.Kelvin)
	if tempInK < minTemperatureK || tempInK > maxTemperatureK {
		return 0, fmt.Errorf("%w: temperature %f %v out of range", ErrInvalidTemperature, temp, unit)
	}

	targetPressure := humidity / 100 * saturationVaporPressure(tempInK)
	dpInK, err := solveSaturationTemperature(targetPressure, tempInK)
	if err != nil {
		return 0, err
	}
	return temperature.Convert(dpInK, temperature.Kelvin, unit), nil
}

func validUnit(unit temperature.Unit) bool {
	switch unit {
	case temperature.Kelvin, temperature.Celsius, temperature.Fahrenheit:
		return true
	default:
		return false
	}
}

func solveSaturationTemperature(targetPressure, upperTemperatureK float64) (float64, error) {
	minimumPressure := saturationVaporPressure(minTemperatureK)
	if targetPressure < minimumPressure {
		return 0, fmt.Errorf("%w: vapor pressure %g Pa is below minimum %g Pa",
			ErrDewPointOutOfRange, targetPressure, minimumPressure)
	}
	if targetPressure == saturationVaporPressure(upperTemperatureK) {
		return upperTemperatureK, nil
	}

	targetLogPressure := math.Log(targetPressure)
	lowerTemperatureK := minTemperatureK
	for upperTemperatureK-lowerTemperatureK > solverToleranceK {
		midpointK := (lowerTemperatureK + upperTemperatureK) / 2
		if math.Log(saturationVaporPressure(midpointK)) < targetLogPressure {
			lowerTemperatureK = midpointK
		} else {
			upperTemperatureK = midpointK
		}
	}
	return (lowerTemperatureK + upperTemperatureK) / 2, nil
}

func saturationVaporPressure(tempInK float64) float64 {
	if tempInK < triplePointK {
		return saturationVaporPressureIce(tempInK)
	}
	return saturationVaporPressureWater(tempInK)
}

// Saturation pressure over liquid water:
// IAPWS-IF97, Region 4 saturation-pressure equation.
// https://www.iapws.org/relguide/IF97-Rev.pdf
func saturationVaporPressureWater(tempInK float64) float64 {
	th := tempInK + n9/(tempInK-n10)
	a := (th+n1)*th + n2
	b := (n3*th+n4)*th + n5
	c := (n6*th+n7)*th + n8

	p := 2 * c / (-b + math.Sqrt(b*b-4*a*c))
	p *= p
	p *= p
	return p * 1e6
}

// Saturation pressure over ice:
// Hardy ITS-90 formulation for saturation vapour pressure over ice.
// https://www.rhs.com/wp-content/uploads/2021/03/its90form.pdf
func saturationVaporPressureIce(tempInK float64) float64 {
	lnP := k0/tempInK + k1 + (k2+(k3+k4*tempInK)*tempInK)*tempInK + k5*math.Log(tempInK)
	return math.Exp(lnP)
}
