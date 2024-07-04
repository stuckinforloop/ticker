package broker

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/spf13/viper"
)

type SQS struct {
	client *sqs.SQS
}

func NewSQS() *SQS {
	awsRegion := viper.GetString("aws.region")
	awsConfig := aws.NewConfig().WithRegion(awsRegion)

	anonymousCreds := viper.GetBool("aws.use_anonymous_credentials")
	if anonymousCreds {
		awsConfig.WithCredentials(credentials.AnonymousCredentials)
	}

	awsConfig.WithHTTPClient(http.DefaultClient)

	s, err := session.NewSession(awsConfig)
	if err != nil {
		panic(err)
	}

	sqsClient := sqs.New(s)

	return &SQS{sqsClient}
}

func (q *SQS) Enqueue(
	ctx context.Context, queueName string, message string,
	dedupeID string, groupID string, delay int64,
) (string, error) {
	result, err := q.client.SendMessage(
		&sqs.SendMessageInput{
			DelaySeconds:           aws.Int64(delay),
			MessageBody:            aws.String(message),
			MessageDeduplicationId: &dedupeID,
			MessageGroupId:         &groupID,
			QueueUrl:               aws.String(q.getQueueURL(queueName)),
		})

	if err != nil {
		return "", fmt.Errof("enqueue failed: %w", err)
	}

	return *result.MessageId, nil
}

func (q *SQS) Dequeue(
	ctx context.Context, queueName string, visibilityTimeout int64) (string, string, error,
) {
	result, err := q.client.ReceiveMessage(
		&sqs.ReceiveMessageInput{
			AttributeNames: []*string{
				aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
			},
			MessageAttributeNames: []*string{
				aws.String(sqs.QueueAttributeNameAll),
			},
			QueueUrl:          aws.String(q.getQueueURL(queueName)),
			VisibilityTimeout: aws.Int64(visibilityTimeout),
			// WaitTimeSeconds:     aws.Int64(waitTimeout),
		})

	if err != nil {
		return "", "", fmt.Errof("dequeue failed: %w", err)
	}

	if len(result.Messages) == 0 {
		return "", "", nil
	}

	msg := result.Messages[0]
	return *msg.ReceiptHandle, *msg.Body, nil
}

func (q *SQS) Acknowledge(
	ctx context.Context, id string, queueName string,
) error {
	_, err := q.client.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(q.getQueueURL(queueName)),
		ReceiptHandle: aws.String(id),
	})

	return err
}

func (q *SQS) getQueueURL(queueName string) string {
	prefix := viper.GetString("aws.sqs_queue_prefix")
	return fmt.Sprintf("%s/%s", strings.TrimRight(prefix, "/"), queueName)
}
