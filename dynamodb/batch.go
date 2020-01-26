package dynamodb

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

// RequestType tells us if it's a put or delete
type RequestType int8

const (
	// PutRequest is a PutItem request
	PutRequest RequestType = 0
	// DeleteRequest is a DeleteItem request
	DeleteRequest RequestType = 1
	// requestLimit is the number of requests in a BatchWriteItem. While the numeric limit from AWS is 25, there is also a size limit in megabytes.
	requestLimit = 25
	// After the sleep reaches this value of seconds (for retrying unprocessed items), we error with a time-out
	maxTimeout int64 = 10
)

// Batch provides an aide for dynamodb's BatchWriteItem function.
// This aide wraps the complexities of building the batch and retrying unprocessed items,
// at the cost of being able to only do 1 table at a time.
//
// Use NewBatch() to get a Batch object. SetTableName(), and then
// use the object's Add() method to add as many dynamodb items as you want.
// The object will add them to the queue in batches of 10 (so that's 1 AWS API call every 10 items).
// After you are done adding items, call Send() to finish sending the items. (If you Put() 23 items,
// 20 will get sent automatically in 2 batches, but you need an explicit Send() to send the last 3.)
// Example:
// for _, item := range items {
// 	 capacity, err := batch.Add(PutRequest, item)
// }
// batch.Send()
// Tell Add() whether it's a PutRequest or a DeleteRequest, and pass either the item to be put
// or the Key of the item to be deleted. Either way, pass a map[string]*dynamodb.AttributeValue{}
type Batch struct {
	Table        string
	BwInput      dynamodb.BatchWriteItemInput
	Requests     []*dynamodb.WriteRequest
	RequestLimit int   // number of requests per batch
	MaxTimeout   int64 // maximum time to sleep between retries
	DynamoAPI    dynamodbiface.DynamoDBAPI
	SleepSeconds int64 // Backoff
}

// BatchIface presents a test interface for Batch
type BatchIface interface {
	SetTableName(tableName string)
	SetRequestLimit(newLimit int) error
	SetMaxTimeout(newTimeout int64) error
	Add(requestType RequestType, item map[string]*dynamodb.AttributeValue) ([]*dynamodb.ConsumedCapacity, error)
	Send() ([]*dynamodb.ConsumedCapacity, error)
}

// NewBatch returns a new object with the given table name, request limit, and dynamo API
func NewBatch(ddb dynamodbiface.DynamoDBAPI) *Batch {
	return &Batch{
		DynamoAPI:    ddb,
		SleepSeconds: 1,
		RequestLimit: requestLimit,
		MaxTimeout:   maxTimeout,
		BwInput:      BasicBatchInput(),
	}
}

func BasicBatchInput() dynamodb.BatchWriteItemInput {
	return dynamodb.BatchWriteItemInput{
		RequestItems:           map[string][]*dynamodb.WriteRequest{},
		ReturnConsumedCapacity: aws.String("TOTAL"),
	}
}

// SetTableName sets the table name
func (b *Batch) SetTableName(tableName string) {
	b.Table = tableName
	return
}

// SetRequestLimit sets the number of items to pile up before sending
func (b *Batch) SetRequestLimit(newLimit int) error {
	if newLimit > requestLimit {
		return fmt.Errorf("requestLimit must be %d or less", requestLimit)
	}
	if newLimit < 1 {
		return fmt.Errorf("requestLimit must be %d or less", requestLimit)
	}
	b.RequestLimit = newLimit
	return nil
}

// SetMaxTimeout sets the maximum sleep allowed for retries
func (b *Batch) SetMaxTimeout(newTimeout int64) error {
	if newTimeout > maxTimeout {
		return fmt.Errorf("too much timeout, should be less than %d", maxTimeout)
	}
	if newTimeout < 1 {
		return fmt.Errorf("not enough time out, should be at least 1")
	}
	b.MaxTimeout = int64(newTimeout)
	return nil
}

// Add an item to the batch, and send the batch if it's full
func (b *Batch) Add(requestType RequestType, item map[string]*dynamodb.AttributeValue) ([]*dynamodb.ConsumedCapacity, error) {
	if b.Table == "" {
		return nil, fmt.Errorf("table name required, call SetTableName")
	}
	request := &dynamodb.WriteRequest{}
	switch requestType {
	case PutRequest:
		request.PutRequest = &dynamodb.PutRequest{Item: item}
	case DeleteRequest:
		request.DeleteRequest = &dynamodb.DeleteRequest{Key: item}
	}
	b.Requests = append(b.Requests, request)
	if len(b.Requests) >= b.RequestLimit {
		return b.Send()
	}
	return nil, nil
}

// Send the batch, if there's anything in it
func (b *Batch) Send() ([]*dynamodb.ConsumedCapacity, error) {
	if len(b.Requests) < 1 {
		return nil, nil
	}
	if b.SleepSeconds > b.MaxTimeout {
		return nil, fmt.Errorf("we timed out - maxTimeout is %d", b.MaxTimeout)
	}
	b.BwInput.RequestItems[b.Table] = b.Requests
	out, err := b.DynamoAPI.BatchWriteItem(&b.BwInput)
	if err != nil {
		return nil, err
	}
	b.Requests = out.UnprocessedItems[b.Table]
	if len(b.Requests) > 0 {
		b.SleepSeconds++
		time.Sleep(time.Duration(b.SleepSeconds))
		return b.Send()
	}
	b.SleepSeconds = 0
	delete(b.BwInput.RequestItems, b.Table)
	return out.ConsumedCapacity, nil
}
