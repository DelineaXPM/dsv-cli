package vault

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/DelineaXPM/dsv-sdk-go/v2/auth"
)

const (
	defaultTLD         string = "com"
	defaultURLTemplate string = "https://%s.secretsvaultcloud.%s/v1/%s%s"
	dsvEnvVar          string = "DSV_AT"
)

var (
	errClientID     = errors.New("Credentials.ClientID must be set")
	errClientSecret = errors.New("Credentials.ClientSecret must be set")
	errTenant       = errors.New("tenant must be set")
)

// resourceMetadata are fields common to all complex resources
type resourceMetadata struct {
	ID, Description           string
	Created, LastModified     time.Time
	CreatedBy, LastModifiedBy string
	Version                   string
}

// simpleResourceMetadata are fields common to all simple resources
type simpleResourceMetadata struct {
	ID        string `json:"id"`
	Created   time.Time
	CreatedBy string
}

// ClientCredential contains the client_id and client_secret that the API will
// use to make requests
type ClientCredential struct {
	ClientID, ClientSecret string
}

// Configuration used to request an accessToken for the API
type Configuration struct {
	Credentials              ClientCredential
	Tenant, TLD, URLTemplate string
	Provider                 auth.Provider
}

// Vault provides access to secrets stored in Delinea DSV
type Vault struct {
	Configuration
}

//nolint:tagliatelle // the json is coming from an external API call
type TokenCache struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// New returns a Vault or an error if the Configuration is invalid
func New(config Configuration) (*Vault, error) {
	if config.Provider == auth.CLIENT {
		if config.Credentials.ClientID == "" {
			return nil, errClientID
		}
		if config.Credentials.ClientSecret == "" {
			return nil, errClientSecret
		}
	}

	if config.Tenant == "" {
		return nil, errTenant
	}
	if config.TLD == "" {
		config.TLD = defaultTLD
	}
	if config.URLTemplate == "" {
		config.URLTemplate = defaultURLTemplate
	}

	return &Vault{config}, nil
}

// accessResource uses the accessToken to access the API resource.
// It assumes an appropriate combination of method, resource, path and input.
func (v Vault) accessResource(method, resource, path string, input interface{}) ([]byte, error) {
	switch resource {
	case clientsResource, rolesResource, secretsResource:
	default:
		return nil, fmt.Errorf("unrecognized resource: %s", resource)
	}

	accessToken, err := v.getAccessToken()
	if err != nil {
		log.Print("[DEBUG] error getting accessToken: ", err)
		return nil, err
	}

	var body io.Reader
	if input != nil {
		data, err := json.Marshal(input)
		if err != nil {
			log.Print("[DEBUG] marshaling the request body to JSON:", err)
			return nil, err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, v.urlFor(resource, path), body)
	if err != nil {
		log.Printf("[DEBUG] creating req: %s /%s/%s: %s", method, resource, path, err)
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	switch method {
	case http.MethodPost, http.MethodPut:
		req.Header.Set("Content-Type", "application/json")
	}

	log.Printf("[DEBUG] calling %s", req.URL.String())

	data, err := handleResponse((&http.Client{}).Do(req))

	return data, err
}

type accessTokenRequest struct {
	GrantType string `json:"grant_type"`

	// Fields for "client_credentials" grant type.
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`

	// Fields for "aws_iam" grant type.
	AwsBody    string `json:"aws_body,omitempty"`
	AwsHeaders string `json:"aws_headers,omitempty"`

	// Fields for "Azure" grant type.
	Jwt string `json:"jwt,omitempty"`
}

//nolint:tagliatelle // the json is coming from an external API call
type accessTokenResponse struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"`
}

func (v Vault) setCacheAccessToken(value string, expiresIn int) bool {
	percentage := 0.9
	cache := TokenCache{}
	cache.AccessToken = value
	cache.ExpiresIn = (int(time.Now().Unix()) + expiresIn) - int(math.Floor(float64(expiresIn)*percentage))

	data, err := json.Marshal(cache)
	if err != nil {
		return false
	}
	_ = os.Setenv(dsvEnvVar, string(data))
	return true
}

func (v Vault) getCacheAccessToken() (string, bool) {
	data, ok := os.LookupEnv(dsvEnvVar)
	if !ok {
		_ = os.Setenv(dsvEnvVar, "")
		return "", ok
	}
	cache := TokenCache{}
	if err := json.Unmarshal([]byte(data), &cache); err != nil {
		return "", false
	}
	if time.Now().Unix() < int64(cache.ExpiresIn) {
		return cache.AccessToken, true
	}
	return "", false
}

// getAccessToken returns access token fetched from DSV.
//
//nolint:cyclop //function is not overly complex :)
func (v Vault) getAccessToken() (string, error) {
	accessToken, found := v.getCacheAccessToken()
	if found {
		return accessToken, nil
	}
	var rBody accessTokenRequest
	//nolint:exhaustive //not necessary
	switch v.Provider {
	case auth.AWS:
		auth, err := auth.New(auth.Config{Provider: auth.AWS})
		if err != nil {
			return "", err
		}
		header, body, err := auth.GetSTSHeaderAndBody()
		if err != nil {
			return "", err
		}

		rBody.GrantType = "aws_iam"
		rBody.AwsHeaders = header
		rBody.AwsBody = body
	case auth.AZURE:
		ath, _ := auth.New(auth.Config{Provider: auth.AZURE})
		data, err := ath.BuildAzureParams()
		if err != nil {
			return "", err
		}
		rBody.GrantType = data.GrantType
		rBody.Jwt = data.Jwt
	default:
		rBody.GrantType = "client_credentials"
		rBody.ClientID = v.Credentials.ClientID
		rBody.ClientSecret = v.Credentials.ClientSecret
	}
	request, err := json.Marshal(&rBody)
	if err != nil {
	}

	url := v.urlFor("token", "")
	response, err := handleResponse(http.Post(url, "application/json", bytes.NewReader(request)))
	if err != nil {
		return "", fmt.Errorf("fetching token: %w", err)
	}

	resp := &accessTokenResponse{}
	if err = json.Unmarshal(response, &resp); err != nil {
		return "", fmt.Errorf("unmarshaling token response: %w", err)
	}
	ok := v.setCacheAccessToken(resp.AccessToken, resp.ExpiresIn)
	if !ok {
		return "", fmt.Errorf("unable to cache access token")
	}
	return resp.AccessToken, nil
}

// urlFor the URL of the given resource and path in the current Vault
func (v Vault) urlFor(resource, path string) string {
	if path != "" {
		path = "/" + strings.TrimLeft(path, "/")
	}
	return fmt.Sprintf(v.URLTemplate, v.Tenant, v.TLD, resource, path)
}
