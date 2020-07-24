package dynamodbiface

import (
	"context"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	aide "github.com/cleardataeng/aidews/dynamodb"
)

// Service provides access to data in DynamoDB.
type Service interface {
	GetItem(*dynamodb.GetItemInput, interface{}) error
	GetItemWithContext(context.Context, *dynamodb.GetItemInput, interface{}) error
	PutItem(*dynamodb.PutItemInput, interface{}) (*dynamodb.PutItemOutput, error)
	PutItemWithContext(context.Context, *dynamodb.PutItemInput, interface{}) (*dynamodb.PutItemOutput, error)
	Query(*dynamodb.QueryInput, interface{}) error
	QueryWithContext(context.Context, *dynamodb.QueryInput, interface{}) error
	QueryPages(in *dynamodb.QueryInput, outItems interface{}, outPager func(interface{}, bool) bool) error
	QueryPagesWithContext(ctx context.Context, in *dynamodb.QueryInput, outItems interface{}, outPager func(interface{}, bool) bool) error
	Scan(*dynamodb.ScanInput, interface{}) error
	ScanWithContext(context.Context, *dynamodb.ScanInput, interface{}) error
	ScanPages(in *dynamodb.ScanInput, outItems interface{}, outPager func(interface{}, bool) bool) error
	ScanPagesWithContext(ctx context.Context, in *dynamodb.ScanInput, outItems interface{}, outPager func(interface{}, bool) bool) error
}

var _ Service = (*aide.Service)(nil) // test that the aide satisfies the interface
