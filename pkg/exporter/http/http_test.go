package http

import (
	"context"
	"encoding/json"
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
)

var testData = sensor.Data{
	Addr:        "CC:CA:7E:52:CC:34",
	Name:        "Backyard",
	Temperature: 21.5,
	Humidity:    60,
	Pressure:    1002,
	Timestamp:   time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
}

type received struct {
	headers nethttp.Header
	body    map[string]any
}

func newTestServer(t *testing.T, status int, rec *received) *httptest.Server {
	t.Helper()
	return httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		assert.Equal(t, nethttp.MethodPost, r.Method)
		rec.headers = r.Header
		content, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.NoError(t, json.Unmarshal(content, &rec.body)) {
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte("response body"))
	}))
}

func TestExport(t *testing.T) {
	rec := new(received)
	srv := newTestServer(t, nethttp.StatusOK, rec)
	defer srv.Close()
	columns := map[string]string{
		"time":        "ts",
		"mac":         "addr",
		"name":        "name",
		"temperature": "temp",
	}
	exp, err := New(context.Background(), "test", Config{URL: srv.URL, Token: "secret", Columns: columns})
	require.NoError(t, err)
	assert.Equal(t, "test", exp.Name())
	require.NoError(t, exp.Export(context.Background(), testData))
	require.NoError(t, exp.Close())
	assert.Equal(t, "application/json", rec.headers.Get("Content-Type"))
	assert.Equal(t, "Bearer secret", rec.headers.Get("Authorization"))
	assert.Equal(t, map[string]any{
		"ts":   "2020-01-01T00:00:00Z",
		"addr": "CC:CA:7E:52:CC:34",
		"name": "Backyard",
		"temp": 21.5,
	}, rec.body)
}

func TestExportNoToken(t *testing.T) {
	rec := new(received)
	srv := newTestServer(t, nethttp.StatusAccepted, rec)
	defer srv.Close()
	exp, err := New(context.Background(), "test", Config{URL: srv.URL, Columns: sensor.DefaultColumnMap})
	require.NoError(t, err)
	require.NoError(t, exp.Export(context.Background(), testData))
	assert.Empty(t, rec.headers.Get("Authorization"))
}

func TestExportErrorStatus(t *testing.T) {
	rec := new(received)
	srv := newTestServer(t, nethttp.StatusInternalServerError, rec)
	defer srv.Close()
	exp, err := New(context.Background(), "test", Config{URL: srv.URL, Columns: sensor.DefaultColumnMap})
	require.NoError(t, err)
	err = exp.Export(context.Background(), testData)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "response body")
}

func TestValidate(t *testing.T) {
	require.Error(t, Validate(Config{Columns: sensor.DefaultColumnMap}))
	require.Error(t, Validate(Config{URL: "http://localhost"}))
	require.NoError(t, Validate(Config{URL: "http://localhost", Columns: sensor.DefaultColumnMap}))
}
