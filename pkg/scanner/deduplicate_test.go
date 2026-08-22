package scanner

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"

	commonsensor "github.com/niktheblak/ruuvitag-common/pkg/sensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

func TestMeasurementDeduplicator(t *testing.T) {
	t.Parallel()

	var deduplicator measurementDeduplicator
	assert.False(t, deduplicator.IsDuplicate(testAddr1, 42))
	assert.True(t, deduplicator.IsDuplicate(testAddr1, 42))
	assert.False(t, deduplicator.IsDuplicate(testAddr1, 43))
	assert.True(t, deduplicator.IsDuplicate(testAddr1, 43))
}

func TestMeasurementDeduplicatorTreatsRolloverAsNew(t *testing.T) {
	t.Parallel()

	var deduplicator measurementDeduplicator
	assert.False(t, deduplicator.IsDuplicate(testAddr1, ^uint16(0)))
	assert.True(t, deduplicator.IsDuplicate(testAddr1, ^uint16(0)))
	assert.False(t, deduplicator.IsDuplicate(testAddr1, 0))
	assert.True(t, deduplicator.IsDuplicate(testAddr1, 0))
}

func TestMeasurementDeduplicatorTracksDevicesIndependently(t *testing.T) {
	t.Parallel()

	var deduplicator measurementDeduplicator
	assert.False(t, deduplicator.IsDuplicate(testAddr1, 42))
	assert.False(t, deduplicator.IsDuplicate(testAddr2, 42))
	assert.True(t, deduplicator.IsDuplicate(testAddr1, 42))
	assert.True(t, deduplicator.IsDuplicate(testAddr2, 42))
}

func TestMeasurementsDeduplicateDataFormat5Advertisements(t *testing.T) {
	t.Parallel()

	measurements := collectMeasurements(t,
		format5Advertisement(t, testAddr1, ^uint16(0)),
		format5Advertisement(t, testAddr1, ^uint16(0)),
		format5Advertisement(t, testAddr1, 0),
		format5Advertisement(t, testAddr1, 0),
	)
	require.Len(t, measurements, 2)
	assert.Equal(t, int(^uint16(0)), measurements[0].MeasurementNumber)
	assert.Zero(t, measurements[1].MeasurementNumber)
}

func TestMeasurementsDoNotDeduplicateDataFormat3Advertisements(t *testing.T) {
	t.Parallel()

	measurements := collectMeasurements(t, testAdvertisement.advertisement(), testAdvertisement.advertisement())
	require.Len(t, measurements, 2)
}

func TestMeasurementsRetainDeduplicationStateAcrossScans(t *testing.T) {
	t.Parallel()

	measurement := format5Advertisement(t, testAddr1, 42)
	measurements := &Measurements{
		BLE:         batchBLEAdapter{measurement},
		Peripherals: map[string]string{testAddr1: "Test"},
		Logger:      logger,
	}
	require.Len(t, collectFromMeasurements(t, measurements), 1)

	measurements.BLE = batchBLEAdapter{measurement}
	require.Empty(t, collectFromMeasurements(t, measurements))

	measurements.BLE = batchBLEAdapter{format5Advertisement(t, testAddr1, 43)}
	require.Len(t, collectFromMeasurements(t, measurements), 1)
}

type batchBLEAdapter []Advertisement

func (a batchBLEAdapter) Scan(_ context.Context, handler AdvertisementHandler) error {
	for _, advertisement := range a {
		handler(advertisement)
	}
	return nil
}

func (batchBLEAdapter) Close() error { return nil }

func collectMeasurements(t *testing.T, advertisements ...Advertisement) []commonsensor.Data {
	t.Helper()
	measurements := &Measurements{
		BLE:         batchBLEAdapter(advertisements),
		Peripherals: map[string]string{testAddr1: "Test", testAddr2: "Test 2"},
		Logger:      logger,
	}
	return collectFromMeasurements(t, measurements)
}

func collectFromMeasurements(t *testing.T, measurements *Measurements) []commonsensor.Data {
	t.Helper()
	ch, done := measurements.Channel(context.Background())
	var result []commonsensor.Data
	for measurement := range ch {
		result = append(result, measurement)
	}
	require.NoError(t, <-done)
	return result
}

func format5Advertisement(t *testing.T, address string, measurementNumber uint16) Advertisement {
	t.Helper()
	data := sensor.DataFormat5{
		ManufacturerID:    0x9904,
		DataFormat:        sensor.DataFormat5ID,
		Temperature:       5000,
		Humidity:          20000,
		Pressure:          50000,
		MeasurementNumber: measurementNumber,
	}
	buf := new(bytes.Buffer)
	require.NoError(t, binary.Write(buf, binary.BigEndian, data))
	return mockAdvertisement{
		addr:             address,
		manufacturerData: buf.Bytes(),
	}.advertisement()
}
