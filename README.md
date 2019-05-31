aidews
======

aidews is a helper (aide) for AWS. [aws-sdk-go](https://github.com/aws/aws-sdk-go) is fantastic, but this simplifies some of its uses.

## Session
Session is the backbone helper of aidews and makes getting an [aws-sdk-go] session straight-forward

The `region` and `role_arn` parameters are optional. If neither are given the
session returned is built with a blank config. If region is given, the config
used to get the session includes the region. If role_arn given, we first STS,
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
adiews apigateway package provides helpers for making signed requests to api gateways

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
