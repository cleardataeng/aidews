aidews
======

aidews is a helper (aide) for AWS. [aws-sdk-go](https://github.com/aws/aws-sdk-go) is fantastic, but this simplifies some of its uses.

## Session
Session is the backbone helper of aidews and makes getting an aws-sdk-go session straight-forward

The `region` and `role_arn` parameters are optional. If neither are given the
session returned is built with a blank config. If region is given, the config
used to get the session includes the region. If `role_arn` given, we first STS,
then get a session in that region using the credentials from the STS call.
All Sessions are constructed using the SharedConfigEnable setting allowing
the use of local credential resolution.


``` go
// Session with no additional configuration
sess := aidews.Session(nil, nil)
```

``` go
// Session with region set
region := "us-west-2"
sess := aidews.Session(&region, nil)
```


``` go
// Session with role set, this will get a session assumed into the role passed in
role := "arn:aws:iam::{{accounttId}}:role/role_name"
sess := aidews.Session(nil, &role)
```

## apigateway
aidews apigateway package provides helpers for making signed requests to api gateways

``` go
// Create client
host, _ := url.Parse("apigatewayUrl") // url.Url for apigateway
role := "arn:aws:iam::{{accounttId}}:role/role_name" // role with access to execute api
region := "us-west-2" // region of gateway
client := apigateway.New(host, &region, &role)

// Get
queryString :=  map[string][]string{
	"hokey":    []string{"pokey"},
}
resp, err := client.Get("do/the", queryString)

// Put
body := struct{
    turnYourself string
}{
    turnYourself: "around",
}
resp, err := client.Put("hokey/pokey", body)

// Post
body := struct{
    thatsWhat string
}{
    thatsWhat: "its all about",
}
resp, err := client.Post("hokey/pokey", body)
```

If your favorite HTTP verb is not present in our helpers, you may use the Do function

``` go
//Do
req, _ := http.NewRequest("DELETE", "we/all/fall/down", nil)
resp, err := client.Do(req)
```

## dynamodb

Package dynamodb provides a DynamoDB wrapper object.

QueryPages queries the table and unmarshals each page of results to provided function.
Caller should provide an item of some type that will be populated and each item will be
returned to the provided pager function.
Provided pager function should take a single interface{} and assert the type of the item (e.g. Item),
and a boolean which will indicate whether this is the last item.
Provided pager function should return false if it wants to stop processing.
Example:

```go
items := Item
pager := func(out interface{}, last bool) bool {
  if out.(Item).Found { return false }
  return !last
}
if err := QueryPages(queryInput, item, pager); err != nil {
 return err
}
```

ScanPages scans the table and unmarshals each page of results to provided function.
Caller should provide an item of some type that will be populated and each item will be
returned to the provided pager function.
Provided pager function should take a single interface{} and assert the type of the item (e.g. Item),
and a boolean which will indicate whether this is the last item.
Provided pager function should return false if it wants to stop processing.
Example:

```go
items := Item
pager := func(out interface{}, last bool) bool {
  if out.(Item).Found { return false }
  return !last
}
if err := ScanPages(queryInput, item, pager); err != nil {
 return err
}
```

### Batch

Package batch provides an aide for dynamodb's BatchWriteItem function.
This aide wraps the complexities of building the batch and retrying unprocessed items,
at the cost of being able to only do 1 table at a time.

Use NewBatch() to get a Batchobject. SetTableName(), and then
use the object's Add() method to add as many dynamodb items as you want.

The object will add them to the queue in batches of 10 (so that's 1 AWS API call every 10 items).
After you are done adding items, call Send() to finish sending the items. (If you Put() 23 items,
20 will get sent automatically in 2 batches, but you need an explicit Send() to send the last 3.)
Example:

```go
for _, item := range items {
	 capacity, err := batch.Add(PutRequest, item)
}
batch.Send()
```

Tell Add() whether it's a PutRequest or a DeleteRequest, and pass either the item to be put
or the Key of the item to be deleted. Either way, pass a map[string]*dynamodb.AttributeValue{}
