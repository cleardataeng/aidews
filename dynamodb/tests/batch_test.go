package batch_test

//go:generate mockgen -destination=extmocks/aws-sdk-go/service/dynamodb/mock.go github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface DynamoDBAPI

import (
	"fmt"
	"testing"

	aide "github.com/cleardataeng/aidews/dynamodb"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	mock_dynamodbiface "github.com/cleardataeng/aidews/dynamodb/tests/extmocks/aws-sdk-go/service/dynamodb"
	"github.com/golang/mock/gomock"
)

func TestIface(t *testing.T) {
	type consumer struct{ bwiface aide.BatchIface }
	_ = consumer{bwiface: &aide.Batch{}}
}

func TestAddWithoutTableName(t *testing.T) {
	b := &aide.Batch{}
	_, err := b.Add(aide.PutRequest, map[string]*dynamodb.AttributeValue{})

	wantedError := "table name required, call SetTableName"
	if err == nil {
		t.Error("did not complain about table name")
	} else {
		if err.Error() != wantedError {
			t.Errorf(`error wanted: "%s", got: "%s"`, wantedError, err)
		}

	}
}

func TestAdd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ddb := mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)

	ddb.EXPECT().BatchWriteItem(gomock.Any()).Times(0)

	b := &aide.Batch{RequestLimit: 3, BwInput: aide.BasicBatchInput(), DynamoAPI: ddb}
	b.SetTableName("foo")
	if _, err := b.Add(aide.PutRequest, map[string]*dynamodb.AttributeValue{"slug": {S: aws.String("a")}}); err != nil {
		t.Error(err)
	}
	if _, err := b.Add(aide.DeleteRequest, map[string]*dynamodb.AttributeValue{"id": {S: aws.String("b")}}); err != nil {
		t.Error(err)
	}

	if b.Requests[0].PutRequest == nil {
		t.Error("put request was nil")
	} else {
		if b.Requests[0].DeleteRequest != nil {
			t.Error("put request has delete request")
		}
		if *b.Requests[0].PutRequest.Item["slug"].S != "a" {
			t.Error("put request did not have the right item")
		}
	}
	if b.Requests[1].DeleteRequest == nil {
		t.Error("delete request was nil")
	} else {
		if b.Requests[1].PutRequest != nil {
			t.Error("put request has delete request")
		}
		if *b.Requests[1].DeleteRequest.Key["id"].S != "b" {
			t.Error("delete request did not have the right key")
		}
	}
}

func TestAddAndSend(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tableName := "foo"
	ddb := mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)

	ddb.EXPECT().BatchWriteItem(gomock.Any()).DoAndReturn(
		func(input *dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error) {
			items, found := input.RequestItems[tableName]
			if !found {
				t.Error("no entry found for table foo")
			}
			if len(items) != 2 {
				t.Errorf("wanted 2 items, got %v", items)
			}
			return &dynamodb.BatchWriteItemOutput{
				ConsumedCapacity: []*dynamodb.ConsumedCapacity{
					{TableName: &tableName, CapacityUnits: aws.Float64(23.3)},
				},
			}, nil
		},
	)

	b := &aide.Batch{RequestLimit: 2, BwInput: aide.BasicBatchInput(), DynamoAPI: ddb}
	b.SetTableName(tableName)
	for i := 0; i < 3; i++ {
		out, err := b.Add(aide.PutRequest, map[string]*dynamodb.AttributeValue{
			"index": &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", i))},
		})
		if err != nil {
			t.Error(err)
		}
		if out != nil {
			if *out[0].CapacityUnits != 23.3 {
				t.Errorf("capacity units want: %f got: %f", 23.3, *out[0].CapacityUnits)
			}
			if *out[0].TableName != tableName {
				t.Errorf("out table name: want: %s got: %s", tableName, *out[0].TableName)
			}
		}
	}

}

func TestSendWithRetry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tableName := "foo"
	ddb := mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)

	ddb.EXPECT().BatchWriteItem(gomock.Any()).DoAndReturn(
		func(input *dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error) {
			items, found := input.RequestItems[tableName]
			if !found {
				t.Error("no entry found for table foo")
			}
			if len(items) != 3 {
				t.Errorf("wanted 3 items, got %v", items)
			}
			return &dynamodb.BatchWriteItemOutput{
				ConsumedCapacity: []*dynamodb.ConsumedCapacity{
					{TableName: &tableName, CapacityUnits: aws.Float64(23.3)},
				},
				UnprocessedItems: map[string][]*dynamodb.WriteRequest{
					tableName: []*dynamodb.WriteRequest{items[2]},
				},
			}, nil
		},
	)
	ddb.EXPECT().BatchWriteItem(gomock.Any()).DoAndReturn(
		func(input *dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error) {
			items, found := input.RequestItems[tableName]
			if !found {
				t.Error("no entry found for table foo")
			}
			if *items[0].PutRequest.Item["index"].N != "2" {
				t.Errorf(`wanted "2", got %s`, *items[0].PutRequest.Item["index"].N)
			}
			return &dynamodb.BatchWriteItemOutput{
				ConsumedCapacity: []*dynamodb.ConsumedCapacity{
					{TableName: &tableName, CapacityUnits: aws.Float64(24.3)},
				},
			}, nil
		},
	)

	b := &aide.Batch{RequestLimit: 2, BwInput: aide.BasicBatchInput(), DynamoAPI: ddb, SleepSeconds: -1}
	b.SetTableName(tableName)
	b.Requests = []*dynamodb.WriteRequest{
		&dynamodb.WriteRequest{PutRequest: &dynamodb.PutRequest{Item: map[string]*dynamodb.AttributeValue{
			"index": &dynamodb.AttributeValue{N: aws.String("0")},
		}}},
		&dynamodb.WriteRequest{PutRequest: &dynamodb.PutRequest{Item: map[string]*dynamodb.AttributeValue{
			"index": &dynamodb.AttributeValue{N: aws.String("1")},
		}}},
		&dynamodb.WriteRequest{PutRequest: &dynamodb.PutRequest{Item: map[string]*dynamodb.AttributeValue{
			"index": &dynamodb.AttributeValue{N: aws.String("2")},
		}}},
	}
	out, err := b.Send()
	if err != nil {
		t.Error(err)
	}
	if out != nil {
		if *out[0].CapacityUnits != 24.3 {
			t.Errorf("capacity units want: %f got: %f", 24.3, *out[0].CapacityUnits)
		}
		if *out[0].TableName != tableName {
			t.Errorf("out table name: want: %s got: %s", tableName, *out[0].TableName)
		}
	}
}
