package cmd

import (
	"testing"

	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/tests/fake"
	"github.com/DelineaXPM/dsv-cli/vaultcli"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetRoleCmd(t *testing.T) {
	_, err := GetRoleCmd()
	assert.Nil(t, err)
}

func TestGetRoleReadCmd(t *testing.T) {
	_, err := GetRoleReadCmd()
	assert.Nil(t, err)
}

func TestGetRoleSearchCmd(t *testing.T) {
	_, err := GetRoleSearchCmd()
	assert.Nil(t, err)
}

func TestGetRoleDeleteCmd(t *testing.T) {
	_, err := GetRoleDeleteCmd()
	assert.Nil(t, err)
}

func TestGetRoleRestoreCmd(t *testing.T) {
	_, err := GetRoleRestoreCmd()
	assert.Nil(t, err)
}

func TestGetRoleUpdateCmd(t *testing.T) {
	_, err := GetRoleUpdateCmd()
	assert.Nil(t, err)
}

func TestGetRoleCreateCmd(t *testing.T) {
	_, err := GetRoleCreateCmd()
	assert.Nil(t, err)
}

func TestHandleRoleReadCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name:        "Happy path",
			args:        "rolename",
			out:         []byte(`test`),
			expectedErr: nil,
		},
		{
			name:        "API error",
			args:        "rolename",
			expectedErr: errors.NewS("error"),
		},
		{
			name:        "Missing role name",
			expectedErr: errors.NewS("error: must specify " + cst.DataName),
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
				return tt.out, tt.expectedErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.Version, "v1")

			_ = handleRoleReadCmd(vcli, []string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleRoleSearchCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name: "Happy path",
			args: "rolename",
			out:  []byte(`test`),
		},
		{
			name:        "api Error",
			args:        "rolename",
			expectedErr: errors.NewS("error"),
		},
		{
			name:        "Missing query",
			expectedErr: errors.NewS("error: must specify " + cst.Query),
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
				return tt.out, tt.expectedErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()

			_ = handleRoleSearchCmd(vcli, []string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleRoleDeleteCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name: "Happy path",
			args: "rolename",
			out:  []byte(`test`),
		},
		{
			name:        "api Error",
			args:        "rolename",
			expectedErr: errors.NewS("error"),
		},
		{
			name:        "Missing role name",
			expectedErr: errors.NewS("error: must specify " + cst.DataName),
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
				return tt.out, tt.expectedErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()

			_ = handleRoleDeleteCmd(vcli, []string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleRoleRestoreCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name: "Happy path",
			args: "rolename",
			out:  []byte(`test`),
		},
		{
			name:        "api Error",
			args:        "rolename",
			expectedErr: errors.NewS("error"),
		},
		{
			name:        "Missing role name",
			expectedErr: errors.NewS("error: must specify " + cst.DataName),
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
				return tt.out, tt.expectedErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()

			_ = handleRoleRestoreCmd(vcli, []string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleRoleUpsertCmd(t *testing.T) {
	testCase := []struct {
		name        string
		roleName    string
		provider    string
		externalID  string
		args        []string
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name:     "Successful create",
			roleName: "role1",
			args:     []string{"--name", "role1"},
			out:      []byte(`test`),
		},
		{
			name:        "Create fails no name",
			args:        []string{"--desc", "new role"},
			expectedErr: errors.NewS("error: must specify " + cst.DataName),
		},
		{
			name:        "Create fails external ID is missing",
			roleName:    "role2",
			provider:    "aws-dev",
			args:        []string{"--name", "role2", "--provider", "aws-dev"},
			expectedErr: errors.NewS("error: must specify both provider and external ID for third-party roles"),
		},
		{
			name:        "Create fails provider is missing",
			roleName:    "role2",
			externalID:  "1234",
			args:        []string{"--name", "role2", "--external-id", "1234"},
			expectedErr: errors.NewS("error: must specify both provider and external ID for third-party roles"),
		},
		{
			name:       "Successful 3rd party role create",
			roleName:   "role1",
			provider:   "aws-dev",
			externalID: "1234",
			args:       []string{"--name", "role2", "--provider", "aws-dev", "--external-id", "1234"},
			out:        []byte(`test`),
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
				return tt.out, tt.expectedErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.DataName, tt.roleName)
			viper.Set(cst.DataProvider, tt.provider)
			viper.Set(cst.DataExternalID, tt.externalID)

			_ = handleRoleCreateCmd(vcli, tt.args)

			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestHandleRoleUpdateCmd(t *testing.T) {
	testCase := []struct {
		name        string
		roleName    string
		args        []string
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name:     "Successful update",
			roleName: "role1",
			args:     []string{"--name", "role1"},
			out:      []byte(`test`),
		},
		{
			name:        "Update fails no name",
			args:        []string{"--desc", "updated role"},
			expectedErr: errors.NewS("error: must specify " + cst.DataName),
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
				return tt.out, tt.expectedErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.DataName, tt.roleName)

			_ = handleRoleUpdateCmd(vcli, tt.args)

			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			}
		})
	}
}
