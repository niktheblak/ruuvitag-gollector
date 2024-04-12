//go:build aws

package sqs

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSQSClient struct {
	sqsiface.SQSAPI
	t *testing.T
}

func (m *mockSQSClient) SendMessageWithContext(ctx aws.Context, input *sqs.SendMessageInput, opts ...request.Option) (*sqs.SendMessageOutput, error) {
	assert := assert.New(m.t)
	assert.Equal("CC:CA:7E:52:CC:34", *input.MessageAttributes["mac"].StringValue)
	assert.Equal("Backyard", *input.MessageAttributes["name"].StringValue)
	assert.Equal("{\"mac\":\"CC:CA:7E:52:CC:34\",\"name\":\"Backyard\",\"temperature\":21.5,\"humidity\":60,\"pressure\":1002,\"battery_voltage\":50,\"acceleration_x\":0,\"acceleration_y\":0,\"acceleration_z\":0,\"movement_counter\":1,\"measurement_number\":0,\"ts\":\"2020-01-01T00:00:00Z\"}", *input.MessageBody)
	return &sqs.SendMessageOutput{}, nil
}

func TestExport(t *testing.T) {
	exp := &sqsExporter{
		sqs:      &mockSQSClient{t: t},
		queueUrl: "http://localhost/test_queue",
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
