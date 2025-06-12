package scanner

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanWithInterval(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	peripherals := map[string]string{
		testAddr1: "Backyard",
		testAddr2: "Upstairs",
		testAddr3: "Downstairs",
	}
	exp := new(mockExporter)
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, testData); err != nil {
		t.Fatal(err)
	}
	bleScanner := NewMockBLEScanner(
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
	device := mockDevice{}
	scn, err := NewInterval(Config{
		Exporters:     []exporter.Exporter{exp},
		DeviceName:    "default",
		BLEScanner:    bleScanner,
		Peripherals:   peripherals,
		DeviceCreator: mockDeviceCreator{device: device},
		Logger:        logger,
	})
	require.NoError(t, err)
	defer func() {
		err := scn.Close()
		require.NoError(t, err)
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	errs := make(chan error, 1)
	go func() {
		if err := scn.Scan(ctx, 100*time.Millisecond); err != nil {
			errs <- err
		}
		close(errs)
	}()
	// Wait a bit for messages to appear in the measurements channel
	time.Sleep(500 * time.Millisecond)
	cancel()
	require.NoError(t, <-errs)
	require.Len(t, exp.events, 3)
	e := exp.events[0]
	assert.Equal(t, "Backyard", e.Name)
	assert.Equal(t, testAddr1, e.Addr)
	assert.Equal(t, 55.0, e.Temperature)
	assert.Equal(t, 60.0, e.Humidity)
	assert.Equal(t, 510.0, e.Pressure)
	assert.Equal(t, 500.0, e.BatteryVoltage)
}
