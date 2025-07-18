package scanner

import (
	"bytes"
	"encoding/binary"
	"io"
	"log/slog"
	"testing"

	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
	"github.com/stretchr/testify/assert"
)

const (
	testAddr1 = "cc:ca:7e:52:cc:34"
	testAddr2 = "fb:e1:b7:04:95:ee"
	testAddr3 = "e8:e0:c6:0b:b8:c5"
)

var (
	testData = sensor.DataFormat3{
		ManufacturerID:      0x9904,
		DataFormat:          3,
		Humidity:            120,
		Temperature:         55,
		TemperatureFraction: 0,
		Pressure:            1000,
		AccelerationX:       0,
		AccelerationY:       0,
		AccelerationZ:       0,
		BatteryVoltageMv:    500,
	}
	peripherals = map[string]string{
		testAddr1: "Test",
	}
	testAdvertisement mockAdvertisement
	logger            *slog.Logger
)

func init() {
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, testData); err != nil {
		panic(err)
	}
	testAdvertisement = mockAdvertisement{
		addr:             testAddr1,
		manufacturerData: buf.Bytes(),
	}
}

func TestContainsKeys(t *testing.T) {
	t.Parallel()

	assert.True(t, containsKeys(map[string]string{
		testAddr1: "Living Room",
		testAddr2: "Bedroom",
		testAddr3: "Bathroom",
	}, map[string]bool{
		testAddr1: true,
		testAddr2: true,
		testAddr3: true,
	}))
	assert.False(t, containsKeys(map[string]string{
		testAddr1: "Living Room",
		testAddr2: "Bedroom",
		testAddr3: "Bathroom",
	}, map[string]bool{
		testAddr1: true,
		testAddr2: true,
	}))
}
