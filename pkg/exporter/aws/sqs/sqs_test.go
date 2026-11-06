//go:build aws

package sqs

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
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

type mockSQSClient struct {
	input *sqs.SendMessageInput
}

func (m *mockSQSClient) SendMessage(_ context.Context, input *sqs.SendMessageInput, _ ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	m.input = input
	return &sqs.SendMessageOutput{}, nil
}

func TestExport(t *testing.T) {
	mock := &mockSQSClient{}
	exp := newExporter("test", mock, "http://localhost/test_queue", sensor.DefaultColumnMap, nil)
	assert.False(t, exp.fifo)
	err := exp.Export(context.Background(), testData)
	require.NoError(t, err)
	require.NotNil(t, mock.input)
	assert.Equal(t, "http://localhost/test_queue", *mock.input.QueueUrl)
	assert.Equal(t, "CC:CA:7E:52:CC:34", *mock.input.MessageAttributes["mac"].StringValue)
	assert.Equal(t, "Backyard", *mock.input.MessageAttributes["name"].StringValue)
	assert.Nil(t, mock.input.MessageGroupId)
	assert.Nil(t, mock.input.MessageDeduplicationId)
	var body map[string]any
	require.NoError(t, json.Unmarshal([]byte(*mock.input.MessageBody), &body))
	assert.Equal(t, 21.5, body["temperature"])
	assert.Equal(t, "Backyard", body["name"])
	assert.Equal(t, "2020-01-01T00:00:00Z", body["time"])
}

func TestExportFIFO(t *testing.T) {
	mock := &mockSQSClient{}
	exp := newExporter("test", mock, "https://sqs.us-east-2.amazonaws.com/123456789012/measurements.fifo", sensor.DefaultColumnMap, nil)
	assert.True(t, exp.fifo)
	err := exp.Export(context.Background(), testData)
	require.NoError(t, err)
	require.NotNil(t, mock.input)
	assert.Equal(t, "CC:CA:7E:52:CC:34", *mock.input.MessageGroupId)
	assert.Equal(t, "CC:CA:7E:52:CC:34-1577836800000000000", *mock.input.MessageDeduplicationId)
}

func TestExportCustomColumns(t *testing.T) {
	mock := &mockSQSClient{}
	columns := map[string]string{
		"time":        "ts",
		"mac":         "addr",
		"temperature": "temp",
	}
	exp := newExporter("test", mock, "http://localhost/test_queue", columns, nil)
	err := exp.Export(context.Background(), testData)
	require.NoError(t, err)
	require.NotNil(t, mock.input)
	assert.Equal(t, "CC:CA:7E:52:CC:34", *mock.input.MessageAttributes["addr"].StringValue)
	assert.NotContains(t, mock.input.MessageAttributes, "name")
	var body map[string]any
	require.NoError(t, json.Unmarshal([]byte(*mock.input.MessageBody), &body))
	assert.Equal(t, map[string]any{
		"ts":   "2020-01-01T00:00:00Z",
		"addr": "CC:CA:7E:52:CC:34",
		"temp": 21.5,
	}, body)
}

func TestIsFIFO(t *testing.T) {
	assert.True(t, isFIFO("https://sqs.us-east-2.amazonaws.com/123456789012/measurements.fifo"))
	assert.True(t, isFIFO("https://sqs.us-east-2.amazonaws.com/123456789012/measurements.fifo/"))
	assert.False(t, isFIFO("https://sqs.us-east-2.amazonaws.com/123456789012/measurements"))
	assert.False(t, isFIFO("https://sqs.us-east-2.amazonaws.com/123456789012/fifo"))
}

func TestValidate(t *testing.T) {
	require.Error(t, Validate(Config{Columns: sensor.DefaultColumnMap}))
	require.Error(t, Validate(Config{QueueName: "test"}))
	require.NoError(t, Validate(Config{QueueName: "test", Columns: sensor.DefaultColumnMap}))
	require.NoError(t, Validate(Config{QueueURL: "http://localhost/test", Columns: sensor.DefaultColumnMap}))
}
