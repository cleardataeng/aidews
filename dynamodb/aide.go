// Package dynamodb provides a DynamoDB wrapper object.
//
// This object makes retrieving from DynamoDB more simple and more uniform for
// our processes. It is not required, but since it provides a ReadWriter
// interface it makes testing much easier.
package dynamodb

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/cleardataeng/aidews"
)

// Service provides access to data in DynamoDB.
type Service struct {
	svc dynamodbiface.DynamoDBAPI
}

// New returns an initialized DB aide.
func New(region, roleARN *string) *Service {
	return &Service{
		svc: dynamodb.New(aidews.Session(region, roleARN)),
	}
}

// GetItem and unmarshal response items into given interface{}.
func (svc *Service) GetItem(in *dynamodb.GetItemInput, out interface{}) error {
	resp, err := svc.svc.GetItem(in)
	if err != nil {
		return err
	}
	if resp.Item == nil {
		return fmt.Errorf("GetItem failed")
	}
	return dynamodbattribute.UnmarshalMap(resp.Item, out)
}

// Query the table and unmarshal all results.
func (svc *Service) Query(in *dynamodb.QueryInput, out interface{}) error {
	items := []map[string]*dynamodb.AttributeValue{}
	pager := func(out *dynamodb.QueryOutput, last bool) bool {
		items = append(items, out.Items...)
		return !last
	}
	if err := svc.svc.QueryPages(in, pager); err != nil {
		return err
	}
	return dynamodbattribute.UnmarshalListOfMaps(items, out)
}

// Scan the table and return all results.
func (svc *Service) Scan(in *dynamodb.ScanInput, out interface{}) error {
	items := []map[string]*dynamodb.AttributeValue{}
	pager := func(out *dynamodb.ScanOutput, last bool) bool {
		items = append(items, out.Items...)
		return !last
	}
	if err := svc.svc.ScanPages(in, pager); err != nil {
		return err
	}
	return dynamodbattribute.UnmarshalListOfMaps(items, out)
}

// PutItem in the table.
func (svc *Service) PutItem(in *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return svc.svc.PutItem(in)
}
