package auth

import (
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sts"
)

func (a *authorization) GetSTSHeaderAndBody() (string, string, error) {
	r := a.getCallerIdentity()
	headers, err := json.Marshal(r.HTTPRequest.Header)
	if err != nil {
		return "", "", err
	}

	body, err := io.ReadAll(r.HTTPRequest.Body)
	if err != nil {
		return "", "", err
	}

	return base64.StdEncoding.EncodeToString(headers), base64.StdEncoding.EncodeToString(body), nil
}

func (a *authorization) getCallerIdentityRequest() *request.Request {
	r, _ := sts.New(a.sess).GetCallerIdentityRequest(nil)
	_ = r.Sign()
	return r
}
