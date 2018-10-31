package s3

import (
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/cleardataeng/aidews"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"fmt"
)

// Service for reading and writing to the given bucket.
type Service struct {
	// acl is the default ACL for bucket objects.
	acl *string

	// name of the bucket.
	name string

	// sse is the default server side encryption setting.
	sse *string

	// svc is the S3API client.
	svc s3iface.S3API
}

type Reader struct {
	Key string
	svc Service
}

// New returns a pointer to a new Service.
// ACL is bucket-owner-full-control by default, but can be changed with SetACL.
// SSE is AES256 by default, but can be changed with SetSSE.
func New(name string, region, roleARN *string) *Service {
	return &Service{
		acl:  aws.String("bucket-owner-full-control"),
		name: name,
		sse:  aws.String("AES256"),
		svc:  s3.New(aidews.Session(region, roleARN)),
	}
}

// Put puts the content to the bucket at the key.
func (svc *Service) Put(key string, content io.Reader) (*s3.PutObjectOutput, error) {
	in := &s3.PutObjectInput{
		ACL:                  svc.acl,
		Body:                 aws.ReadSeekCloser(content),
		Bucket:               aws.String(svc.name),
		Key:                  aws.String(key),
		ServerSideEncryption: svc.sse,
	}
	return svc.svc.PutObject(in)
}

func (r *Reader) Read() (*io.ReadCloser, error) {
	return read(r.svc.name, r.Key, r.svc.svc)
}

// Read gets the object from the bucket at the key.
func (svc *Service) Read(key string) (*io.ReadCloser, error) {
	return read(svc.name, key, svc.svc)
}

// ReadUnmarshal gets the object from the bucket at the key and unmarshals into out.
func (svc *Service) ReadUnmarshal(key string, out interface{}) error {
	obj, err := svc.Read(key)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(*obj)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

// ListObjects list the requested number of items in a bucket
func (svc *Service) ListObjects(maxObjects uint64) ([]Reader, error) {
	input := &s3.ListObjectsInput{
		Bucket:  aws.String(svc.name),
		MaxKeys: aws.Int64(int64(maxObjects)),
	}

	result, err := svc.svc.ListObjects(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				return nil, aerr
			default:
				return nil, aerr
			}
		} else {
			return nil, err
		}
		return nil, err
	}

	contents := result.Contents
	var readers []Reader
	for _, v := range contents {
		reader := Reader{Key: *v.Key, svc: *svc}
		readers = append(readers, reader)
	}

	return readers, nil
}

// SetACL sets the ACL with which the objects will be stored.
func (svc *Service) SetACL(v *string) {
	svc.acl = v
}

// SetSSE sets the server side encryption string for the bucket.
func (svc *Service) SetSSE(v *string) {
	svc.sse = v
}

func read(name string, key string, s3 s3iface.S3API) (*io.ReadCloser, error){
	in := &s3.GetObjectInput{
		Bucket: aws.String(name),
	}
	in.SetKey(key)
	res, err := s3.GetObject(in)
	if err != nil {
		return nil, err
	}
	return &res.Body, nil
}
