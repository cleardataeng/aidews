// Package aidews provides utility helpers for interacting with the AWS API and the
// AWS Go SDK.
package aidews

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
)

func session(session *session.Session, region, roleARN *string) *session.Session {
	cfg := aws.Config{
		Region: region,
	}
	if session == nil {
		session = sessionWithConfig(cfg)
	}
	if roleARN != nil {
		creds := stscreds.NewCredentials(
			session,
			*roleARN,
		)
		cfg.Credentials = creds
	}
	return sessionWithConfig(cfg)
}

// Session returns an aws session.
// The region and role_arn parameters are optional. If neither are given the
// session returned is built with a blank config. If region is given, the config
// used to get the session includes the region. If role_arn given, we first STS,
// then get a session in that region using the credentials from the STS call.
//
// All Sessions are constructed using the SharedConfigEnable setting allowing
// the use of local credential resolution.
func Session(region, roleARN *string) *session.Session {
	return session(nil, region, roleARN)
}

// SessionHop returns an aws session constructed from a given Session.
// This is very similar to Session, but allows hopping (assume role) from a given
// session, to the next destination role. Using SessionHop, a program can assume
// role any number of times.
//
// For example:
// start := Session(region, startingRoleARN)
// hop1 := SessionHop(start, region, hop1ARN)
// hop2 := SessionHop(hop1, region, hop2ARN)
// destination := SessionHop(hop2, region, destARN)
func SessionHop(s *session.Session, region, roleARN *string) *session.Session {
	return session(s, region, roleARN)
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
