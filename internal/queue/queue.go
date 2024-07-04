package queue

import (
	"context"

	"github.com/spf13/viper"
	"github.com/stuckinforloop/ticker/internal/queue/broker"
)

const QueueBrokerSQS = "sqs"

type Queue interface {
	Enqueue(
		ctx context.Context,
		queueName string,
		message string,
		dedupeID string,
		groupID string,
		delay int64) (string, error)

	Dequeue(
		ctx context.Context,
		queueName string,
		visibilityTimeout int64) (string, string, error)

	Acknowledge(
		ctx context.Context,
		id string,
		queueName string) error
}

func New() Queue {
	queueBroker := viper.GetString("queue_broker")
	switch queueBroker {
	case QueueBrokerSQS:
		return broker.NewSQS()
	default:
		panic("unsupported queue broker")
	}
}
