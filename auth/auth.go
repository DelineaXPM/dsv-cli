package auth

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	cst "thy/constants"
	"thy/errors"
	"thy/paths"
	"thy/requests"
	"thy/store"

	"github.com/spf13/viper"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ../tests/fake/fake_authenticator.go . Authenticator

const (
	leewaySecondsTokenExp   = 10
	refreshTokenLifeSeconds = 60 * 60 * 720
)

// Note that this global error variable is of type *ApiError, not regular error.
var KeyfileNotFoundError = errors.NewS("failed to find the encryption key")

// AuthType is the type of authentication.
type AuthType string

// Types of supported authentication.
const (
	Password         = AuthType("password")
	Refresh          = AuthType("refresh")
	ClientCredential = AuthType("clientcred")
	Certificate      = AuthType("cert")
	FederatedThyOne  = AuthType("thy-one")
	FederatedAws     = AuthType("aws")
	FederatedAzure   = AuthType("azure")
	FederatedGcp     = AuthType("gcp")
	Oidc             = AuthType("oidc")
)

// authTypeToGrantType maps authentication type to grant type which will be sent to DSV.
var authTypeToGrantType = map[AuthType]string{
	Password:         "password",
	Refresh:          "refresh_token",
	ClientCredential: "client_credentials",
	Certificate:      "certificate",
	FederatedThyOne:  "oidc",
	FederatedAws:     "aws_iam",
	FederatedAzure:   "azure",
	FederatedGcp:     "gcp",
	Oidc:             "oidc",
}

// authTypeToCachePrefix maps authentication type to cache key prefix.
// Note: prefix must start with "token".
var authTypeToCachePrefix = map[AuthType]string{
	Password:         "token-password-",
	Refresh:          "token-password-", // Refresh and Password are stored the same.
	ClientCredential: "token-clientcred-",
	Certificate:      "token-cert-",
	FederatedThyOne:  "token-thy-one-",
	FederatedAws:     "token-aws-",
	FederatedAzure:   "token-azure-",
	FederatedGcp:     "token-gcp-",
	Oidc:             "token-oidc-",
}

func getTokenCacheKey(a AuthType, tenant string, profile string) string {
	prefix := authTypeToCachePrefix[a]
	key := prefix + tenant + "-" + profile
	return key
}

// Authenticator is the interface used for authentication funcs.
type Authenticator interface {
	GetToken() (*TokenResponse, *errors.ApiError)
	WipeCachedTokens() *errors.ApiError
}

type authenticator struct {
	store         store.Store
	requestClient requests.Client
}

// NewAuthenticatorDefault gets a new default authenticator.
func NewAuthenticatorDefault() Authenticator {
	st := viper.GetString(cst.StoreType)
	if s, err := store.GetStore(st); err != nil {
		panic(err)
	} else {
		return &authenticator{s, requests.NewHttpClient()}
	}
}

// NewAuthenticator returns a new authenticator.
func NewAuthenticator(store store.Store, client requests.Client) Authenticator {
	return &authenticator{store, client}
}

func (a *authenticator) GetToken() (*TokenResponse, *errors.ApiError) {
	authType := viper.GetString(cst.AuthType)

	at := Password
	if authType != "" {
		at = AuthType(authType)

		if _, ok := authTypeToGrantType[at]; !ok {
			return nil, errors.NewF("unknown authentication type %q", authType)
		}

		if _, ok := authTypeToCachePrefix[at]; !ok {
			return nil, errors.NewF("unknown authentication type %q", authType)
		}
	}

	skipCache := viper.GetBool(cst.AuthSkipCache)
	var cacheKey string
	if !skipCache {
		tenant := viper.GetString(cst.Tenant)
		profile := viper.GetString(cst.Profile)
		if profile == "" {
			profile = cst.DefaultProfile
		}
		cacheKey = getTokenCacheKey(at, tenant, profile)
	}

	return a.getToken(at, cacheKey)
}

// WipeCachedTokens removes all cached tokens for the current profile.
func (a *authenticator) WipeCachedTokens() *errors.ApiError {
	profile := viper.GetString(cst.Profile)
	if profile == "" {
		profile = cst.DefaultProfile
	}
	tokens, err := a.store.List("token")
	if err != nil {
		return err
	}
	for _, t := range tokens {
		if strings.HasSuffix(t, profile) {
			err = a.store.Delete(t)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *authenticator) getToken(at AuthType, cacheKey string) (*TokenResponse, *errors.ApiError) {
	if cacheKey != "" {
		tr := &TokenResponse{}
		err := a.store.Get(cacheKey, tr)
		if err != nil {
			return nil, err
		}

		if !tr.IsNil() {
			// lifetime defines how many seconds passed since token was fetched.
			lifetime := int64(time.Now().UTC().Sub(tr.Granted).Seconds())

			// Add some space for actions.
			lifetime = lifetime + leewaySecondsTokenExp

			// If access token is still valid, return it.
			if (lifetime - tr.ExpiresIn) <= 0 {
				return tr, nil
			}

			// If refresh token is present and still valid, use it to get a new token.
			if tr.RefreshToken != "" && ((lifetime - refreshTokenLifeSeconds) <= 0) {
				log.Print("Token expired but valid refresh token found. Attempting to refresh.")

				data := &requestBody{
					GrantType:    authTypeToGrantType[Refresh],
					RefreshToken: tr.RefreshToken,
				}

				if tr, err := a.fetchTokenVault(Refresh, data); err != nil {
					log.Printf("Refresh authentication failed: %s", err.Error())
				} else {
					log.Printf("Refresh authentication succeeded.")
					if err := a.store.Store(cacheKey, tr); err != nil {
						return nil, err.Grow("Failed caching token.")
					}
					return tr, nil
				}
			} else {
				log.Printf("Refresh token expired. Attempting to reauthenticate.")
			}
		}
	}

	reqBody, stdErr := a.newRequestBody(at)
	if stdErr != nil {
		return nil, errors.New(stdErr).Grow(
			fmt.Sprintf("Failed to build token request for %s based auth:", at),
		)
	}

	if err := reqBody.validate(at); err != nil {
		return nil, errors.New(err)
	}

	tr, err := a.fetchTokenVault(at, reqBody)
	if err != nil {
		return tr, err.Grow(fmt.Sprintf("Failed to authenticate with auth type '%s'. Please check parameters and try again", at))
	}

	log.Printf("%s authentication succeeded.\n", strings.Title(string(at)))

	if cacheKey != "" {
		if err := a.store.Store(cacheKey, tr); err != nil {
			return nil, err.Grow("Failed caching token")
		}
	}
	return tr, nil
}

type requestBody struct {
	GrantType          string `json:"grant_type"`
	Username           string `json:"username,omitempty"`
	Provider           string `json:"provider,omitempty"`
	Password           string `json:"password,omitempty"`
	AuthClientID       string `json:"client_id,omitempty"`
	AuthClientSecret   string `json:"client_secret,omitempty"`
	RefreshToken       string `json:"refresh_token,omitempty"`
	AwsBody            string `json:"aws_body,omitempty"`
	AwsHeaders         string `json:"aws_headers,omitempty"`
	Jwt                string `json:"jwt,omitempty"`
	AuthorizationCode  string `json:"authorization_code,omitempty"`
	CallbackUrl        string `json:"callback_url,omitempty"`
	State              string `json:"state,omitempty"`
	CertChallengeID    string `json:"cert_challenge_id,omitempty"`
	DecryptedChallenge string `json:"decrypted_challenge,omitempty"`
}

func (a *authenticator) newRequestBody(at AuthType) (*requestBody, error) {
	var data *requestBody
	var stdErr error
	switch at {
	case Password:
		data, stdErr = buildPasswordParams()

	case ClientCredential:
		data, stdErr = buildClientcredParams()

	case Certificate:
		cert := viper.GetString(cst.AuthCert)
		privKey := viper.GetString(cst.AuthPrivateKey)
		data, stdErr = a.buildCertParams(cert, privKey)

	case FederatedThyOne:
		fallthrough
	case Oidc:
		provider := viper.GetString(cst.AuthProvider)
		callback := viper.GetString(cst.Callback)
		data, stdErr = a.buildOIDCParams(at, provider, callback)

	case FederatedAws:
		awsProfile := viper.GetString(cst.AwsProfile)
		data, stdErr = buildAwsParams(awsProfile)

	case FederatedAzure:
		data, stdErr = buildAzureParams()

	case FederatedGcp:
		token := viper.GetString(cst.GcpToken)
		gcpAuthType := viper.GetString(cst.GcpAuthType)
		data, stdErr = buildGcpParams(token, gcpAuthType)

	default:
		stdErr = fmt.Errorf("unexpected authentication type %q", at)
	}

	if stdErr != nil {
		return nil, stdErr
	}
	return data, nil
}

func (a *authenticator) fetchTokenVault(at AuthType, data *requestBody) (*TokenResponse, *errors.ApiError) {
	response := &TokenResponse{}
	uri := paths.CreateURI(cst.NounToken, nil)
	if err := a.requestClient.DoRequestOut(http.MethodPost, uri, data, response); err != nil {
		return nil, err
	} else if response.IsNil() {
		return nil, errors.NewS("Empty token")
	}
	response.Granted = time.Now().UTC()
	return response, nil
}

type TokenResponse struct {
	Token        string    `json:"accessToken"`
	TokenType    string    `json:"tokenType"`
	ExpiresIn    int64     `json:"expiresIn"`
	RefreshToken string    `json:"refreshToken"`
	Granted      time.Time `json:"granted"`
}

func (r *TokenResponse) IsNil() bool {
	return r.Token == "" && r.RefreshToken == ""
}

func (r *requestBody) validate(at AuthType) error {
	ref := reflect.Indirect(reflect.ValueOf(r))
	for _, k := range paramSpecDict[at] {
		f := ref.FieldByName(k.PropName)
		if f.String() == "" {
			return errors.NewF("--%s must be set", k.ArgName)
		}
	}
	return nil
}

type paramSpec struct {
	PropName string
	ArgName  string
}

var paramSpecDict = map[AuthType][]paramSpec{
	Password: {
		{PropName: "Password", ArgName: cst.Password},
		{PropName: "Username", ArgName: cst.Username},
	},
	ClientCredential: {
		{PropName: "AuthClientID", ArgName: cst.AuthClientID},
		{PropName: "AuthClientSecret", ArgName: cst.AuthClientSecret},
	},
	FederatedAws: {
		{PropName: "AwsBody", ArgName: ""},
		{PropName: "AwsHeaders", ArgName: ""},
	},
	Refresh: {
		{PropName: "RefreshToken", ArgName: cst.RefreshToken},
	},
	FederatedAzure: {{PropName: "JwtToken", ArgName: "jwt"}},
	FederatedGcp:   {{PropName: "JwtToken", ArgName: "jwt"}},
}
