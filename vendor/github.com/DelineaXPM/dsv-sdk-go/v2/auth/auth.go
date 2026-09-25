package auth

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Provider int64

const (
	CLIENT Provider = iota
	AWS
	GCP
	AZURE
)

var errAwsSession = errors.New("failed to create aws session")

type Config struct {
	Profile  string
	Provider Provider
}

type authorization struct {
	config            Config
	getCallerIdentity func() *request.Request
	sess              *session.Session
}

func New(config Config) (*authorization, error) {
	if config.Profile == "" {
		config.Profile = "default"
	}

	ath := &authorization{}
	if config.Provider == AWS {
		opts := session.Options{
			SharedConfigState:       session.SharedConfigEnable,
			AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
			Profile:                 config.Profile,
		}

		sess, err := session.NewSessionWithOptions(opts)
		if err != nil {
			return nil, errAwsSession
		}
		ath.sess = sess
		ath.getCallerIdentity = ath.getCallerIdentityRequest
	}

	ath.config = config

	return ath, nil
}
