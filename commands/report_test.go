package cmd

import (
	e "errors"
	"testing"

	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/tests/fake"
	"github.com/DelineaXPM/dsv-cli/vaultcli"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetReportCmd(t *testing.T) {
	_, err := GetReportCmd()
	assert.Nil(t, err)
}

func TestGetSecretReportCmd(t *testing.T) {
	_, err := GetSecretReportCmd()
	assert.Nil(t, err)
}

func TestGetGroupReportCmd(t *testing.T) {
	_, err := GetGroupReportCmd()
	assert.Nil(t, err)
}

func TestHandleSecretReport(t *testing.T) {
	testCase := []struct {
		name        string
		user        string
		group       string
		role        string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"User",
			"user1",
			"",
			"",
			[]byte(`user data`),
			[]byte(`user data`),
			nil,
		},
		{
			"group",
			"",
			"group1",
			"",
			[]byte(`group data`),
			[]byte(`group data`),
			nil,
		},
		{
			"role",
			"",
			"",
			"role1",
			[]byte(`role data`),
			[]byte(`role data`),
			nil,
		},
		{
			"Sign in User",
			"",
			"",
			"",
			[]byte(`user data`),
			[]byte(`user data`),
			nil,
		},
		{
			"api Error",
			"user1",
			"",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error")),
		},
	}

	viper.Set(cst.Version, "v1")
	for _, tt := range testCase {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			acmd := &fake.FakeOutClient{}
			var data []byte
			var err *errors.ApiError
			acmd.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			req := &fake.FakeGraphClient{}
			req.DoRequestStub = func(string, interface{}, map[string]interface{}) ([]byte, *errors.ApiError) {
				return tt.out, tt.expectedErr
			}
			viper.Set(cst.NounUser, tt.user)
			viper.Set(cst.NounGroup, tt.group)
			viper.Set(cst.NounRole, tt.role)

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithGraphQLClient(req),
				vaultcli.WithOutClient(acmd),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			_ = handleSecretReport(vcli, []string{})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleGroupReport(t *testing.T) {
	testCase := []struct {
		name        string
		user        string
		role        string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"User",
			"user1",
			"",
			[]byte(`user data`),
			[]byte(`user data`),
			nil,
		},
		{
			"group",
			"",
			"",
			[]byte(`group data`),
			[]byte(`group data`),
			nil,
		},
		{
			"role",
			"",
			"role1",
			[]byte(`role data`),
			[]byte(`role data`),
			nil,
		},
		{
			"Sign in User",
			"",
			"",
			[]byte(`user data`),
			[]byte(`user data`),
			nil,
		},
		{
			"api Error",
			"user1",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error")),
		},
	}

	viper.Set(cst.Version, "v1")
	for _, tt := range testCase {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			acmd := &fake.FakeOutClient{}
			var data []byte
			var err *errors.ApiError
			acmd.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			req := &fake.FakeGraphClient{}
			req.DoRequestStub = func(string, interface{}, map[string]interface{}) ([]byte, *errors.ApiError) {
				return tt.out, tt.expectedErr
			}
			viper.Set(cst.NounUser, tt.user)
			viper.Set(cst.NounRole, tt.role)

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithGraphQLClient(req),
				vaultcli.WithOutClient(acmd),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			_ = handleSecretReport(vcli, []string{})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}
