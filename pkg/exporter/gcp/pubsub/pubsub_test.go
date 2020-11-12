// +build gcp
// +build integration

package pubsub

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

func TestPublish(t *testing.T) {
	project := os.Getenv("RUUVITAG_GOOGLE_PROJECT")
	topic := os.Getenv("RUUVITAG_PUBSUB_TOPIC")
	ctx := context.Background()
	e, err := New(ctx, project, topic)
	require.NoError(t, err)
	defer e.Close()
	err = e.Export(ctx, sensor.Data{
		Addr:           "CC:CA:7E:52:CC:34",
		Name:           "TestRuuviTag",
		Temperature:    20.1,
		Humidity:       65,
		Pressure:       1001,
		BatteryVoltage: 50,
		AccelerationX:  0,
		AccelerationY:  0,
		AccelerationZ:  0,
		Timestamp:      time.Now(),
	})
	require.NoError(t, err)
}
