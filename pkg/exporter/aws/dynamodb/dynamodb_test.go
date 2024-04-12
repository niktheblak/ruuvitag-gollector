//go:build aws

package dynamodb

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
	t *testing.T
}

func (m *mockDynamoDBClient) PutItemWithContext(ctx aws.Context, input *dynamodb.PutItemInput, opts ...request.Option) (*dynamodb.PutItemOutput, error) {
	assert := assert.New(m.t)
	assert.Equal("CC:CA:7E:52:CC:34", *input.Item["mac"].S)
	assert.Equal("Backyard", *input.Item["name"].S)
	assert.Equal("21.5", *input.Item["temperature"].N)
	assert.Equal("2020-01-01T00:00:00Z", *input.Item["ts"].S)
	return &dynamodb.PutItemOutput{}, nil
}

func TestExport(t *testing.T) {
	exp := &dynamoDBExporter{
		db:    &mockDynamoDBClient{t: t},
		table: "test_table",
	}
	ctx := context.Background()
	data := sensor.Data{
		Addr:            "CC:CA:7E:52:CC:34",
		Name:            "Backyard",
		Temperature:     21.5,
		Humidity:        60,
		Pressure:        1002,
		BatteryVoltage:  50,
		AccelerationX:   0,
		AccelerationY:   0,
		AccelerationZ:   0,
		MovementCounter: 1,
		Timestamp:       time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	err := exp.Export(ctx, data)
	require.NoError(t, err)
}
