package batch

//go:generate mockgen -destination=extmocks/aws-sdk-go/service/dynamodb/mock.go github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface DynamoDBAPI

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/cleardataeng/aidews/dynamodb/batch/extmocks/aws-sdk-go/service/dynamodb"
	"github.com/golang/mock/gomock"
	"testing"
)

func TestIface(t *testing.T) {
	type consumer struct{ bwiface Iface }
	_ = consumer{bwiface: &Batch{}}
}

func TestAddWithoutTableName(t *testing.T) {
	b := &Batch{}
	_, err := b.Add(PutRequest, map[string]*dynamodb.AttributeValue{})

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

	b := &Batch{requestLimit: 3, bwInput: basicBatchInput(), dynamoapi: ddb}
	b.SetTableName("foo")
	if _, err := b.Add(PutRequest, map[string]*dynamodb.AttributeValue{"slug": {S: aws.String("a")}}); err != nil {
		t.Error(err)
	}
	if _, err := b.Add(DeleteRequest, map[string]*dynamodb.AttributeValue{"id": {S: aws.String("b")}}); err != nil {
		t.Error(err)
	}

	if b.requests[0].PutRequest == nil {
		t.Error("put request was nil")
	} else {
		if b.requests[0].DeleteRequest != nil {
			t.Error("put request has delete request")
		}
		if *b.requests[0].PutRequest.Item["slug"].S != "a" {
			t.Error("put request did not have the right item")
		}
	}
	if b.requests[1].DeleteRequest == nil {
		t.Error("delete request was nil")
	} else {
		if b.requests[1].PutRequest != nil {
			t.Error("put request has delete request")
		}
		if *b.requests[1].DeleteRequest.Key["id"].S != "b" {
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

	b := &Batch{requestLimit: 2, bwInput: basicBatchInput(), dynamoapi: ddb}
	b.SetTableName(tableName)
	for i := 0; i < 3; i++ {
		out, err := b.Add(PutRequest, map[string]*dynamodb.AttributeValue{
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

	b := &Batch{requestLimit: 2, bwInput: basicBatchInput(), dynamoapi: ddb, sleepSeconds: -1}
	b.SetTableName(tableName)
	b.requests = []*dynamodb.WriteRequest{
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
