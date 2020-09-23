// +build influxdb

package influxdb

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReport(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, ",mac=CC:CA:7E:52:CC:34,name=Backyard acceleration_x=0i,acceleration_y=0i,acceleration_z=0i,battery=50i,dew_point=9.6,humidity=45,pressure=1002,temperature=22.1 1569924000000000000\n", string(msg))
		w.WriteHeader(http.StatusNoContent)
		_, err = w.Write([]byte(""))
		require.NoError(t, err)
	}))
	defer srv.Close()
	exporter, err := New(Config{
		Addr: srv.URL,
	})
	require.NoError(t, err)
	ctx := context.Background()
	err = exporter.Export(ctx, sensor.Data{
		Addr:          "CC:CA:7E:52:CC:34",
		Name:          "Backyard",
		Temperature:   22.1,
		Humidity:      45.0,
		DewPoint:      9.6,
		Pressure:      1002.0,
		Battery:       50,
		AccelerationX: 0,
		AccelerationY: 0,
		AccelerationZ: 0,
		Timestamp:     time.Date(2019, time.October, 1, 10, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
}
