aidews
======

aidews is a helper (aide) for AWS. [aws-sdk-go] is fantastic, but this simplifies some of its uses.

## Session
Session is the backbone helper of aidews and makes getting an [aws-sdk-go] session straight-forword

The `region` and `role_arn` parameters are optional. If neither are given the
session returned is built with a blank config. If region is given, the config
used to get the session includes the region. If role_arn given, we first STS,
then get a session in that region using the credentials from the STS call.
All Sessions are constructed using the SharedConfigEnable setting allowing
the use of local credential resolution.


```
// Session with no additional configuration

sess := aidews.Session(nil, nil)
```

```
// Session with region set
region := "us-west-2"
sess := aidews.Session(&region, nil)
```


```
// Session with role set, this will get a session assumed into the role passed in
role := "arn:aws:iam::{{accounttId}}:role/accounts"
sess := aidews.Session(nil, &role)
```

