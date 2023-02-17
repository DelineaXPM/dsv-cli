package auth_test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/DelineaXPM/dsv-cli/auth"
	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/internal/store"
	"github.com/DelineaXPM/dsv-cli/requests"
	"github.com/DelineaXPM/dsv-cli/tests/fake"
	"github.com/DelineaXPM/dsv-cli/utils/test_helpers"

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

func getAuthenticator(t *testing.T) auth.Authenticator {
	t.Helper()
	st, err := store.GetStore("none")
	if err != nil {
		t.Fatalf("error getting store got :: %v", err)
	}

	return auth.NewAuthenticator(st, requests.NewHttpClient())
}

func TestNewAuthenticatorDefault(t *testing.T) {
	viper.Reset()
	viper.Set(cst.StoreType, "none")

	a := auth.NewAuthenticatorDefault()
	if a == nil {
		t.Fatal("Unexpected returned value: <nil>")
	}
}

func TestWipeCachedTokens(t *testing.T) {
	st := &fake.FakeStore{}

	st.ListStub = func(s string) ([]string, error) {
		return []string{"token-a-profilename", "token-b-differentprofile", "b"}, nil
	}

	deleteCalledFor := []string{}
	st.DeleteStub = func(s string) error {
		deleteCalledFor = append(deleteCalledFor, s)
		return nil
	}

	a := auth.NewAuthenticator(st, nil)

	viper.Reset()
	viper.Set(cst.Profile, "profilename")

	err := a.WipeCachedTokens()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(deleteCalledFor) != 1 {
		t.Fatalf("Unexpected length: %d", len(deleteCalledFor))
	}
	if deleteCalledFor[0] != "token-a-profilename" {
		t.Fatalf("Unexpected deleted token: %s", deleteCalledFor[0])
	}
}

func TestGetToken_UnknownMethod(t *testing.T) {
	a := auth.NewAuthenticator(nil, nil)

	viper.Reset()
	viper.Set(cst.AuthType, "unknown-auth")

	_, err := a.GetToken()
	if err == nil {
		t.Fatal("Expected an error, but got <nil>")
	}
}

func TestGetToken_CachedToken(t *testing.T) {
	st := &fake.FakeStore{}

	var usedCacheKey string
	st.GetStub = func(s string, out any) error {
		usedCacheKey = s

		tr := out.(*auth.TokenResponse)
		*tr = auth.TokenResponse{
			Token:        "aaa-token-bbb",
			RefreshToken: "aaa-refreshtoken-bbb",
			Granted:      time.Now(),
			ExpiresIn:    3600,
		}

		return nil
	}

	a := auth.NewAuthenticator(st, nil)

	viper.Reset()
	viper.Set(cst.Profile, "profilename")
	viper.Set(cst.Tenant, "tenantname")

	tr, err := a.GetToken()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if tr == nil {
		t.Fatal("Unexpected returned value: <nil>")
	}
	if tr.Token != "aaa-token-bbb" {
		t.Fatalf("Unexpected token: %s", tr.Token)
	}
	if usedCacheKey != "token-password-tenantname-profilename" {
		t.Fatalf("Unexpected token cache key: %s", usedCacheKey)
	}
}

func TestGetToken_CachedRefreshToken(t *testing.T) {
	st := &fake.FakeStore{}

	var usedCacheKey string
	st.GetStub = func(s string, out any) error {
		usedCacheKey = s

		tr := out.(*auth.TokenResponse)
		*tr = auth.TokenResponse{
			Token:        "aaa-token-bbb",
			RefreshToken: "aaa-refreshtoken-bbb",
			Granted:      time.Now().AddDate(0, 0, -1),
			ExpiresIn:    3600,
		}

		return nil
	}

	httpClient := &fake.FakeClient{}

	var usedHTTPMethod string
	httpClient.DoRequestOutStub = func(s1, s2 string, i1, i2 interface{}) *errors.ApiError {
		usedHTTPMethod = s1

		tr := i2.(*auth.TokenResponse)
		*tr = auth.TokenResponse{
			Token:        "aaa-new-token-bbb",
			RefreshToken: "aaa-new-refreshtoken-bbb",
			Granted:      time.Now().AddDate(0, 0, -1),
			ExpiresIn:    3600,
		}

		return nil
	}

	a := auth.NewAuthenticator(st, httpClient)

	viper.Reset()
	viper.Set(cst.Profile, "profilename")
	viper.Set(cst.Tenant, "tenantname")

	tr, err := a.GetToken()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if tr == nil {
		t.Fatal("Unexpected returned value: <nil>")
	}
	if tr.Token != "aaa-new-token-bbb" {
		t.Fatalf("Unexpected token: %s", tr.Token)
	}
	if usedCacheKey != "token-password-tenantname-profilename" {
		t.Fatalf("Unexpected token cache key: %s", usedCacheKey)
	}
	if usedHTTPMethod != "POST" {
		t.Fatalf("Unexpected HTTP method used: %s", usedHTTPMethod)
	}
}

// TODO:  need to refactor the code and rewrite
func TestGetToken(t *testing.T) {
	testCases := []struct {
		auth          string
		storeType     string
		want          *auth.TokenResponse
		expectedError error
	}{
		{"password", "none", nil, fmt.Errorf("error")},
		// {"azure", "none", nil, errors.New("error")},
		//{"gcp", "none", nil, errors.New("error")},
		{"aws", "none", nil, fmt.Errorf("error")},
		{"refresh", "none", nil, fmt.Errorf("error")},
		{"clientcred", "none", nil, fmt.Errorf("error")},
	}

	for _, tt := range testCases {
		authDef := getAuthenticator(t)

		viper.Reset()
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authDef := getAuthenticator(t)

			viper.Reset()
			viper.Set("auth.type", tc.auth)
			viper.Set("auth.username", "admin")
			viper.Set("auth.password", tc.password)
			viper.Set("auth.securePassword", tc.securePassword)
			viper.Set("domain", domain)
			viper.Set("tenant", tn)

			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			registerResponse(tokenEndpoint, tc.apiResponse["status"].(int), http.MethodPost, tc.apiResponse["response"].(map[string]interface{}))

			rsp, err := authDef.GetToken()

			if tc.expectedError != nil && err == nil || err != nil && tc.expectedError == nil {
				t.Fatalf("Failed. Expected: %v, got: %v", tc.expectedError, err)
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
	path, err := os.MkdirTemp("", "dsv-testing-*")
	if err != nil {
		t.Fatalf("os.MkdirTemp: %v", err)
	}
	defer os.RemoveAll(path)

	t.Log("Temp dir path:", path)
	viper.Reset()
	viper.Set("store.path", path)
	viper.Set("domain", domain)
	viper.Set("tenant", "tenantname")
	viper.Set("auth.type", "password")
	viper.Set("auth.username", "testuser")

	filename := auth.GetEncryptionKeyFilename("tenantname", "testuser")
	securePass, key, err := auth.StorePassword(filename, "testpassword")
	if err != nil {
		t.Fatalf("auth.StorePassword: %v", err)
	}

	os.WriteFile(filepath.Join(path, filename), []byte(key), 0o644)
	t.Log("Encryption file:", filepath.Join(path, filename))

	viper.Set("auth.securePassword", securePass)

	st := &fake.FakeStore{}
	st.GetStub = func(s string, i any) error {
		t.Logf("fakeStore.Get(%s)", s)
		return nil
	}

	httpClient := &fake.FakeClient{}

	var usedHTTPMethod string
	httpClient.DoRequestOutStub = func(s1, s2 string, i1, i2 interface{}) *errors.ApiError {
		t.Logf("fakeHTTPClient.DoRequestOut(%s, %s, ...)", s1, s2)
		usedHTTPMethod = s1

		tr := i2.(*auth.TokenResponse)
		*tr = auth.TokenResponse{
			Token:        "aaa-new-token-bbb",
			RefreshToken: "aaa-new-refreshtoken-bbb",
			Granted:      time.Now().AddDate(0, 0, -1),
			ExpiresIn:    3600,
		}

		return nil
	}

	authDef := auth.NewAuthenticator(st, httpClient)

	tr, authErr := authDef.GetToken()
	if authErr != nil {
		t.Fatalf("Unexpected error: %v", authErr)
	}
	if tr == nil {
		t.Fatal("Unexpected returned value: <nil>")
	}
	if tr.Token != "aaa-new-token-bbb" {
		t.Fatalf("Unexpected token: %s", tr.Token)
	}
	if usedHTTPMethod != "POST" {
		t.Fatalf("Unexpected HTTP method used: %s", usedHTTPMethod)
	}
}

func TestGetToken_Azure(t *testing.T) {
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
			errors.NewF("Failed to build token request for azure based auth:"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authDef := getAuthenticator(t)

			defer func() {
				_ = os.Remove("MSI_SECRET")
				_ = os.Remove("MSI_ENDPOINT")
				_ = os.Remove("AZURE_ENVIRONMENT")
			}()
			_ = os.Setenv("MSI_SECRET", tc.msiSecret)
			_ = os.Setenv("MSI_ENDPOINT", tc.msiEndpoint)
			_ = os.Setenv("AZURE_ENVIRONMENT", tc.azureEnv)

			viper.Reset()
			viper.Set("auth.type", tc.auth)
			viper.Set("domain", domain)
			viper.Set("tenant", tn)

			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			registerResponse(tokenEndpoint, tc.apiResponse["status"].(int), http.MethodPost, tc.apiResponse["response"].(map[string]interface{}))
			registerResponse(msiEndpoint, tc.msiResponse["status"].(int), http.MethodGet, tc.msiResponse["response"].(map[string]interface{}))

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
			}
		})
	}
}

func TestGetToken_GCP(t *testing.T) {
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
			authDef := getAuthenticator(t)

			viper.Reset()
			viper.Set("auth.type", tc.auth)
			viper.Set("auth.gcp.token", tc.gcpToken)
			viper.Set("domain", domain)
			viper.Set("tenant", tn)

			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			registerResponse(tokenEndpoint, tc.apiResponse["status"].(int), http.MethodPost, tc.apiResponse["response"].(map[string]interface{}))

			rsp, err := authDef.GetToken()

			if tc.expectedError != nil && err == nil || err != nil && tc.expectedError == nil {
				t.Fatalf("Expected: %v, got: %v", tc.expectedError, err)
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
	if err := os.WriteFile("app_creds.json", credsRaw, 0o644); err != nil {
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

			viper.Reset()
			viper.Set("domain", domain)
			viper.Set("tenant", tn)

			// act
			gcpJwt, err := client.GetJwtToken("")

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
