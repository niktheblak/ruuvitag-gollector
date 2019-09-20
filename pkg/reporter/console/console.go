package console

import (
	"fmt"

	"github.com/niktheblak/ruuvitag-gollector/pkg/ruuvitag"
)

type Reporter struct {
}

func (r Reporter) Name() string {
	return "Console"
}

func (r Reporter) Report(data ruuvitag.SensorData) error {
	fmt.Println(data)
	return nil
}

func (r Reporter) Close() error {
	return nil
}
