package cmd

import (
	"net/http"
	"testing"

	cst "thy/constants"
	"thy/errors"
	"thy/tests/fake"
	"thy/vaultcli"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetAuthProviderCmd(t *testing.T) {
	_, err := GetAuthProviderCmd()
	assert.Nil(t, err)
}

func TestGetAuthProviderReadCmd(t *testing.T) {
	_, err := GetAuthProviderReadCmd()
	assert.Nil(t, err)
}

func TestGetAuthProviderDeleteCmd(t *testing.T) {
	_, err := GetAuthProviderDeleteCmd()
	assert.Nil(t, err)
}

func TestGetAuthProviderRestoreCmd(t *testing.T) {
	_, err := GetAuthProviderRestoreCmd()
	assert.Nil(t, err)
}

func TestGetAuthProviderCreateCmd(t *testing.T) {
	_, err := GetAuthProviderCreateCmd()
	assert.Nil(t, err)
}

func TestGetAuthProviderUpdateCmd(t *testing.T) {
	_, err := GetAuthProviderUpdateCmd()
	assert.Nil(t, err)
}

func TestGetAuthProviderEditCmd(t *testing.T) {
	_, err := GetAuthProviderEditCmd()
	assert.Nil(t, err)
}

func TestGetAuthProviderRollbackCmd(t *testing.T) {
	_, err := GetAuthProviderRollbackCmd()
	assert.Nil(t, err)
}

func TestGetAuthProviderSearchCmd(t *testing.T) {
	_, err := GetAuthProviderSearchCmd()
	assert.Nil(t, err)
}

func TestHandleAuthProviderReadCmd(t *testing.T) {
	testCase := []struct {
		name                string
		fName               string
		fVersion            string
		args                []string
		apiOut              []byte
		apiErr              *errors.ApiError
		wantOut             []byte
		wantErr             *errors.ApiError
		wantNotZeroExitCode bool
	}{
		{
			name:    "read auth provider",
			fName:   "aws-dev",
			apiOut:  []byte("api response"),
			wantOut: []byte("api response"),
		},
		{
			name:     "read auth provider versions",
			fName:    "aws-dev",
			fVersion: "1",
			apiOut:   []byte("api response"),
			wantOut:  []byte("api response"),
		},
		{
			name:    "read auth provider name from args",
			args:    []string{"aws-dev"},
			apiOut:  []byte("api response"),
			wantOut: []byte("api response"),
		},
		{
			name:                "missing auth provider name",
			wantNotZeroExitCode: true,
		},
		{
			name:                "api error",
			fName:               "aws-dev",
			apiErr:              errors.NewS("some api error"),
			wantErr:             errors.NewS("some api error"),
			wantNotZeroExitCode: true,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.apiOut, tt.apiErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.DataName, tt.fName)
			viper.Set(cst.Version, tt.fVersion)

			code := handleAuthProviderReadCmd(vcli, tt.args)

			if tt.wantErr == nil {
				assert.Equal(t, tt.wantOut, data)
			} else {
				assert.Equal(t, tt.wantErr, err)
			}

			if tt.wantNotZeroExitCode {
				assert.NotEqual(t, 0, code)
			} else {
				assert.Equal(t, 0, code)
			}
		})
	}
}

func TestHandleAuthProviderCreate(t *testing.T) {
	testCase := []struct {
		name    string
		fName   string
		args    []string
		apiOut  []byte
		apiErr  *errors.ApiError
		wantOut []byte
		wantErr *errors.ApiError
	}{
		{
			name:    "success-create",
			fName:   "aws-dev",
			args:    []string{"--name", "aws-dev"},
			apiOut:  []byte("api response"),
			wantOut: []byte("api response"),
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.apiOut, tt.apiErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.DataName, tt.fName)

			_ = handleAuthProviderCreate(vcli, tt.args)

			if tt.wantErr == nil {
				assert.Equal(t, tt.wantOut, data)
			} else {
				assert.Equal(t, tt.wantErr, err)
			}
		})
	}
}

func TestHandleAuthProviderUpdate(t *testing.T) {
	testCase := []struct {
		name    string
		fName   string
		args    []string
		apiOut  []byte
		apiErr  *errors.ApiError
		wantOut []byte
		wantErr *errors.ApiError
	}{
		{
			name:    "success-update",
			fName:   "aws-dev",
			args:    []string{"--name", "aws-dev"},
			apiOut:  []byte("api response"),
			wantOut: []byte("api response"),
		},
		{
			name:    "fail-validation-error",
			fName:   "aws-dev",
			args:    []string{"--name", "aws-dev"},
			apiErr:  errors.NewS("some api error"),
			wantErr: errors.NewS("some api error"),
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.apiOut, tt.apiErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.DataName, tt.fName)

			_ = handleAuthProviderUpdate(vcli, tt.args)

			if tt.wantErr == nil {
				assert.Equal(t, tt.wantOut, data)
			} else {
				assert.Equal(t, tt.wantErr, err)
			}
		})
	}
}

func TestHandleAuthProviderDeleteCmd(t *testing.T) {
	testCase := []struct {
		name    string
		args    []string
		apiOut  []byte
		apiErr  *errors.ApiError
		wantOut []byte
		wantErr *errors.ApiError
	}{
		{
			name:    "success",
			args:    []string{"azure-dev"},
			apiOut:  []byte("api response"),
			wantOut: []byte("api response"),
		},
		{
			name:    "validation error",
			args:    []string{"missing"},
			apiErr:  errors.NewS("some api error"),
			wantErr: errors.NewS("some api error"),
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.apiOut, tt.apiErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()

			_ = handleAuthProviderDeleteCmd(vcli, tt.args)
			if tt.wantErr == nil {
				assert.Equal(t, tt.wantOut, data)
			} else {
				assert.Equal(t, tt.wantErr, err)
			}
		})
	}
}

func TestHandleAuthProviderRestoreCmd(t *testing.T) {
	testCase := []struct {
		name    string
		fName   string // flag: --name
		args    []string
		apiOut  []byte
		apiErr  *errors.ApiError
		wantOut []byte
		wantErr *errors.ApiError
		wantURL string
	}{
		{
			name:    "Provider name from args",
			args:    []string{"azure-dev"},
			apiOut:  []byte("api response"),
			wantOut: []byte("api response"),
			wantURL: "https://tenat.secretsvaultcloud.com/v1/config/auth/azure-dev/restore",
		},
		{
			name:    "Provider name from flag",
			fName:   "azure-dev",
			apiOut:  []byte("api response"),
			wantOut: []byte("api response"),
			wantURL: "https://tenat.secretsvaultcloud.com/v1/config/auth/azure-dev/restore",
		},
		{
			name:    "API error",
			args:    []string{"azure-dev"},
			apiErr:  errors.NewS("some api error"),
			wantErr: errors.NewS("some api error"),
			wantURL: "https://tenat.secretsvaultcloud.com/v1/config/auth/azure-dev/restore",
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError
			var address string

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				address = s2
				return tt.apiOut, tt.apiErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.Tenant, "tenat")
			viper.Set(cst.DataName, tt.fName)

			_ = handleAuthProviderRestoreCmd(vcli, tt.args)
			if tt.wantErr == nil {
				assert.Equal(t, tt.wantOut, data)
			} else {
				assert.Equal(t, tt.wantErr, err)
			}

			if tt.wantURL != "" {
				assert.Equal(t, tt.wantURL, address)
			}
		})
	}
}

func TestHandleAuthProviderSearchCmd(t *testing.T) {
	testCase := []struct {
		name    string
		args    []string
		apiOut  []byte
		apiErr  *errors.ApiError
		wantOut []byte
		wantErr *errors.ApiError
	}{
		{
			name:    "success",
			args:    []string{"-q", "azure"},
			apiOut:  []byte("api response"),
			wantOut: []byte("api response"),
		},
		{
			name:    "error",
			args:    []string{"-q", "what happens when a server error occurs?"},
			apiErr:  errors.NewS("some api error"),
			wantErr: errors.NewS("some api error"),
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.apiOut, tt.apiErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()

			_ = handleAuthProviderSearchCmd(vcli, tt.args)
			if tt.wantErr == nil {
				assert.Equal(t, tt.wantOut, data)
			} else {
				assert.Equal(t, tt.wantErr, err)
			}
		})
	}
}

func TestHandleAuthProviderRollbackCmd(t *testing.T) {
	testCase := []struct {
		name      string
		args      []string
		apiGetOut []byte
		apiGetErr *errors.ApiError
		apiPutOut []byte
		apiPutErr *errors.ApiError
		wantOut   []byte
		wantErr   *errors.ApiError
	}{
		{
			name:      "success (no version passed in)",
			args:      []string{"azure-dev"},
			apiGetOut: []byte(`{"version": "3"}`),
			apiPutOut: []byte("api response"),
			wantOut:   []byte("api response"),
		},
		{
			name:      "success (no version passed in)",
			args:      []string{"azure-dev"},
			apiGetOut: []byte(`{"version": "3"}`),
			apiPutOut: []byte("api response"),
			wantOut:   []byte("api response"),
		},
		{
			name:      "error during read",
			args:      []string{"azure-dev"},
			apiGetErr: errors.NewS("some api error"),
			wantErr:   errors.NewS("some api error"),
		},
		{
			name:      "error during update",
			args:      []string{"azure-dev"},
			apiGetOut: []byte(`{"version": "3"}`),
			apiPutErr: errors.NewS("some api error"),
			wantErr:   errors.NewS("some api error"),
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				switch s {
				case http.MethodGet:
					return tt.apiGetOut, tt.apiGetErr
				case http.MethodPut:
					return tt.apiPutOut, tt.apiPutErr
				default:
					return nil, nil
				}
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()

			_ = handleAuthProviderRollbackCmd(vcli, tt.args)
			if tt.wantErr == nil {
				assert.Equal(t, tt.wantOut, data)
			} else {
				assert.Equal(t, tt.wantErr, err)
			}
		})
	}
}
