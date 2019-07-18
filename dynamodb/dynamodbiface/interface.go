package dynamodbiface

import "github.com/aws/aws-sdk-go/service/dynamodb"

// Service provides access to data in DynamoDB.
type Service interface {
	GetItem(*dynamodb.GetItemInput, interface{}) error
	PutItem(*dynamodb.PutItemInput, interface{}) (*dynamodb.PutItemOutput, error)
	Query(*dynamodb.QueryInput, interface{}) error
	QueryPages(in *dynamodb.QueryInput, outItems []interface{}, outPager func(interface{}, bool) bool) error
	Scan(*dynamodb.ScanInput, interface{}) error
	ScanPages(in *dynamodb.ScanInput, outItems []interface{}, outPager func(interface{}, bool) bool) error
}
