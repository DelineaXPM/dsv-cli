package requests_test

import (
	"net/http"
	"testing"

	"github.com/DelineaXPM/dsv-cli/requests"

	"github.com/shurcooL/graphql"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestDoRequest(t *testing.T) {
	c := requests.NewGraphClient()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	jsonResponse := []byte(`{"data": {"UserName": "json"}`)
	httpmock.RegisterResponder(http.MethodPost, "https://localhost:8088",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewBytesResponse(200, jsonResponse)
			return resp, nil
		},
	)

	var query struct {
		Me struct {
			UserName graphql.String
		}
	}
	_, err := c.DoRequest("https://localhost:8088", &query, map[string]interface{}{})
	assert.NotNil(t, err)
}
