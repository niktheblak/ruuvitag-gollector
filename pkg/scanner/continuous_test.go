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
	exp := new(mockExporter)
	device := mockDevice{}
	scn, err := NewContinuousWithOpts(Config{
		Exporters:     []exporter.Exporter{exp},
		DeviceName:    "default",
		BLEScanner:    NewMockBLEScanner(testAdvertisement),
		Peripherals:   peripherals,
		DeviceCreator: mockDeviceCreator{device},
		Logger:        logger,
	})
	require.NoError(t, err)
	defer scn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	require.NoError(t, err)
	errs := make(chan error, 1)
	go func() {
		if err := scn.Scan(ctx, 0); err != nil {
			errs <- err
		}
		close(errs)
	}()
	// Wait a bit for messages to appear in the measurements channel
	time.Sleep(2 * time.Second)
	cancel()
	require.NoError(t, <-errs)
	err = scn.Close()
	require.NoError(t, err)
	require.NotEmpty(t, exp.events)
	e := exp.events[0]
	assert.Equal(t, "Test", e.Name)
	assert.Equal(t, testAddr1, e.Addr)
	assert.Equal(t, 55.0, e.Temperature)
	assert.Equal(t, 60.0, e.Humidity)
	assert.Equal(t, 510.0, e.Pressure)
	assert.Equal(t, 500.0, e.BatteryVoltage)
}
