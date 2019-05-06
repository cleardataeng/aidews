package dynamodbiface

import "github.com/aws/aws-sdk-go/service/dynamodb"

// Service provides access to data in DynamoDB.
type Service interface {
	GetItem(*dynamodb.GetItemInput, interface{}) error
	Query(*dynamodb.QueryInput, interface{}) error
	Scan(*dynamodb.ScanInput, interface{}) error
	PutItem(*dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
}
