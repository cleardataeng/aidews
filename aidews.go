// Package aidews provides utility helpers for interacting with the AWS API and the
// AWS Go SDK.
package aidews

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

// Session returns an aws session.
// The region and role_arn parameters are optional. If neither are given the
// session returned is built with a blank config. If region is given, the config
// used to get the session includes the region. If role_arn given, we first STS,
// then get a session in that region using the credentials from the STS call.
//
// All Sessions are constructed using the SharedConfigEnable setting allowing
// the use of local credential resolution.
func Session(region, roleARN *string) *session.Session {
	cfg := aws.Config{
		Region: region,
	}
	if roleARN != nil {
		creds := stscreds.NewCredentials(
			sessionWithConfig(cfg),
			*roleARN,
		)
		cfg.Credentials = creds
	}
	return sessionWithConfig(cfg)
}

// SessionWithConfig returns an aws session.
// The role_arn parameter is optional. If not given given the
// session returned is built with the config passed in. If role_arn given, we first STS,
// then get a session with those using the credentials added to the passed in config.
//
// All Sessions are constructed using the SharedConfigEnable setting allowing
// the use of local credential resolution.
func SessionWithConfig(cfg aws.Config, roleARN *string) *session.Session {
	if roleARN != nil {
		creds := stscreds.NewCredentials(
			sessionWithConfig(cfg),
			*roleARN,
		)
		cfg.Credentials = creds
	}
	return sessionWithConfig(cfg)
}

func sessionWithConfig(cfg aws.Config) *session.Session {
	return session.Must(session.NewSessionWithOptions(session.Options{
		Config:            cfg,
		SharedConfigState: session.SharedConfigEnable,
	}))
}
