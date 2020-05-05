package predictors_test

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"thy/auth"
	cst "thy/constants"
	"thy/errors"
	"thy/predictors"
	"thy/requests"
	"time"

	"github.com/posener/complete"
	flag "github.com/spf13/pflag"
	"github.com/thycotic-rd/viper"

	"github.com/stretchr/testify/assert"
)

func TestNewFlagValue(t *testing.T) {
	params := predictors.Params{
		Name:      "config.name1",
		Shorthand: "n",
		Usage:     "usage of name1",
		Default:   "x",
		ValueType: "bool",
		Global:    true,
		Hidden:    false,
	}
	f := predictors.NewFlagValue(params)
	assert.Equal(t, params.Name, f.Name)
	assert.Equal(t, params.Global, f.Global)
	assert.Equal(t, params.Usage, f.Usage)
	assert.Equal(t, params.Shorthand, f.Shorthand)
	assert.Equal(t, params.ValueType, f.Val.Type())
	assert.Equal(t, params.Default, f.Val.DefaultValue)
	asFlag := flag.Lookup("config-name1")
	assert.NotNil(t, asFlag)
	nilFlag := flag.Lookup("fjdsklfjsdl")
	assert.Nil(t, nilFlag)
}

func TestLongFlag(t *testing.T) {
	assert.Equal(t, "--flag", predictors.LongFlag("flag"))
}

func TestCmdFriendlyName(t *testing.T) {
	assert.Equal(t, "flag1-flag2-flag3", predictors.CmdFriendlyName("flag1.flag2-flag3"))
}

func TestNewStringPredictor(t *testing.T) {
	predWrappers := predictors.NewStringPredictor("flag1", "f", "f usage")
	assert.Equal(t, "flag1", predWrappers[predictors.LongFlag("flag1")].Flag.Name)
}

var secretPathPredictorCases = []struct {
	Args     complete.Args
	Expected []string
}{
	{
		Args: complete.Args{
			All: []string{
				"secret", "read", "xxx",
			},
			Last:          "xxx",
			LastCompleted: "read",
		},
		Expected: []string{},
	},
	{
		Args: complete.Args{
			All: []string{
				"secret", "read", "resources/us-east-1",
			},
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
		// Should not autocomplete a second path
		Args: complete.Args{
			All: []string{
				"secret", "read", "resources/us-east-1/secret1", "resources",
			},
			Last:          "resources",
			LastCompleted: "resources/us-east-1/secret1",
		},
		Expected: []string{},
	},
	{
		// Should not autocomplete non-starts-with match
		Args: complete.Args{
			All: []string{
				"secret", "read", "servers",
			},
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

	for i, c := range secretPathPredictorCases {
		t.Run(fmt.Sprintf("Case %d. Args(%s)", i, strings.Join(c.Args.All, " ")), func(t *testing.T) {

			c.Args.Last = c.Args.All[len(c.Args.All)-1]
			viper.Set(cst.Domain, "somedomain")
			viper.Set(cst.Tenant, "sometenant")

			doRequestOutFunc := func(method string, uri string, body interface{}, dataOut interface{}) *errors.ApiError {

				prefix := ""
				parsedUri, err := url.Parse(uri)
				if err != nil {
					return errors.New(err)
				}
				if strings.HasPrefix(parsedUri.Path, "/v1") {
					parsedUri.Path = parsedUri.Path[len("/v1"):]
				}
				if !strings.HasPrefix(parsedUri.Path, "/secrets/") {
					return nil
				} else {
					prefix = parsedUri.Path[len("/secrets/"):]
				}
				if strings.HasSuffix(prefix, "::listpaths") {
					prefix = prefix[0 : len(prefix)-len("::listpaths")]
				}
				prefix = strings.Replace(prefix, "/", ":", -1)
				filteredList := make([]string, 0, 30)
				for _, r := range listResources {
					if strings.HasPrefix(r, prefix) {
						filteredList = append(filteredList, r)
					}
				}
				predictout := dataOut.(*predictors.PathAutocompleteResult)
				predictout.Data = filteredList
				return nil
			}

			pred := predictors.NewSecretPathPredictor(func() auth.Authenticator { return auth.GetTokenFunc(authFunc) },
				func() requests.Client { return requests.NewMockedClient(nil, doRequestOutFunc, nil) })

			predictions := pred.Predict(c.Args)
			assert.Equal(t, c.Expected, predictions)

		})
	}

}

func TestPrefixFilePredictor(t *testing.T) {
	args := complete.Args{}
	args.Last = "@"
	pred := predictors.NewPrefixFilePredictor("*")
	preds := pred.Predict(args)
	assert.NotEqual(t, 0, len(preds))

	args.Last = ""
	preds = pred.Predict(args)
	assert.Equal(t, 0, len(preds))

	args.Last = "@"
	pred = predictors.NewPrefixFilePredictor("*.fjksljflds")
	preds = pred.Predict(args)
	assert.Equal(t, 0, len(preds))
}
