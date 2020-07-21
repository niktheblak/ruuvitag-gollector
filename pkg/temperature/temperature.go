package temperature

type Unit int

const (
	Kelvin Unit = iota
	Celsius
	Fahrenheit
)

const CelsiusOffset = 273.15

func Convert(value float64, from, to Unit) float64 {
	switch from {
	case Kelvin:
		switch to {
		case Kelvin:
			return value
		case Celsius:
			return value - CelsiusOffset
		case Fahrenheit:
			return value*1.8 - 459.67
		}
	case Celsius:
		switch to {
		case Kelvin:
			return value + CelsiusOffset
		case Celsius:
			return value
		case Fahrenheit:
			return value*1.8 + 32.0
		}
	case Fahrenheit:
		switch to {
		case Kelvin:
			return (value + 459.67) / 1.8
		case Celsius:
			return (value - 32) / 1.8
		case Fahrenheit:
			return value
		}
	}
	return value
}
