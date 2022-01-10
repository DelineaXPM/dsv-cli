package auth_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"testing"
	"thy/auth"
	cst "thy/constants"
	"thy/errors"
	"thy/requests"
	"thy/store"
	"thy/utils/test_helpers"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var (
	tn            = "mocktenant"
	domain        = "alsdkfjasdfasdfasdfasdfasdfasdf.com"
	tokenEndpoint = fmt.Sprintf("https://%s.%s/v1/token", tn, domain)
	at            = "eyJhasldsflks.asdlkfjasldkfjlaskdjf.lsakdfjlaskdjf"
	rt            = "this-is-a-refresh-token"
	vtr           = map[string]interface{}{
		"status": 200,
		"response": map[string]interface{}{
			"accessToken":  at,
			"tokenType":    "bearer",
			"expiresIn":    3600,
			"refreshToken": rt,
		},
	}
	btr = map[string]interface{}{
		"status": 401,
		"response": map[string]interface{}{
			"code":    401,
			"message": "unable to authenticate",
		},
	}
)

func TestGetTokenKey(t *testing.T) {
	tests := []struct {
		in   string
		auth string
		want string
	}{
		{"refresh", "auth.gcp.service", "token-password--auth.gcp.service"},
		{"cert", "auth.gcp.service", "token-cert--auth.gcp.service"},
	}
	au := auth.AuthType("password")
	au.GetTokenKey("p")

	for _, tt := range tests {
		au := auth.AuthType(tt.in)
		key := au.GetTokenKey(tt.auth)
		if key != tt.want {
			t.Fatalf("Failed got  expected :: %v, got :: %v", tt.want, key)
		}

	}
}

//TODO:  need to refactor the code and rewrite
func TestGetToken(t *testing.T) {
	testCases := []struct {
		auth          string
		storeType     string
		want          *auth.TokenResponse
		expectedError error
	}{
		{"password", "none", nil, fmt.Errorf("error")},
		//{"azure", "none", nil, errors.New("error")},
		//{"gcp", "none", nil, errors.New("error")},
		{"aws", "none", nil, fmt.Errorf("error")},
		{"refresh", "none", nil, fmt.Errorf("error")},
		{"clientcred", "none", nil, fmt.Errorf("error")},
	}

	for _, tt := range testCases {
		authDef := auth.NewAuthenticatorDefault()
		viper.Set("auth.type", tt.auth)
		rsp, err := authDef.GetToken()

		if tt.expectedError != nil && err == nil || err != nil && tt.expectedError == nil {
			t.Fatalf("Failed got  expected :: %v, got :: %v", tt.expectedError, err)
		}

		if !reflect.DeepEqual(rsp, tt.want) {
			t.Fatalf("Failed got  expected :: %v, got :: %v", tt.want, rsp)
		}
	}
}

func TestGetToken_Password(t *testing.T) {
	storeType := "none"

	testCases := []struct {
		name           string
		auth           string
		password       string
		securePassword string
		apiResponse    map[string]interface{}
		expectedError  *errors.ApiError
	}{
		{
			"valid response should succeed",
			"password",
			"somePass12#",
			"",
			vtr,
			nil,
		},
		{
			"bad response should error",
			"password",
			"somePass12#",
			"",
			btr,
			errors.NewS("Failed to authenticate with auth type 'password'"),
		},
		{
			"should default to password auth",
			"",
			"somePass12#",
			"",
			vtr,
			nil,
		},
		{
			"missing encryption key should fail",
			"password",
			"",
			"someSecureEncryptedPass12",
			btr,
			errors.NewS("failed to find the encryption key"),
		},
		{
			"should fail on invalid auth type",
			"fastword",
			"somePass12#",
			"",
			btr,
			errors.NewS("failure to auth. Token cache key not found for authentication type fastword"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			st, err := store.GetStore(storeType)
			if err != nil {
				t.Fatalf("error getting store got :: %v", err)
			}
			authDef := auth.NewAuthenticator(st, requests.NewHttpClient())
			viper.Set("auth.type", tc.auth)
			viper.Set("auth.username", "admin")
			viper.Set("auth.password", tc.password)
			viper.Set("auth.securePassword", tc.securePassword)
			viper.Set("domain", domain)
			viper.Set("tenant", tn)

			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			registerResponse(tokenEndpoint, tc.apiResponse["status"].(int), http.MethodPost, tc.apiResponse["response"].(map[string]interface{}))

			rsp, err := authDef.GetTokenCacheOverride(false)

			if tc.expectedError != nil && err == nil || err != nil && tc.expectedError == nil {
				t.Fatalf("Failed got  expected :: %v, got :: %v", tc.expectedError, err)
			}

			if tc.expectedError != nil {
				assert.Contains(t, err.Error(), tc.expectedError.Error())
			}

			if err == nil {
				expectedResponse := tc.apiResponse["response"].(map[string]interface{})
				assert.Equal(t, expectedResponse["accessToken"], rsp.Token)
				assert.Equal(t, expectedResponse["refreshToken"], rsp.RefreshToken)
				assert.Equal(t, int64(expectedResponse["expiresIn"].(int)), rsp.ExpiresIn)
			}
		})
	}
}

func TestGetToken_SecurePassword(t *testing.T) {
	storeType := "file"

	expiredResponse := map[string]interface{}{
		"status": 200,
		"response": map[string]interface{}{
			"accessToken":  at,
			"tokenType":    "bearer",
			"expiresIn":    3,
			"refreshToken": rt,
		},
	}

	testCases := []struct {
		name          string
		auth          string
		clearStore    bool
		password      string
		apiResponse   map[string]interface{}
		expectedError *errors.ApiError
	}{
		{
			"should decrypt secure password",
			"password",
			true,
			"someSecureEncryptedPass12",
			expiredResponse,
			nil,
		},
		{
			"should get refresh token from cache",
			"password",
			false,
			"someSecureEncryptedPass12",
			vtr,
			nil,
		},
		{
			"should get access refresh token from cache",
			"password",
			false,
			"someSecureEncryptedPass12",
			vtr,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			st, err := store.GetStore(storeType)
			if err != nil {
				t.Fatalf("error getting store got :: %v", err)
			}
			if tc.clearStore {
				_ = st.Wipe(cst.TokenRoot + "-password-" + tn)
			}
			userName := "mock-user-soup"
			viper.Set("auth.type", tc.auth)
			viper.Set("auth.username", userName)
			viper.Set("domain", domain)
			viper.Set("tenant", tn)

			authDef := auth.NewAuthenticator(st, requests.NewHttpClient())
			_ = test_helpers.AddEncryptionKey(tn, userName, tc.password)
			securePass, passErr := auth.EncipherPassword(tc.password)
			assert.NoError(t, passErr)

			viper.Set("auth.password", "")
			viper.Set("auth.securePassword", securePass)

			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			registerResponse(tokenEndpoint, tc.apiResponse["status"].(int), http.MethodPost, tc.apiResponse["response"].(map[string]interface{}))

			rsp, err := authDef.GetToken()

			if tc.expectedError != nil && err == nil || err != nil && tc.expectedError == nil {
				t.Fatalf("Failed got  expected :: %v, got :: %v", tc.expectedError, err)
			}

			if tc.expectedError != nil {
				assert.Contains(t, err.Error(), tc.expectedError.Error())
			}

			if err == nil {
				expectedResponse := tc.apiResponse["response"].(map[string]interface{})
				assert.Equal(t, expectedResponse["accessToken"], rsp.Token)
				assert.Equal(t, expectedResponse["refreshToken"], rsp.RefreshToken)
				assert.Equal(t, int64(expectedResponse["expiresIn"].(int)), rsp.ExpiresIn)
			}
		})
	}
}

func TestGetToken_RefreshToken(t *testing.T) {
	storeType := "none"

	testCases := []struct {
		name          string
		auth          string
		refreshToken  string
		apiResponse   map[string]interface{}
		expectedError *errors.ApiError
	}{
		{
			"refresh token should succeed",
			"refresh",
			rt,
			vtr,
			nil,
		},
		{
			"bad response should error",
			"refresh",
			"bad-token",
			btr,
			errors.NewF("Refresh authentication failed. Please re-authenticate with password or other supported authentication type"),
		},
		{
			"refresh token should fail if not set",
			"refresh",
			"",
			vtr,
			errors.NewF("Refresh authentication failed: refreshtoken flag must be set"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			st, err := store.GetStore(storeType)
			if err != nil {
				t.Fatalf("error getting store got :: %v", err)
			}
			authDef := auth.NewAuthenticator(st, requests.NewHttpClient())
			viper.Set("auth.username", "")
			viper.Set("auth.password", "")
			viper.Set("auth.type", tc.auth)
			viper.Set("refreshtoken", tc.refreshToken)
			viper.Set("domain", domain)
			viper.Set("tenant", tn)

			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			registerResponse(tokenEndpoint, tc.apiResponse["status"].(int), http.MethodPost, tc.apiResponse["response"].(map[string]interface{}))

			rsp, err := authDef.GetTokenCacheOverride(false)

			if tc.expectedError != nil && err == nil || err != nil && tc.expectedError == nil {
				t.Fatalf("Failed got  expected :: %v, got :: %v", tc.expectedError, err)
			}

			if tc.expectedError != nil {
				assert.Contains(t, err.Error(), tc.expectedError.Error())
			}

			if err == nil {
				expectedResponse := tc.apiResponse["response"].(map[string]interface{})
				assert.Equal(t, expectedResponse["accessToken"], rsp.Token)
				assert.Equal(t, expectedResponse["refreshToken"], rsp.RefreshToken)
				assert.Equal(t, int64(expectedResponse["expiresIn"].(int)), rsp.ExpiresIn)
			}
		})
	}
}

func TestGetToken_Azure(t *testing.T) {
	storeType := "none"
	msiEndpoint := "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https%3A%2F%2Fmanagement.azure.com%2F"

	msiResponse := map[string]interface{}{
		"status": 200,
		"response": map[string]interface{}{
			"access_token": "abc.123.abc",
			"expires_in":   3600,
		},
	}

	testCases := []struct {
		name          string
		auth          string
		azureEnv      string
		msiSecret     string
		msiEndpoint   string
		msiResponse   map[string]interface{}
		apiResponse   map[string]interface{}
		expectedError *errors.ApiError
	}{
		{
			"azure auth should fail if environment is invalid",
			"azure",
			"NOT A REAL AZURE ENV",
			"",
			"not-even-a-valid-url",
			msiResponse,
			vtr,
			errors.NewF("Failed to create azure authorizer"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			st, err := store.GetStore(storeType)
			if err != nil {
				t.Fatalf("error getting store got :: %v", err)
			}
			defer func() {
				_ = os.Remove("MSI_SECRET")
				_ = os.Remove("MSI_ENDPOINT")
				_ = os.Remove("AZURE_ENVIRONMENT")
			}()
			_ = os.Setenv("MSI_SECRET", tc.msiSecret)
			_ = os.Setenv("MSI_ENDPOINT", tc.msiEndpoint)
			_ = os.Setenv("AZURE_ENVIRONMENT", tc.azureEnv)
			authDef := auth.NewAuthenticator(st, requests.NewHttpClient())
			viper.Set("auth.username", "")
			viper.Set("auth.password", "")
			viper.Set("auth.type", tc.auth)
			viper.Set("refreshtoken", "")
			viper.Set("domain", domain)
			viper.Set("tenant", tn)

			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			registerResponse(tokenEndpoint, tc.apiResponse["status"].(int), http.MethodPost, tc.apiResponse["response"].(map[string]interface{}))
			registerResponse(msiEndpoint, tc.msiResponse["status"].(int), http.MethodGet, tc.msiResponse["response"].(map[string]interface{}))

			rsp, err := authDef.GetTokenCacheOverride(false)

			if tc.expectedError != nil && err == nil || err != nil && tc.expectedError == nil {
				t.Fatalf("Failed got  expected :: %v, got :: %v", tc.expectedError, err)
			}

			if tc.expectedError != nil {
				assert.Contains(t, err.Error(), tc.expectedError.Error())
			}

			if err == nil {
				expectedResponse := tc.apiResponse["response"].(map[string]interface{})
				assert.Equal(t, expectedResponse["accessToken"], rsp.Token)
			}
		})
	}
}

func TestGetToken_GCP(t *testing.T) {
	storeType := "none"
	testCases := []struct {
		name          string
		auth          string
		gcpToken      string
		apiResponse   map[string]interface{}
		expectedError *errors.ApiError
	}{
		{
			"gcp auth should succeed",
			"gcp",
			"i-am-a-gcp-token",
			vtr,
			nil,
		},
		{
			"gcp auth should error on failure",
			"gcp",
			"i-am-expired",
			btr,
			errors.NewF("Failed to authenticate with auth type 'gcp'"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			st, err := store.GetStore(storeType)
			if err != nil {
				t.Fatalf("error getting store got :: %v", err)
			}
			authDef := auth.NewAuthenticator(st, requests.NewHttpClient())
			viper.Set("auth.username", "")
			viper.Set("auth.password", "")
			viper.Set("auth.type", tc.auth)
			viper.Set("auth.gcp.token", tc.gcpToken)
			viper.Set("refreshtoken", "")
			viper.Set("domain", domain)
			viper.Set("tenant", tn)

			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			registerResponse(tokenEndpoint, tc.apiResponse["status"].(int), http.MethodPost, tc.apiResponse["response"].(map[string]interface{}))

			rsp, err := authDef.GetTokenCacheOverride(false)

			if tc.expectedError != nil && err == nil || err != nil && tc.expectedError == nil {
				t.Fatalf("Failed got  expected :: %v, got :: %v", tc.expectedError, err)
			}

			if tc.expectedError != nil {
				assert.Contains(t, err.Error(), tc.expectedError.Error())
			}

			if err == nil {
				expectedResponse := tc.apiResponse["response"].(map[string]interface{})
				assert.Equal(t, expectedResponse["accessToken"], rsp.Token)
			}
		})
	}
}

func TestToken_GcpSignJwt(t *testing.T) {
	var credsRaw []byte
	var err error
	creds := test_helpers.GetGcpCreds()
	if creds == nil {
		t.Skipf("gcp creds missing, skipping test")
	}
	if credsRaw, err = json.Marshal(creds); err != nil {
		log.Fatal("unable to marshal test gcp creds")
	}
	if err := ioutil.WriteFile("app_creds.json", credsRaw, 0644); err != nil {
		log.Fatal("unable to write test gcp credential file")
	}
	defer func() {
		os.Remove("app_creds.json")
	}()

	testCases := []struct {
		name          string
		auth          string
		gcpCredFile   string
		expectedError *errors.ApiError
	}{
		{
			"gcp auth should sign token",
			"gcp",
			"app_creds.json",
			nil,
		},
		{
			"gcp auth should sign token",
			"gcp",
			"app_creds_fake.json",
			errors.NewF("unable to find default gcp credentials for iam authentication"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &auth.GcpClient{}
			var dir string
			if dir, err = os.Getwd(); err != nil {
				log.Fatal("unable to find current working directory")
			}
			if err = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", fmt.Sprintf("%s%s%s", dir, string(os.PathSeparator), tc.gcpCredFile)); err != nil {
				log.Fatal("error setting gcp creds environment variable")
			}
			defer func() {
				os.Remove("GOOGLE_APPLICATION_CREDENTIALS")
			}()

			viper.Set("domain", domain)
			viper.Set("tenant", tn)

			// act
			gcpJwt, err := client.GetJwtToken()

			if tc.expectedError != nil && err == nil || err != nil && tc.expectedError == nil {
				t.Fatalf("Failed got  expected :: %v, got :: %v", tc.expectedError, err)
			}

			if tc.expectedError != nil {
				assert.Contains(t, err.Error(), tc.expectedError.Error())
			}

			if err == nil {
				claims := &jwt.RegisteredClaims{}
				parser := jwt.Parser{}
				if _, _, err := parser.ParseUnverified(gcpJwt, claims); err != nil {
					t.Fatalf("failed to parse gcp token: %v", err)
				}
				assert.Equal(t, creds.ClientEmail, claims.Subject)
				assert.Equal(t, "https://accounts.google.com", claims.Issuer)
				assert.Equal(t, fmt.Sprintf("https://%s.%s", tn, domain), claims.Audience)
			}
		})
	}
}

func registerResponse(url string, status int, method string, jsonResponse map[string]interface{}) {
	httpmock.RegisterResponder(method, url,
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(status, jsonResponse)
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)
}
