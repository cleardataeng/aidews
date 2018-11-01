package s3

import (
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/cleardataeng/aidews"
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

// Read gets the object from the bucket at the key.
func (svc *Service) Read(key string) (*io.ReadCloser, error) {
	in := &s3.GetObjectInput{
		Bucket: aws.String(svc.name),
	}
	in.SetKey(key)
	res, err := svc.svc.GetObject(in)
	if err != nil {
		return nil, err
	}
	return &res.Body, nil
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

// SetACL sets the ACL with which the objects will be stored.
func (svc *Service) SetACL(v *string) {
	svc.acl = v
}

// SetSSE sets the server side encryption string for the bucket.
func (svc *Service) SetSSE(v *string) {
	svc.sse = v
}
