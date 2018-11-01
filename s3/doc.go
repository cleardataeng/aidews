// Package s3 is an aide for simplifying calls to retrieve and store
// objects in AWS S3 buckets.
//
// The use of type Server is deprecated and will be removed. For
// new development use the other types documented.
//
// Example
//
// s := s3.New(aidews.Session("us-west-2", nil))
// b := Bucket{
// 	Name: "my-bukcet",
// 	Svc:  s,
// }
// o := Object{
// 	Bucket: b,
// 	Key:    "some/object/path",
// }
// b, err := ioutil.ReadAll(rc)
// if err != nil {
// 	panic(err)
// }
// fmt.Println(string(b))
package s3
