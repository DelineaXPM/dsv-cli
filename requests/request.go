package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"

	cst "thy/constants"
	"thy/errors"
	"thy/format"
	"thy/version"

	"github.com/spf13/viper"
)

var (
	MarshalingError   = errors.NewS("Failed to marshal response")
	UnmarshalingError = errors.NewS("Failed to unmarshal response")
)

type Header interface {
	Set(key, value string)
}

type clientMock struct {
	doRequestFunc    func(method string, uri string, body interface{}) ([]byte, *errors.ApiError)
	doRequestOutFunc func(method string, uri string, body interface{}, dataOut interface{}) *errors.ApiError
	setCredsFunc     func(rh Header)
}

func NewMockedClient(doRequestFunc func(method string, uri string, body interface{}) ([]byte, *errors.ApiError),
	doRequestOutFunc func(method string, uri string, body interface{}, dataOut interface{}) *errors.ApiError,
	setCredsFunc func(rh Header)) Client {
	return &clientMock{
		doRequestFunc:    doRequestFunc,
		doRequestOutFunc: doRequestOutFunc,
		setCredsFunc:     setCredsFunc,
	}
}
func (c *clientMock) DoRequest(method string, uri string, body interface{}) ([]byte, *errors.ApiError) {
	return c.doRequestFunc(method, uri, body)
}

func (c *clientMock) DoRequestOut(method string, uri string, body interface{}, dataOut interface{}) *errors.ApiError {
	return c.doRequestOutFunc(method, uri, body, dataOut)
}

func (c *clientMock) SetCreds(rh Header) {
	c.setCredsFunc(rh)
}

type Client interface {
	DoRequest(method string, uri string, body interface{}) ([]byte, *errors.ApiError)
	DoRequestOut(method string, uri string, body interface{}, dataOut interface{}) *errors.ApiError
	SetCreds(rh Header)
}

type httpClient struct{}

func NewHttpClient() Client {
	return &httpClient{}
}

func (c *httpClient) SetCreds(rh Header) {
	rh.Set("Authorization", viper.GetString(cst.NounToken))
}

func (c *httpClient) DoRequest(method string, uri string, body interface{}) ([]byte, *errors.ApiError) {
	resp, err := c.sendRequest(method, uri, body)
	if err != nil {
		return nil, errors.NewS("failed to send API request")
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, errors.NewS("malformed api response")
	}
	return getResponse(b, resp.StatusCode)
}

func (c *httpClient) DoRequestOut(method string, uri string, body interface{}, dataOut interface{}) *errors.ApiError {
	resp, err := c.sendRequest(method, uri, body)
	if err != nil {
		return errors.NewS("failed to send API request")
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if success := statusCodeIsSuccess(resp.StatusCode); !success {
		return errors.NewS(string(bodyBytes))
	}
	if len(bodyBytes) == 0 {
		return nil
	}
	if err := json.Unmarshal(bodyBytes, &dataOut); err != nil {
		return UnmarshalingError
	}
	return nil
}

func (c *httpClient) sendRequest(method string, uri string, body interface{}) (*http.Response, error) {
	var req *http.Request
	var err error

	if body == nil {
		req, err = http.NewRequest(method, uri, nil)
	} else if d, ok := body.([]byte); ok {
		req, err = http.NewRequest(method, uri, bytes.NewReader(d))
	} else {
		serialized, serr := json.Marshal(body)
		if serr != nil {
			return nil, errors.NewS("error serializing request body")
		}
		req, err = http.NewRequest(method, uri, bytes.NewReader(serialized))
	}
	if err != nil {
		return nil, errors.NewS("error creating api request")
	}
	c.SetCreds(&req.Header)
	req.Header.Set("Content-Type", "application/json")
	if version.Version != "undefined" {
		agent := fmt.Sprintf("%s-%s-%s-%s", cst.CmdRoot, version.Version, runtime.GOOS, runtime.GOARCH)
		req.Header.Set("User-Agent", agent)
	}
	return http.DefaultClient.Do(req)
}

func statusCodeIsSuccess(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

func getResponse(bodyBytes []byte, statusCode int) ([]byte, *errors.ApiError) {
	if !statusCodeIsSuccess(statusCode) {
		return nil, errors.NewS(string(bodyBytes))
	} else {
		if len(bodyBytes) == 0 {
			return bodyBytes, nil
		}
		var unmarshalled map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &unmarshalled); err != nil {
			return nil, UnmarshalingError
		} else {
			data := unmarshalled["data"]

			// If there is more than just `data`, marshal and return everything.
			if len(unmarshalled) > 1 || data == nil {
				res, err := format.JsonMarshal(unmarshalled)
				if err != nil {
					return nil, MarshalingError
				} else {
					return res, nil
				}
			} else {
				if marshalled, err := format.JsonMarshal(data); err != nil {
					return nil, MarshalingError
				} else {
					return marshalled, nil
				}
			}
		}
	}
}
