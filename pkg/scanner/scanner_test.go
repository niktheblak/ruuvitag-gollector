package scanner

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testData = sensor.DataFormat3{
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

const testAddr = "cc:ca:7e:52:cc:34"

var peripherals = map[string]string{
	testAddr: "Test",
}

var testAdvertisement mockAdvertisement

func init() {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, testData); err != nil {
		panic(err)
	}
	testAdvertisement = mockAdvertisement{
		localName:        "RuuviTag",
		addr:             testAddr,
		manufacturerData: buf.Bytes(),
	}
}

func TestScanOnce(t *testing.T) {
	var logger = NewTestLogger(t)
	scn := New(logger, "default", peripherals)
	exp := new(mockExporter)
	scn.Exporters = []exporter.Exporter{exp}
	device := mockDevice{}
	scn.ble = mockBLEScanner{
		advertisement: testAdvertisement,
	}
	scn.dev = mockDeviceCreator{device: device}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := scn.ScanOnce(ctx)
	require.NoError(t, err)
	// Wait a bit for messages to appear in the measurements channel
	time.Sleep(100 * time.Millisecond)
	assert.NotEmpty(t, exp.events)
	e := exp.events[0]
	assert.Equal(t, "Test", e.Name)
	assert.Equal(t, testAddr, e.Addr)
	assert.Equal(t, 55.0, e.Temperature)
	assert.Equal(t, 60.0, e.Humidity)
	assert.Equal(t, 510.0, e.Pressure)
	assert.Equal(t, 500, e.Battery)
}

func TestScanContinuously(t *testing.T) {
	var logger = NewTestLogger(t)
	scn := New(logger, "default", peripherals)
	defer scn.Close()
	exp := new(mockExporter)
	scn.Exporters = []exporter.Exporter{exp}
	device := mockDevice{}
	scn.ble = mockBLEScanner{
		advertisement: testAdvertisement,
	}
	scn.dev = mockDeviceCreator{device: device}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := scn.ScanContinuously(ctx)
	require.NoError(t, err)
	// Wait a bit for messages to appear in the measurements channel
	time.Sleep(100 * time.Millisecond)
	scn.Stop()
	assert.NotEmpty(t, exp.events)
	e := exp.events[0]
	assert.Equal(t, "Test", e.Name)
	assert.Equal(t, testAddr, e.Addr)
	assert.Equal(t, 55.0, e.Temperature)
	assert.Equal(t, 60.0, e.Humidity)
	assert.Equal(t, 510.0, e.Pressure)
	assert.Equal(t, 500, e.Battery)
}

func TestScanWithInterval(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	var logger = NewTestLogger(t)
	scn := New(logger, "default", peripherals)
	defer scn.Close()
	exp := new(mockExporter)
	scn.Exporters = []exporter.Exporter{exp}
	device := mockDevice{}
	scn.ble = mockBLEScanner{
		advertisement: testAdvertisement,
	}
	scn.dev = mockDeviceCreator{device: device}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := scn.ScanWithInterval(ctx, 1*time.Second)
	require.NoError(t, err)
	// Wait a bit for messages to appear in the measurements channel
	time.Sleep(2 * time.Second)
	scn.Stop()
	assert.NotEmpty(t, exp.events)
	e := exp.events[0]
	assert.Equal(t, "Test", e.Name)
	assert.Equal(t, testAddr, e.Addr)
	assert.Equal(t, 55.0, e.Temperature)
	assert.Equal(t, 60.0, e.Humidity)
	assert.Equal(t, 510.0, e.Pressure)
	assert.Equal(t, 500, e.Battery)
}
