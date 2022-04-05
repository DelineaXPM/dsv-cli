package predictor

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/posener/complete"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"thy/auth"
	cst "thy/constants"
	"thy/errors"
	"thy/fake"
	"thy/requests"
)

var secretPathPredictorCases = []struct {
	Args     complete.Args
	Expected []string
}{
	{
		Args: complete.Args{
			All:           []string{"secret", "read", "xxx"},
			Last:          "xxx",
			LastCompleted: "read",
		},
		Expected: []string{},
	},
	{
		Args: complete.Args{
			All:           []string{"secret", "read", "resources/us-east-1"},
			Last:          "resources/us-east-1",
			LastCompleted: "read",
		},
		Expected: []string{
			"resources/us-east-1/secret1",
			"resources/us-east-1/secret4",
			"resources/us-east-1/secret10",
			"resources/us-east-1/resources/secret1",
		},
	},
	{
		Args: complete.Args{
			All:           []string{"secret", "read", "--path", "resources/us-east-1"},
			Last:          "resources/us-east-1",
			LastCompleted: "--path",
		},
		Expected: []string{
			"resources/us-east-1/secret1",
			"resources/us-east-1/secret4",
			"resources/us-east-1/secret10",
			"resources/us-east-1/resources/secret1",
		},
	},
	{
		// Should not autocomplete a second path
		Args: complete.Args{
			All:           []string{"secret", "read", "resources/us-east-1/secret1", "resources"},
			Last:          "resources",
			LastCompleted: "resources/us-east-1/secret1",
		},
		Expected: []string{},
	},
	{
		// Should not autocomplete non-starts-with match
		Args: complete.Args{
			All:           []string{"secret", "read", "servers"},
			Last:          "servers",
			LastCompleted: "read",
		},
		Expected: []string{
			"servers/us-east-1/secret5",
			"servers/us-east-1/resources/secret10",
		},
	},
}

var listResources = []string{
	"resources:us-east-1:secret1",
	"resources:us-east-1:secret4",
	"resources:us-east-1:secret10",
	"servers:us-east-1:secret5",
	"resources:us-east-2:secret5",
	"resources:us-east-3:secret5",
	"resources:us-east-1:resources:secret1",
	"servers:us-east-1:resources:secret10",
	"y:servers:1",
	"resources:us-west-2:secret5",
}

func TestSecretPathPredictor(t *testing.T) {
	authFunc := func() (*auth.TokenResponse, *errors.ApiError) {
		return &auth.TokenResponse{
			Token:        "token",
			RefreshToken: "refresh-token",
			ExpiresIn:    3600,
			Granted:      time.Now().UTC(),
		}, nil
	}

	const (
		tenantName        = "sometenant"
		expectedURIPrefix = "https://sometenant.secretsvaultcloud.com/v1/secrets/"
		expectedURISuffix = "::listpaths"
	)

	for i, c := range secretPathPredictorCases {
		t.Run(fmt.Sprintf("Case %d. Args(%s)", i, strings.Join(c.Args.All, " ")), func(t *testing.T) {
			viper.Set(cst.Tenant, tenantName)

			doRequestOutFunc := func(method string, uri string, body interface{}, dataOut interface{}) *errors.ApiError {
				if !strings.HasPrefix(uri, expectedURIPrefix) {
					t.Fatalf("missing expected prefix in uri: %s", uri)
				}
				if !strings.HasSuffix(uri, expectedURISuffix) {
					t.Fatalf("missing expected suffix in uri: %s", uri)
				}

				prefix := strings.TrimPrefix(uri, expectedURIPrefix)
				prefix = strings.TrimSuffix(prefix, expectedURISuffix)
				prefix = strings.ReplaceAll(prefix, "/", ":")

				filteredList := []string{}
				for _, r := range listResources {
					if strings.HasPrefix(r, prefix) {
						filteredList = append(filteredList, r)
					}
				}
				predictout := dataOut.(*pathAutocompleteResult)
				predictout.Data = filteredList
				return nil
			}

			fakeClient := &fake.FakeClient{
				DoRequestOutStub: doRequestOutFunc,
			}

			pred := NewSecretPathPredictor(
				func() auth.Authenticator { return auth.GetTokenFunc(authFunc) },
				func() requests.Client { return fakeClient },
			)

			predictions := pred.Predict(c.Args)
			assert.Equal(t, c.Expected, predictions)
		})
	}
}
