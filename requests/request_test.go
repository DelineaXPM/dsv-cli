package requests_test

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.com/DelineaXPM/dsv-cli/requests"
)

func TestHttpClient_DoRequest(t *testing.T) {
	c := requests.NewHttpClient()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	jsonResponse := map[string]interface{}{
		"data": "ok",
	}
	body := map[string]interface{}{
		"data": "ok",
	}

	httpmock.RegisterResponder(http.MethodPost, "https://localhost:8088",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, jsonResponse)
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)
	_, err := c.DoRequest(http.MethodPost, "https://localhost:8088", body)
	assert.Nil(t, err)

}

func TestHttpClient_DoRequestOut(t *testing.T) {
	c := requests.NewHttpClient()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	jsonResponse := map[string]interface{}{
		"data": "ok",
	}

	httpmock.RegisterResponder(http.MethodGet, "https://localhost:8088",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, jsonResponse)
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)
	result := map[string]interface{}{}
	err := c.DoRequestOut(http.MethodGet, "https://localhost:8088", nil, &result)
	assert.Nil(t, err)
	assert.Equal(t, result["data"], jsonResponse["data"])
}

func TestHttpClient_DoRequestInvalidBody(t *testing.T) {
	c := requests.NewHttpClient()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	jsonResponse := map[string]interface{}{
		"data": "ok",
	}
	//need somethign that won't unmarshal
	body := make(chan int)

	httpmock.RegisterResponder(http.MethodPost, "https://localhost:8088",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, jsonResponse)
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)
	result := map[string]interface{}{}
	err := c.DoRequestOut(http.MethodPost, "https://localhost:8088", body, &result)
	assert.NotNil(t, err)
}
