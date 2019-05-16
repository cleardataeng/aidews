// Package sqs helps us send messages to an SQS queue.
//
// Specifically, batch sending:
// Use NewBatch() to get a Batch object. Use the object's Add() method to add as many messages as you want.
// The object will add them to the queue in batches of 10 (so that's 1 AWS API call every 10 messages).
// After you are done adding messages, call Send() to finish sending the messages. (If you Add() 23 messages,
// 20 will get sent automatically in 2 batches, but you need an explicit Send() to send the last 3.)
// Example:
// for _, msg := range messages {
// 	sqsbatch.Add(msg)
// }
// sqsbatch.Send()
package sqs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	log "github.com/sirupsen/logrus"
)

// maxSqsBatchEntries is the maximum entries allowed by AWS in an SQS SendMessageBatch call.
// https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-limits.html#limits-messages
const maxSqsBatchEntries = 10

// Batch object
type Batch struct {
	sqsiface.SQSAPI
	queueURL *string
	messages []*sqs.SendMessageBatchRequestEntry
	*sqs.SendMessageBatchInput
}

// BatchIface for testing
type BatchIface interface {
	SetQueueURL(*string)
	Add(message *sqs.SendMessageBatchRequestEntry) error
	Send() error
}

// NewBatch takes an SQS API interface and SQS queue URL for the target SQS queue.
// Returns a struct that can Add() and Send() messages.
// queueURL can be set later with SetQueueURL()
func NewBatch(sqsapi sqsiface.SQSAPI, queueURL *string) *Batch {
	return &Batch{SQSAPI: sqsapi, queueURL: queueURL}
}

// SetQueueURL sets the queue URL
func (r *Batch) SetQueueURL(queueURL *string) {
	r.queueURL = queueURL
}

// Add the message to the batch
func (r *Batch) Add(message *sqs.SendMessageBatchRequestEntry) (err error) {
	if err := message.Validate(); err != nil {
		return err
	}
	r.messages = append(r.messages, message)
	if len(r.messages) >= maxSqsBatchEntries {
		return r.Send()
	}
	return nil
}

// Send any batched messages to the queue
func (r *Batch) Send() (err error) {
	if len(r.messages) < 1 {
		return nil
	}
	if r.SendMessageBatchInput == nil {
		r.SendMessageBatchInput = &sqs.SendMessageBatchInput{
			QueueUrl: r.queueURL,
		}
	}
	r.SendMessageBatchInput.SetEntries(r.messages)
	result, err := r.SQSAPI.SendMessageBatch(r.SendMessageBatchInput)
	log.Info(result)
	r.messages = nil
	if err == nil && len(result.Failed) > 0 {
		return fmt.Errorf("error sending SQS batch: %s", result.String())
	}
	return err
}
