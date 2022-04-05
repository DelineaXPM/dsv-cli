package predictor

import (
	"net/http"
	"strings"

	"thy/auth"
	cst "thy/constants"
	"thy/paths"
	"thy/requests"

	"github.com/posener/complete"
	"github.com/spf13/viper"
)

type pathAutocompleteResult struct {
	Data []string `json:"data"`
}

// NewSecretPathPredictorDefault returns a new SecretPathPredictor
func NewSecretPathPredictorDefault() complete.Predictor {
	return &secretPathPredictor{
		authenticatorFunc: auth.NewAuthenticatorDefault,
		requestClientFunc: requests.NewHttpClient,
	}
}

// NewSecretPathPredictor returns a new SecretPathPredictor
func NewSecretPathPredictor(
	authenticatorFunc func() auth.Authenticator,
	requestClientFunc func() requests.Client,
) complete.Predictor {
	return &secretPathPredictor{
		authenticatorFunc: authenticatorFunc,
		requestClientFunc: requestClientFunc,
	}
}

type secretPathPredictor struct {
	authenticator     auth.Authenticator
	authenticatorFunc func() auth.Authenticator
	requestClient     requests.Client
	requestClientFunc func() requests.Client
}

func (p *secretPathPredictor) getAuthenticator() auth.Authenticator {
	if p.authenticator == nil {
		a := p.authenticatorFunc()
		p.authenticator = a
	}
	return p.authenticator
}

func (p *secretPathPredictor) getRequestClient() requests.Client {
	if p.requestClient == nil {
		c := p.requestClientFunc()
		p.requestClient = c
	}
	return p.requestClient
}

func (p *secretPathPredictor) Predict(a complete.Args) (prediction []string) {
	// NOTE : This is to prevent predict from predicting past first completed term.
	if len(a.All) > 0 && a.LastCompleted != "" && !strings.HasPrefix(a.LastCompleted, "-") {
		index := -1
		for i, arg := range a.All {
			if arg == a.LastCompleted {
				index = i - 1
			}
		}
		if index >= 0 {
			oneBefore := a.All[index]
			if oneBefore == "--path" || oneBefore == "read" {
				return []string{}
			}
		}
	}

	token := viper.Get("token")
	if token == nil {
		if tr, err := p.getAuthenticator().GetToken(); err != nil || tr == nil || tr.Token == "" {
			return []string{}
		} else {
			token = tr.Token
			viper.Set("token", token)
		}

	}
	uri := paths.CreateResourceURI(cst.NounSecrets, a.Last, cst.SuffixListPaths, true, nil)
	res := pathAutocompleteResult{}
	if err := p.getRequestClient().DoRequestOut(http.MethodGet, uri, nil, &res); err == nil {
		return preparePaths(res.Data)
	}
	return []string{}
}

func preparePaths(unprepared []string) []string {
	prepared := []string{}
	for _, p := range unprepared {
		friendlyTail := paths.GetURIPathFromInternalPath(p)
		if len(friendlyTail) > 0 {
			prepared = append(prepared, friendlyTail)
		}
	}
	return prepared
}
