package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"time"

	cst "thy/constants"
	"thy/errors"
	"thy/version"

	"github.com/spf13/viper"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ../tests/fake/fake_client.go . Client

type Client interface {
	DoRequest(method string, uri string, body interface{}) ([]byte, *errors.ApiError)
	DoRequestOut(method string, uri string, body interface{}, dataOut interface{}) *errors.ApiError
}

type httpClient struct{}

func NewHttpClient() Client {
	return &httpClient{}
}

func (c *httpClient) DoRequest(method string, uri string, body interface{}) ([]byte, *errors.ApiError) {
	req, err := c.buildRequest(method, uri, body)
	if err != nil {
		return nil, err
	}
	respBytes, err := c.do(req)
	if err != nil {
		return nil, err
	}
	return respBytes, nil
}

func (c *httpClient) DoRequestOut(method string, uri string, body interface{}, dataOut interface{}) *errors.ApiError {
	respBytes, err := c.DoRequest(method, uri, body)
	if err != nil {
		return err
	}
	if len(respBytes) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBytes, &dataOut); err != nil {
		return errors.New(err).Grow("Failed to unmarshal response")
	}
	return nil
}

func (c *httpClient) buildRequest(method string, uri string, body interface{}) (*http.Request, *errors.ApiError) {
	var reqBody io.Reader
	if body == nil {
		reqBody = nil
	} else if d, ok := body.([]byte); ok {
		reqBody = bytes.NewReader(d)
	} else {
		serialized, serr := json.Marshal(body)
		if serr != nil {
			return nil, errors.New(serr).Grow("Error serializing request body")
		}
		reqBody = bytes.NewReader(serialized)
	}

	req, err := http.NewRequest(method, uri, reqBody)
	if err != nil {
		return nil, errors.New(err).Grow("Error creating API request")
	}

	agent := fmt.Sprintf("%s-%s-%s-%s", cst.CmdRoot, version.Version, runtime.GOOS, runtime.GOARCH)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", viper.GetString(cst.NounToken))
	req.Header.Set("User-Agent", agent)
	req.Header.Set("Delinea-DSV-Client", fmt.Sprintf(
		"cli-%s-%s/%s", version.Version, runtime.GOOS, runtime.GOARCH))

	return req, nil
}

func (c *httpClient) do(req *http.Request) ([]byte, *errors.ApiError) {
	log.Printf("-> %s %s", req.Method, req.URL)

	startTime := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.New(err).Grow("Failed to send API request")
	}

	log.Printf("<- %s %s | %s (took: %s)", req.Method, req.URL, resp.Status, time.Since(startTime))
	defer resp.Body.Close()

	bodyBytes, rerr := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(rerr).Grow("Malformed API response")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if len(bodyBytes) == 0 {
			return nil, errors.NewS("Error processing API response")
		}
		return nil, errors.NewS(string(bodyBytes)).WithResponse(resp)
	}

	return bodyBytes, nil
}
