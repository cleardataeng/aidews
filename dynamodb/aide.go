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

// PutItem by marshalling the given interafce{} into the given PutItemInput.
func (svc *Service) PutItem(in *dynamodb.PutItemInput, item interface{}) (out *dynamodb.PutItemOutput, err error) {
	if in.Item, err = dynamodbattribute.MarshalMap(item); err != nil {
		return nil, err
	}
	return svc.svc.PutItem(in)
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

// QueryPages queries the table and unmarshals each page of results to provided function.
// Caller should provide an item of some type that will be populated and each item will be
// returned to the provided pager function.
// Provided pager function should take a single interface{} and assert the type of the item (e.g. Item),
// and a boolean which will indicate whether this is the last item.
// Provided pager function should return false if it wants to stop processing.
// Example:
// items := Item
// pager := func(out interface{}, last bool) bool {
//   if out.(Item).Found { return false }
//   return !last
// }
// if err := QueryPages(queryInput, item, pager); err != nil {
//	return err
// }
func (svc *Service) QueryPages(in *dynamodb.QueryInput, outItem interface{}, outPager func(interface{}, bool) bool) error {
	var marshallErr error
	pager := func(pageOut *dynamodb.QueryOutput, last bool) bool {
		for _, item := range pageOut.Items {
			marshallErr = dynamodbattribute.UnmarshalMap(item, outItem)
			if marshallErr != nil {
				return false
			}
			if !outPager(outItem, last) {
				return false
			}
		}
		return !last
	}
	if err := svc.svc.QueryPages(in, pager); err != nil {
		return err
	}
	return marshallErr
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
