//go:build influxdb && integration_test

package influxdb

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

const queryTmpl = `from(bucket:"%s")
|> range(start:-5m)
|> filter(fn:(r) =>
     r._measurement == "test" and
     r.name == "Backyard"
)`

func TestExporter(t *testing.T) {
	addr := os.Getenv("INFLUXDB_HOST")
	token := os.Getenv("INFLUXDB_TOKEN")
	client := influxdb2.NewClient(addr, token)
	defer client.Close()
	ctx := context.Background()
	org, err := client.OrganizationsAPI().FindOrganizationByName(ctx, "test")
	require.NoError(t, err)
	bucket, err := client.BucketsAPI().FindBucketByName(ctx, "test")
	require.NoError(t, err)
	auth, err := client.AuthorizationsAPI().CreateAuthorizationWithOrgID(ctx, *org.Id, []domain.Permission{
		{
			Action: domain.PermissionActionRead,
			Resource: domain.Resource{
				Id:    bucket.Id,
				OrgID: org.Id,
				Type:  domain.ResourceTypeBuckets,
			},
		},
		{
			Action: domain.PermissionActionWrite,
			Resource: domain.Resource{
				Id:    bucket.Id,
				OrgID: org.Id,
				Type:  domain.ResourceTypeBuckets,
			},
		},
	})
	require.NoError(t, err)
	t.Run("TestReport", func(t *testing.T) {
		exporter := New(Config{
			Addr:        addr,
			Org:         org.Name,
			Token:       *auth.Token,
			Bucket:      bucket.Name,
			Measurement: "test",
		})
		err := exporter.Export(context.Background(), sensor.Data{
			Addr:           "CC:CA:7E:52:CC:34",
			Name:           "Backyard",
			Temperature:    22.1,
			Humidity:       45.0,
			DewPoint:       9.6,
			Pressure:       1002.0,
			BatteryVoltage: 2.755,
			AccelerationX:  0,
			AccelerationY:  0,
			AccelerationZ:  0,
			Timestamp:      time.Now(),
		})
		require.NoError(t, err)
		exporter.Close()
		res, err := client.QueryAPI(org.Name).Query(ctx, fmt.Sprintf(queryTmpl, bucket.Name))
		require.NoError(t, err)
		ok := res.Next()
		require.True(t, ok)
		for ok && res.Record().Field() != "temperature" {
			ok = res.Next()
		}
		assert.Equal(t, 22.1, res.Record().Value())
		res.Close()
	})
}
