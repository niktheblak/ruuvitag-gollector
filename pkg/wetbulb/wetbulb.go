package wetbulb

import (
	"errors"
	"fmt"
	"math"

	"github.com/niktheblak/ruuvitag-gollector/pkg/temperature"
)

const (
	MaxTemperatureC = 80.0
	MinTemperatureC = -20.0
	MaxHumidity     = 100.0
	MinHumidity     = 0.0
	// MaxPressureHPA and MinPressureHPA are the limits of the pressure value
	// encoded by the supported RuuviTag advertisement formats.
	MaxPressureHPA = 1155.35
	MinPressureHPA = 500.0

	// StandardAtmosphericPressureHPA is used when pressure is unavailable.
	StandardAtmosphericPressureHPA = 1013.25
	hectopascalToPascal            = 100.0
	wetBulbToleranceK              = 0.001
	triplePointWaterK              = 273.16
	triplePointPressurePa          = 611.65
	internalEnergyVapor            = 2.3740e6
	internalEnergyIce              = 0.3337e6
	gasConstantDryAir              = 287.04
	gasConstantWaterVapor          = 461.0
	specificHeatDryAirVolume       = 719.0
	specificHeatVaporVolume        = 1418.0
	specificHeatLiquidWater        = 4119.0
	specificHeatIce                = 1861.0
	specificHeatDryAirPressure     = specificHeatDryAirVolume + gasConstantDryAir
	specificHeatVaporPressure      = specificHeatVaporVolume + gasConstantWaterVapor
)

var (
	ErrInvalidHumidity    = errors.New("invalid humidity")
	ErrInvalidPressure    = errors.New("invalid pressure")
	ErrInvalidTemperature = errors.New("invalid temperature")
)

// Calculate returns the thermodynamic wet-bulb temperature in the provided
// temperature unit. Pressure is in hectopascals, matching the RuuviTag
// measurement data. A pressure less than or equal to zero means that no
// measurement is available and is replaced with standard atmospheric pressure
// (1013.25 hPa).
//
// The algorithm solves the Rankine-Kirchhoff thermodynamic wet-bulb balance
//
//	0 = c_pm (T_w - T) (1 - q_v*) + (q_v* - q_v) L_e(T_w)
//
// by bisection, using Rankine-Kirchhoff saturation vapor pressures and the
// paper's optimized thermodynamic constants. The derivation and validation are
// in David M. Romps, "Wet-Bulb Temperature from Pressure, Relative Humidity,
// and Air Temperature", Journal of Applied Meteorology and Climatology 65(2),
// 285-298 (2026): https://doi.org/10.1175/JAMC-D-25-0130.1
func Calculate(temp float64, unit temperature.Unit, humidity, pressureHPA float64) (float64, error) {
	if math.IsNaN(humidity) || math.IsInf(humidity, 0) || humidity < MinHumidity || humidity > MaxHumidity {
		return 0, fmt.Errorf("%w: %v", ErrInvalidHumidity, humidity)
	}
	if math.IsNaN(pressureHPA) || math.IsInf(pressureHPA, 0) {
		return 0, fmt.Errorf("%w: %v", ErrInvalidPressure, pressureHPA)
	}
	if pressureHPA <= 0 {
		pressureHPA = StandardAtmosphericPressureHPA
	}
	if pressureHPA < MinPressureHPA || pressureHPA > MaxPressureHPA {
		return 0, fmt.Errorf("%w: %v hPa", ErrInvalidPressure, pressureHPA)
	}
	if math.IsNaN(temp) || math.IsInf(temp, 0) {
		return 0, fmt.Errorf("%w: %v", ErrInvalidTemperature, temp)
	}
	tempC := temperature.Convert(temp, unit, temperature.Celsius)
	if tempC < MinTemperatureC || tempC > MaxTemperatureC {
		return 0, fmt.Errorf("%w: %v", ErrInvalidTemperature, temp)
	}

	pressurePa := pressureHPA * hectopascalToPascal
	dryBulbK := tempC + temperature.CelsiusOffset
	vaporPressure := humidity / 100 * saturationVaporPressure(dryBulbK)
	if vaporPressure >= pressurePa {
		return 0, fmt.Errorf("%w: %v hPa is not above vapor pressure", ErrInvalidPressure, pressureHPA)
	}
	vaporMassFraction := vaporMassFraction(vaporPressure, pressurePa)
	moistAirSpecificHeat := (1-vaporMassFraction)*specificHeatDryAirPressure +
		vaporMassFraction*specificHeatVaporPressure

	// The balance is monotonic over this interval. Absolute zero is a safe
	// mathematical lower bound; the dry-bulb temperature is the physical upper
	// bound for non-supersaturated air.
	lower := 0.0
	upper := dryBulbK
	balance := func(wetBulbK float64) float64 {
		saturatedMassFraction := saturatedVaporMassFractionLiquid(wetBulbK, pressurePa)
		return moistAirSpecificHeat*(wetBulbK-dryBulbK)*(1-saturatedMassFraction) +
			(saturatedMassFraction-vaporMassFraction)*latentHeatEvaporation(wetBulbK)
	}

	// This is exact for saturated air above the triple point. Below the triple
	// point, relative humidity is defined over ice while wet bulb is defined over
	// liquid water, so 100% RH does not generally imply T_w == T.
	if balance(upper) == 0 {
		return temperature.Convert(tempC, temperature.Celsius, unit), nil
	}

	for upper-lower > wetBulbToleranceK {
		wetBulbK := (lower + upper) / 2
		if balance(wetBulbK) > 0 {
			upper = wetBulbK
		} else {
			lower = wetBulbK
		}
	}

	wetBulbC := (lower+upper)/2 - temperature.CelsiusOffset
	return temperature.Convert(wetBulbC, temperature.Celsius, unit), nil
}

func vaporMassFraction(vaporPressure, pressure float64) float64 {
	return gasConstantDryAir * vaporPressure /
		(gasConstantWaterVapor*(pressure-vaporPressure) +
			gasConstantDryAir*vaporPressure)
}

func saturatedVaporMassFractionLiquid(tempK, pressure float64) float64 {
	return vaporMassFraction(saturationVaporPressureLiquid(tempK), pressure)
}

func latentHeatEvaporation(tempK float64) float64 {
	return internalEnergyVapor +
		(specificHeatVaporVolume-specificHeatLiquidWater)*(tempK-triplePointWaterK) +
		gasConstantWaterVapor*tempK
}

func saturationVaporPressure(tempK float64) float64 {
	if tempK < triplePointWaterK {
		return saturationVaporPressureIce(tempK)
	}
	return saturationVaporPressureLiquid(tempK)
}

func saturationVaporPressureLiquid(tempK float64) float64 {
	if tempK <= 0 {
		return 0
	}
	exponent := (internalEnergyVapor -
		(specificHeatVaporVolume-specificHeatLiquidWater)*triplePointWaterK) /
		gasConstantWaterVapor * (1/triplePointWaterK - 1/tempK)
	return triplePointPressurePa *
		math.Pow(tempK/triplePointWaterK,
			(specificHeatVaporPressure-specificHeatLiquidWater)/gasConstantWaterVapor) *
		math.Exp(exponent)
}

func saturationVaporPressureIce(tempK float64) float64 {
	if tempK <= 0 {
		return 0
	}
	exponent := (internalEnergyVapor + internalEnergyIce -
		(specificHeatVaporVolume-specificHeatIce)*triplePointWaterK) /
		gasConstantWaterVapor * (1/triplePointWaterK - 1/tempK)
	return triplePointPressurePa *
		math.Pow(tempK/triplePointWaterK,
			(specificHeatVaporPressure-specificHeatIce)/gasConstantWaterVapor) *
		math.Exp(exponent)
}
