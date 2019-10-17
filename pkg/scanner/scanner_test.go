package scanner

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	logger            *logging.Logger
)

func init() {
	logging.InitForTesting(logging.CRITICAL)
	logger = logging.MustGetLogger("ruuvitag-gollector-test")
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, testData); err != nil {
		panic(err)
	}
	testAdvertisement = mockAdvertisement{
		addr:             testAddr1,
		manufacturerData: buf.Bytes(),
	}
}

func TestScanOnce(t *testing.T) {
	scn := New(logger, peripherals)
	exp := new(mockExporter)
	scn.Exporters = []exporter.Exporter{exp}
	device := mockDevice{}
	scn.ble = NewMockBLEScanner(testAdvertisement)
	scn.dev = mockDeviceCreator{device: device}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := scn.Init("default")
	require.NoError(t, err)
	err = scn.ScanOnce(ctx)
	require.NoError(t, err)
	// Wait a bit for messages to appear in the measurements channel
	time.Sleep(100 * time.Millisecond)
	assert.NotEmpty(t, exp.events)
	e := exp.events[0]
	assert.Equal(t, "Test", e.Name)
	assert.Equal(t, testAddr1, e.Addr)
	assert.Equal(t, 55.0, e.Temperature)
	assert.Equal(t, 60.0, e.Humidity)
	assert.Equal(t, 510.0, e.Pressure)
	assert.Equal(t, 500, e.Battery)
}

func TestScanContinuously(t *testing.T) {
	scn := New(logger, peripherals)
	defer scn.Close()
	exp := new(mockExporter)
	scn.Exporters = []exporter.Exporter{exp}
	device := mockDevice{}
	scn.ble = NewMockBLEScanner(testAdvertisement)
	scn.dev = mockDeviceCreator{device: device}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := scn.Init("default")
	require.NoError(t, err)
	err = scn.ScanContinuously(ctx)
	require.NoError(t, err)
	// Wait a bit for messages to appear in the measurements channel
	time.Sleep(100 * time.Millisecond)
	scn.Stop()
	assert.NotEmpty(t, exp.events)
	e := exp.events[0]
	assert.Equal(t, "Test", e.Name)
	assert.Equal(t, testAddr1, e.Addr)
	assert.Equal(t, 55.0, e.Temperature)
	assert.Equal(t, 60.0, e.Humidity)
	assert.Equal(t, 510.0, e.Pressure)
	assert.Equal(t, 500, e.Battery)
}

func TestScanWithInterval(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	peripherals := map[string]string{
		testAddr1: "Backyard",
		testAddr2: "Upstairs",
		testAddr3: "Downstairs",
	}
	scn := New(logger, peripherals)
	defer scn.Close()
	exp := new(mockExporter)
	scn.Exporters = []exporter.Exporter{exp}
	device := mockDevice{}
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, testData); err != nil {
		panic(err)
	}
	scn.ble = NewMockBLEScanner(
		mockAdvertisement{
			addr:             testAddr1,
			manufacturerData: buf.Bytes(),
		},
		mockAdvertisement{
			addr:             testAddr2,
			manufacturerData: buf.Bytes(),
		},
		mockAdvertisement{
			addr:             testAddr3,
			manufacturerData: buf.Bytes(),
		},
	)
	scn.dev = mockDeviceCreator{device: device}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := scn.Init("default")
	require.NoError(t, err)
	err = scn.ScanWithInterval(ctx, 100*time.Millisecond)
	require.NoError(t, err)
	// Wait a bit for messages to appear in the measurements channel
	time.Sleep(2 * time.Second)
	scn.Stop()
	require.Len(t, exp.events, 3)
	e := exp.events[0]
	assert.Equal(t, "Backyard", e.Name)
	assert.Equal(t, testAddr1, e.Addr)
	assert.Equal(t, 55.0, e.Temperature)
	assert.Equal(t, 60.0, e.Humidity)
	assert.Equal(t, 510.0, e.Pressure)
	assert.Equal(t, 500, e.Battery)
}
