package dynamodb

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/cleardataeng/aidews/dynamodb/extmocks/github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/golang/mock/gomock"
)

//go:generate mockgen -destination=extmocks/github.com/aws/aws-sdk-go/service/dynamodb/mock.go github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface DynamoDBAPI

type row struct {
	Slug  string
	Title string
}

var (
	ctx = context.Background()

	pagesOutput1 = []map[string]*dynamodb.AttributeValue{
		map[string]*dynamodb.AttributeValue{"slug": {S: aws.String("xkcd")}, "title": {S: aws.String("Some guy")}},
		map[string]*dynamodb.AttributeValue{"slug": {S: aws.String("hijk")}, "title": {S: aws.String("vim")}},
	}
	pagesOutput2 = []map[string]*dynamodb.AttributeValue{
		map[string]*dynamodb.AttributeValue{"slug": {S: aws.String("wasd")}, "title": {S: aws.String("gaming")}},
	}

	pagesTable = "movement_keys"
)

func TestService_QueryPagesWithContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ddbMock := mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)

	ddbMock.EXPECT().QueryPagesWithContext(ctx, gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, input *dynamodb.QueryInput, f func(*dynamodb.QueryOutput, bool) bool) error {
			if *input.TableName != pagesTable {
				t.Errorf(`table: want: "%s", got: "%s"`, pagesTable, *input.TableName)
			}
			pagerReturn := f(&dynamodb.QueryOutput{Items: pagesOutput1}, false)
			if pagerReturn != false {
				t.Error("did not return false when pager returned false")
			}
			f(&dynamodb.QueryOutput{Items: pagesOutput2}, true)
			return nil
		},
	)

	slugs := []string{}
	titles := []string{}
	pager := func(item interface{}, lastPage bool) bool {
		fmt.Printf("item: %#v\n", item)
		slugs = append(slugs, item.(*row).Slug)
		titles = append(titles, item.(*row).Title)
		return false
	}

	input := &dynamodb.QueryInput{TableName: &pagesTable}
	svc := Service{svc: ddbMock}
	if err := svc.QueryPagesWithContext(ctx, input, &row{}, pager); err != nil {
		t.Error(err)
	}

	wantedSlugs := []string{"xkcd", "wasd"}
	if !reflect.DeepEqual(slugs, wantedSlugs) {
		t.Errorf(`slugs: wanted: {%s}, got: {%s}`, wantedSlugs, slugs)
	}
	wantedTitles := []string{"Some guy", "gaming"}
	if !reflect.DeepEqual(slugs, wantedSlugs) {
		t.Errorf(`titles: wanted: {%s}, got: {%s}`, wantedTitles, titles)
	}
}

func TestService_ScanPagesWithContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ddbMock := mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)

	ddbMock.EXPECT().ScanPagesWithContext(ctx, gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, input *dynamodb.ScanInput, f func(*dynamodb.ScanOutput, bool) bool) error {
			if *input.TableName != pagesTable {
				t.Errorf(`table: want: "%s", got: "%s"`, pagesTable, *input.TableName)
			}
			pagerReturn := f(&dynamodb.ScanOutput{Items: pagesOutput1}, false)
			if pagerReturn != false {
				t.Error("did not return false when pager returned false")
			}
			f(&dynamodb.ScanOutput{Items: pagesOutput2}, true)
			return nil
		},
	)

	slugs := []string{}
	titles := []string{}
	pager := func(item interface{}, lastPage bool) bool {
		fmt.Printf("item: %#v\n", item)
		slugs = append(slugs, item.(*row).Slug)
		titles = append(titles, item.(*row).Title)
		return false
	}

	input := &dynamodb.ScanInput{TableName: &pagesTable}
	svc := Service{svc: ddbMock}
	if err := svc.ScanPagesWithContext(ctx, input, &row{}, pager); err != nil {
		t.Error(err)
	}

	wantedSlugs := []string{"xkcd", "wasd"}
	if !reflect.DeepEqual(slugs, wantedSlugs) {
		t.Errorf(`slugs: wanted: {%s}, got: {%s}`, wantedSlugs, slugs)
	}
	wantedTitles := []string{"Some guy", "gaming"}
	if !reflect.DeepEqual(slugs, wantedSlugs) {
		t.Errorf(`titles: wanted: {%s}, got: {%s}`, wantedTitles, titles)
	}
}
