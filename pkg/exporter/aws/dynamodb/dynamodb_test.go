//go:build aws

package dynamodb

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
	input *dynamodb.PutItemInput
}

func (m *mockDynamoDBClient) PutItem(_ context.Context, input *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	m.input = input
	return &dynamodb.PutItemOutput{}, nil
}

func stringValue(t *testing.T, item map[string]types.AttributeValue, key string) string {
	t.Helper()
	v, ok := item[key].(*types.AttributeValueMemberS)
	require.True(t, ok, "attribute %s is not a string: %#v", key, item[key])
	return v.Value
}

func numberValue(t *testing.T, item map[string]types.AttributeValue, key string) string {
	t.Helper()
	v, ok := item[key].(*types.AttributeValueMemberN)
	require.True(t, ok, "attribute %s is not a number: %#v", key, item[key])
	return v.Value
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
	assert.Equal(t, "CC:CA:7E:52:CC:34", stringValue(t, mock.input.Item, "mac"))
	assert.Equal(t, "Backyard", stringValue(t, mock.input.Item, "name"))
	assert.Equal(t, "21.5", numberValue(t, mock.input.Item, "temperature"))
	assert.Equal(t, "2020-01-01T00:00:00Z", stringValue(t, mock.input.Item, "time"))
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
	assert.Equal(t, "CC:CA:7E:52:CC:34", stringValue(t, mock.input.Item, "addr"))
	assert.Equal(t, "21.5", numberValue(t, mock.input.Item, "temp"))
	assert.Equal(t, "2020-01-01T00:00:00Z", stringValue(t, mock.input.Item, "ts"))
	assert.NotContains(t, mock.input.Item, "name")
	assert.NotContains(t, mock.input.Item, "humidity")
}

func TestValidate(t *testing.T) {
	require.Error(t, Validate(Config{Columns: sensor.DefaultColumnMap}))
	require.Error(t, Validate(Config{Table: "test"}))
	require.NoError(t, Validate(Config{Table: "test", Columns: sensor.DefaultColumnMap}))
}
