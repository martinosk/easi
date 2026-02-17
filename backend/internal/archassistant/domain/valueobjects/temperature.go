package valueobjects

import "errors"

var ErrTemperatureOutOfRange = errors.New("temperature must be between 0.0 and 2.0")

const DefaultTemperature = 0.3

type Temperature struct {
	value float64
}

func NewTemperature(value float64) (Temperature, error) {
	if value < 0.0 || value > 2.0 {
		return Temperature{}, ErrTemperatureOutOfRange
	}
	return Temperature{value: value}, nil
}

func DefaultTemperatureValue() Temperature {
	return Temperature{value: DefaultTemperature}
}

func ReconstructTemperature(value float64) Temperature {
	return Temperature{value: value}
}

func (t Temperature) Value() float64 { return t.value }
