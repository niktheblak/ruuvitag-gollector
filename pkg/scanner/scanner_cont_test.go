package scanner

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

func TestScanContinuously(t *testing.T) {
	scn := NewContinuous(logger, peripherals)
	defer scn.Close()
	exp := new(mockExporter)
	scn.Exporters = []exporter.Exporter{exp}
	device := mockDevice{}
	bleScanner := NewMockBLEScanner(testAdvertisement)
	scn.meas.BLE = bleScanner
	scn.dev = mockDeviceCreator{device: device}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := scn.Init("default")
	require.NoError(t, err)
	scn.Scan(ctx)
	// Wait a bit for messages to appear in the measurements channel
	time.Sleep(100 * time.Millisecond)
	scn.Stop()
	require.NotEmpty(t, exp.events)
	e := exp.events[0]
	assert.Equal(t, "Test", e.Name)
	assert.Equal(t, testAddr1, e.Addr)
	assert.Equal(t, 55.0, e.Temperature)
	assert.Equal(t, 60.0, e.Humidity)
	assert.Equal(t, 510.0, e.Pressure)
	assert.Equal(t, 500.0, e.BatteryVoltage)
}
