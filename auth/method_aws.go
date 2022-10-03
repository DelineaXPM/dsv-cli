package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func buildAwsParams(awsProfile string) (*requestBody, error) {
	opts := session.Options{
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
	}

	if awsProfile != "" {
		opts.Profile = awsProfile
	}

	sess, err := session.NewSessionWithOptions(opts)
	if err != nil {
		return nil, fmt.Errorf("create aws session: %w", err)
	}
	stsClient := sts.New(sess)
	r, _ := stsClient.GetCallerIdentityRequest(nil)
	r.Sign()

	headers, err := json.Marshal(r.HTTPRequest.Header)
	if err != nil {
		return nil, fmt.Errorf("marshaling request headers: %w", err)
	}
	body, err := io.ReadAll(r.HTTPRequest.Body)
	if err != nil {
		return nil, fmt.Errorf("reading request body: %w", err)
	}

	hString := base64.StdEncoding.EncodeToString(headers)
	bString := base64.StdEncoding.EncodeToString(body)

	data := &requestBody{
		GrantType:  authTypeToGrantType[FederatedAws],
		AwsHeaders: hString,
		AwsBody:    bString,
	}
	return data, nil
}
