package sqsbatch

//go:generate mockgen -destination=extmocks/aws/sqs/mocks/mock.go github.com/aws/aws-sdk-go/service/sqs/sqsiface SQSAPI

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cleardataeng/aidews/sqsbatch/extmocks/aws/sqs/mocks"
	"github.com/golang/mock/gomock"
	"testing"
)

func entries() (entries []*sqs.SendMessageBatchRequestEntry) {
	i := 0
	for i < 11 {
		entry := &sqs.SendMessageBatchRequestEntry{
			Id:                     aws.String(string(i + 65)),
			MessageDeduplicationId: aws.String("1b"),
			MessageBody:            aws.String("1c"),
		}
		entries = append(entries, entry)
		i++
	}
	return entries
}

func compareInputs(t *testing.T, entries []*sqs.SendMessageBatchRequestEntry, batchInput *sqs.SendMessageBatchInput) (*sqs.SendMessageBatchOutput, error) {
	if *batchInput.QueueUrl != "foo" {
		t.Errorf("bad QueueURL want: %s got: %s", "foo", *batchInput.QueueUrl)
	}
	if len(batchInput.Entries) != len(entries) {
		t.Errorf("entries length want: %d got: %d", len(entries), len(batchInput.Entries))
	} else {
		for i, entry := range batchInput.Entries {
			if *entry.Id != *entries[i].Id {
				t.Errorf("entry %d Id want: %s got: %s", i, *entries[i].Id, *entry.Id)
			}
		}
	}
	return &sqs.SendMessageBatchOutput{Failed: nil}, nil
}

func TestIface(t *testing.T) {
	var _ Iface = &SqsBatch{}
}

func TestBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sqsmock := mock_sqsiface.NewMockSQSAPI(ctrl)

	entries := entries()

	sqsmock.EXPECT().SendMessageBatch(gomock.Any()).Times(1).
		DoAndReturn(
			func(got *sqs.SendMessageBatchInput) (*sqs.SendMessageBatchOutput, error) {
				return compareInputs(t, entries[:10], got)
			},
		)

	b := New(sqsmock, "foo")
	for _, entry := range entries {
		b.Add(entry)
	}
}

func TestSend(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sqsmock := mock_sqsiface.NewMockSQSAPI(ctrl)

	entries := entries()[:2]

	b := New(sqsmock, "foo")
	for _, entry := range entries {
		b.Add(entry)
	}

	sqsmock.EXPECT().SendMessageBatch(gomock.Any()).Times(1).
		DoAndReturn(
			func(got *sqs.SendMessageBatchInput) (*sqs.SendMessageBatchOutput, error) {
				return compareInputs(t, entries, got)
			},
		)

	b.Send()
}

func TestSendWithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sqsmock := mock_sqsiface.NewMockSQSAPI(ctrl)

	entries := entries()[:2]

	b := New(sqsmock, "foo")
	for _, entry := range entries {
		b.Add(entry)
	}

	sqsmock.EXPECT().SendMessageBatch(gomock.Any()).Times(1).
		Return(
			&sqs.SendMessageBatchOutput{
				Failed: []*sqs.BatchResultErrorEntry{
					&sqs.BatchResultErrorEntry{},
				},
			},
			nil,
		)

	err := b.Send()
	expectedError := fmt.Errorf("error sending SQS batch: {\n  Failed: [{\n\n    }]\n}")
	if err == nil || err.Error() != expectedError.Error() {
		t.Errorf("didn't error properly, want: %#v, got: %#v", expectedError, err)
	}
}

func TestSendWithNoMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sqsmock := mock_sqsiface.NewMockSQSAPI(ctrl)

	b := New(sqsmock, "foo")

	sqsmock.EXPECT().SendMessageBatch(gomock.Any()).Times(0)

	err := b.Send()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
