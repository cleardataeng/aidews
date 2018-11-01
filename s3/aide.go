package s3

import (
	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// Bucket allows access and control of an S3 bucket.
type Bucket struct {
	// Name of the bucket.
	Name string

	// Svc is the S3API client.
	// If not provided, default credential resolution will take place.
	Svc s3iface.S3API
}

// Object allows access to S3 objects.
type Object struct {
	// ACL is the ACL for an object.
	// If empty "bucket-owner-full-control" will be used.
	ACL string

	// Bucket in which the object belongs.
	Bucket Bucket

	// Key at which the object is or shall be stored.
	Key string

	// SSE is the server side encryption setting.
	// If empty "AES256" will be used.
	SSE string
}

// Objects is a simple constructor helper function taking a list of
// strings, S3 keys, and generating []Object.
// The idea is to get a list of objects from the main AWS SDK, using
// the full blow input, then passing the returned keys to this function.
// From there, since each object is implementing io.Reader, you can
// easily Read from each.
func (b Bucket) Objects(keys []string) []Object {
	objs := []Object{}
	for _, k := range keys {
		o := Object{
			Bucket: b,
			Key:    k,
		}
		objs = append(objs, o)
	}
	return []Object{}
}

// Read satisfied the io.Reader interface for Object.
// Using this I can use Object anywhere an io.Reader is expected. Of
// course, you could just use s3.GetObject, and then Read from the Body,
// but the intent here is to simplify that process.
func (o Object) Read(b []byte) (n int, err error) {
	in := new(s3.GetObjectInput)
	in.SetBucket(o.Bucket.Name)
	in.SetKey(o.Key)
	res, err := o.Bucket.Svc.GetObject(in)
	if err != nil {
		return 0, err
	}
	return res.Body.Read(b)
}

// Write satisfies the io.Writer interface for Object.
// You can construct an Object complete with Key and Bucket, and
// optionally, ACL and SSE, then use that anywhere an io.Writer is
// expected. You won't have buffered write, since we are using the
// simplified upload. You will only get 0 bytes written for failures
// or the total length of bytes is success.
func (o Object) Write(p []byte) (n int, err error) {
	if o.ACL == "" {
		o.ACL = "bucket-owner-full-control"
	}
	if o.SSE == "" {
		o.SSE = "AES256"
	}
	in := &s3.PutObjectInput{
		ACL:                  aws.String(o.ACL),
		Body:                 bytes.NewReader(p),
		Bucket:               aws.String(o.Bucket.Name),
		Key:                  aws.String(o.Key),
		ServerSideEncryption: aws.String(o.SSE),
	}
	if _, err := o.Bucket.Svc.PutObject(in); err != nil {
		return 0, err
	}
	return len(p), nil
}
