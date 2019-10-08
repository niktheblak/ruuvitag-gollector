package scanner

import (
	"bytes"
	"context"
	"encoding/binary"
	"io/ioutil"
	"log"
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

var logger = log.New(ioutil.Discard, "", log.LstdFlags)

func TestScanOnce(t *testing.T) {
	peripherals := map[string]string{
		"cc:ca:7e:52:cc:34": "Test",
	}
	scn, err := New(logger, "default", peripherals)
	require.NoError(t, err)
	exp := new(mockExporter)
	scn.Exporters = []exporter.Exporter{exp}
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, testData)
	require.NoError(t, err)
	device := mockDevice{}
	scn.ble = mockBLEScanner{
		manufacturerData: buf.Bytes(),
	}
	scn.dev = mockDeviceCreator{device: device}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = scn.ScanOnce(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, exp.events)
	e := exp.events[0]
	assert.Equal(t, "Test", e.Name)
	assert.Equal(t, "CC:CA:7E:52:CC:34", e.Addr)
	assert.Equal(t, 55.0, e.Temperature)
	assert.Equal(t, 60.0, e.Humidity)
	assert.Equal(t, 51000, e.Pressure)
	assert.Equal(t, 500, e.Battery)
}

func TestScan(t *testing.T) {
	peripherals := map[string]string{
		"cc:ca:7e:52:cc:34": "Test",
	}
	scn, err := New(logger, "default", peripherals)
	require.NoError(t, err)
	exp := new(mockExporter)
	scn.Exporters = []exporter.Exporter{exp}
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, testData)
	require.NoError(t, err)
	device := mockDevice{}
	scn.ble = mockBLEScanner{
		manufacturerData: buf.Bytes(),
	}
	scn.dev = mockDeviceCreator{device: device}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = scn.Scan(ctx)
	require.NoError(t, err)
	// Wait a bit for messages to appear in the measurements channel
	time.Sleep(100 * time.Millisecond)
	scn.Stop()
	assert.NotEmpty(t, exp.events)
	e := exp.events[0]
	assert.Equal(t, "Test", e.Name)
	assert.Equal(t, "cc:ca:7e:52:cc:34", e.Addr)
	assert.Equal(t, 55.0, e.Temperature)
	assert.Equal(t, 60.0, e.Humidity)
	assert.Equal(t, 51000, e.Pressure)
	assert.Equal(t, 500, e.Battery)
}
