//go:build aws

package dynamodb

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
)

var testData = sensor.Data{
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

type mockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
	input *dynamodb.PutItemInput
}

func (m *mockDynamoDBClient) PutItemWithContext(_ aws.Context, input *dynamodb.PutItemInput, _ ...request.Option) (*dynamodb.PutItemOutput, error) {
	m.input = input
	return &dynamodb.PutItemOutput{}, nil
}

func TestExport(t *testing.T) {
	mock := &mockDynamoDBClient{}
	exp := &dynamoDBExporter{
		db:      mock,
		table:   "test_table",
		columns: sensor.DefaultColumnMap,
		logger:  slog.New(slog.DiscardHandler),
	}
	err := exp.Export(context.Background(), testData)
	require.NoError(t, err)
	require.NotNil(t, mock.input)
	assert.Equal(t, "test_table", *mock.input.TableName)
	assert.Equal(t, "CC:CA:7E:52:CC:34", *mock.input.Item["mac"].S)
	assert.Equal(t, "Backyard", *mock.input.Item["name"].S)
	assert.Equal(t, "21.5", *mock.input.Item["temperature"].N)
	assert.Equal(t, "2020-01-01T00:00:00Z", *mock.input.Item["time"].S)
}

func TestExportCustomColumns(t *testing.T) {
	mock := &mockDynamoDBClient{}
	exp := &dynamoDBExporter{
		db:     mock,
		table:  "test_table",
		logger: slog.New(slog.DiscardHandler),
		columns: map[string]string{
			"time":        "ts",
			"mac":         "addr",
			"temperature": "temp",
		},
	}
	err := exp.Export(context.Background(), testData)
	require.NoError(t, err)
	require.NotNil(t, mock.input)
	assert.Equal(t, "CC:CA:7E:52:CC:34", *mock.input.Item["addr"].S)
	assert.Equal(t, "21.5", *mock.input.Item["temp"].N)
	assert.Equal(t, "2020-01-01T00:00:00Z", *mock.input.Item["ts"].S)
	assert.NotContains(t, mock.input.Item, "name")
	assert.NotContains(t, mock.input.Item, "humidity")
}

func TestValidate(t *testing.T) {
	require.Error(t, Validate(Config{Columns: sensor.DefaultColumnMap}))
	require.Error(t, Validate(Config{Table: "test"}))
	require.NoError(t, Validate(Config{Table: "test", Columns: sensor.DefaultColumnMap}))
}
