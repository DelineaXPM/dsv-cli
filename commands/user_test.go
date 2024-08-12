package cmd

import (
	"fmt"
	"testing"

	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/tests/fake"
	"github.com/DelineaXPM/dsv-cli/vaultcli"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetUserCmd(t *testing.T) {
	_, err := GetUserCmd()
	assert.Nil(t, err)
}

func TestGetUserReadCmd(t *testing.T) {
	_, err := GetUserReadCmd()
	assert.Nil(t, err)
}

func TestGetUserSearchCmd(t *testing.T) {
	_, err := GetUserSearchCmd()
	assert.Nil(t, err)
}

func TestGetUserDeleteCmd(t *testing.T) {
	_, err := GetUserDeleteCmd()
	assert.Nil(t, err)
}

func TestGetUserRestoreCmd(t *testing.T) {
	_, err := GetUserRestoreCmd()
	assert.Nil(t, err)
}

func TestGetUserCreateCmd(t *testing.T) {
	_, err := GetUserCreateCmd()
	assert.Nil(t, err)
}

func TestGetUserUpdateCmd(t *testing.T) {
	_, err := GetUserUpdateCmd()
	assert.Nil(t, err)
}

func TestHandleUserReadCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		out         []byte
		expectedErr error
	}{
		{
			name:        "Happy path",
			args:        "user1",
			out:         []byte(`test`),
			expectedErr: nil,
		},
		{
			name:        "api Error",
			args:        "user1",
			out:         []byte(`test`),
			expectedErr: errors.NewS("error"),
		},
		{
			name:        "No user passed",
			args:        "",
			out:         []byte(`test`),
			expectedErr: fmt.Errorf("error: must specify --username"),
		},
	}

	viper.Set(cst.Version, "v1")
	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, e *errors.ApiError) {
				data = bytes
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i any) ([]byte, *errors.ApiError) {
				return tt.out, errors.New(tt.expectedErr)
			}

			vcli, err := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if err != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			err = handleUserReadCmd(vcli, []string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleUserDeleteCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		out         []byte
		expectedErr error
	}{
		{
			"Happy path",
			"user1",
			[]byte(`test`),
			nil,
		},
		{
			"api Error",
			"user1",
			[]byte(`test`),
			errors.NewS("error"),
		},
		{
			"No DataUsername",
			"",
			[]byte(`test`),
			fmt.Errorf("error: must specify --username"),
		},
	}

	viper.Set(cst.Version, "v1")
	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, e *errors.ApiError) {
				data = bytes
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i any) ([]byte, *errors.ApiError) {
				return tt.out, errors.New(tt.expectedErr)
			}

			vcli, err := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if err != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			err = handleUserDeleteCmd(vcli, []string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleUserSearchCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		out         []byte
		expectedErr error
	}{
		{
			"Happy path",
			"user1",
			[]byte(`test`),
			nil,
		},
		{
			"api Error",
			"user1",
			[]byte(`test`),
			errors.NewS("error"),
		},
		{
			"No Search query",
			"",
			[]byte(`test`),
			errors.NewS("error: must specify " + cst.Query),
		},
	}

	viper.Set(cst.Version, "v1")
	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, e *errors.ApiError) {
				data = bytes
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i any) ([]byte, *errors.ApiError) {
				return tt.out, errors.New(tt.expectedErr)
			}

			vcli, err := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if err != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			err = handleUserSearchCmd(vcli, []string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleUserCreateCmd(t *testing.T) {
	testCase := []struct {
		name, userName, displayName, password, provider, externalID string
		args                                                        []string
		out                                                         []byte
		expectedErr                                                 string
	}{
		{
			name:     "Successful local user create",
			args:     []string{"--username", "user1", "--password", "password"},
			userName: "user1",
			password: "password",
			out:      []byte(`test`),
		},
		{
			name:        "Successful local user create with displayname",
			args:        []string{"--username", "user1", "--password", "password"},
			userName:    "user1",
			displayName: "user1 display name",
			password:    "password",
			out:         []byte(`test`),
		},
		{
			name:        "Create fails no username",
			args:        []string{"--password", "password"},
			password:    "password",
			expectedErr: `error: username "" is invalid: must specify username`,
		},
		{
			name:        "Create fails no password",
			args:        []string{"--username", "user"},
			userName:    "user1",
			expectedErr: "error: must specify password for local users",
		},
		{
			name:        "3rd party provider missing",
			args:        []string{"--username", "user", "--external-id", "1234"},
			userName:    "user1",
			externalID:  "1234",
			expectedErr: "error: must specify both provider and external ID for third-party users",
		},
		{
			name:        "3rd party external ID missing",
			args:        []string{"--username", "user", "--provider", "aws-dev"},
			userName:    "user1",
			provider:    "aws-dev",
			expectedErr: "error: must specify both provider and external ID for third-party users",
		},
		{
			name:       "Successful 3rd party user create",
			args:       []string{"--username", "user", "--provider", "aws-dev", "--external-id", "1234"},
			userName:   "user1",
			provider:   "aws-dev",
			externalID: "1234",
		},
	}

	viper.Set(cst.Version, "v1")
	for _, tt := range testCase {
		viper.Set(cst.DataUsername, tt.userName)
		viper.Set(cst.DataDisplayname, tt.displayName)
		viper.Set("password", tt.password)
		viper.Set(cst.DataProvider, tt.provider)
		viper.Set(cst.DataExternalID, tt.externalID)
		t.Run(tt.name, func(t *testing.T) {
			var data []byte

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, e *errors.ApiError) {
				data = bytes
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i any) ([]byte, *errors.ApiError) {
				return tt.out, errors.NewS(tt.expectedErr)
			}

			vcli, err := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if err != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			err = handleUserCreateCmd(vcli, tt.args)
			if tt.expectedErr == "" {
				assert.Equal(t, tt.out, data)
				assert.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), tt.expectedErr)
			}
		})
	}
}

func TestHandleUserUpdateCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        []string
		userName    string
		password    string
		displayName string
		out         []byte
		expectedErr error
	}{
		{
			name:     "Happy path with password only",
			args:     []string{"--username", "user1", "--password", "password"},
			userName: "user1",
			password: "password",
			out:      []byte(`test`),
		},
		{
			name:        "no username",
			args:        []string{"--password", "password"},
			password:    "password",
			out:         []byte(`test`),
			expectedErr: fmt.Errorf("error: must specify --username"),
		},
		{
			name:        "no password and no display name",
			args:        []string{"--username", "user"},
			userName:    "user1",
			out:         []byte(`test`),
			expectedErr: errMustSpecifyPasswordOrDisplayname,
		},
		{
			name:        "empty display name",
			args:        []string{"--username", "user", "--displayname", ""},
			userName:    "user1",
			displayName: "",
			out:         []byte(`test`),
			expectedErr: errWrongDisplayName,
		},
		{
			name:        "short display name",
			args:        []string{"--username", "user", "--displayname", "X"},
			userName:    "user1",
			displayName: "X",
			out:         []byte(`test`),
			expectedErr: errWrongDisplayName,
		},
		{
			name:        "Happy path with display name only",
			args:        []string{"--username", "user1", "--displayname", "display name 2"},
			userName:    "user1",
			password:    "password",
			displayName: "display name 2",
			out:         []byte(`test`),
		},
		{
			name:        "Happy path with password and display name",
			args:        []string{"--username", "user1", "--password", "password", "--displayname", "display name 2"},
			userName:    "user1",
			password:    "password",
			displayName: "display name 2",
			out:         []byte(`test`),
		},
	}

	viper.Set(cst.Version, "v1")
	for _, tt := range testCase {
		viper.Set(cst.DataUsername, tt.userName)
		viper.Set("password", tt.password)
		viper.Set(cst.DataDisplayname, tt.displayName)
		t.Run(tt.name, func(t *testing.T) {
			var data []byte

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, e *errors.ApiError) {
				data = bytes
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i any) ([]byte, *errors.ApiError) {
				return tt.out, errors.New(tt.expectedErr)
			}

			vcli, err := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if err != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			err = handleUserUpdateCmd(vcli, tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}
