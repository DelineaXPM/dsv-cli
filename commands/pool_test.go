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

func TestGetPoolCmd(t *testing.T) {
	_, err := GetPoolCmd()
	assert.Nil(t, err)
}

func TestGetPoolCreateCmd(t *testing.T) {
	_, err := GetPoolCreateCmd()
	assert.Nil(t, err)
}

func TestGetPoolReadCmd(t *testing.T) {
	_, err := GetPoolReadCmd()
	assert.Nil(t, err)
}

func TestGetPoolListCmd(t *testing.T) {
	_, err := GetPoolListCmd()
	assert.Nil(t, err)
}

func TestGetPoolDeleteCmd(t *testing.T) {
	_, err := GetPoolDeleteCmd()
	assert.Nil(t, err)
}

func TestHandlePoolReadCmd(t *testing.T) {
	testCases := []struct {
		name        string
		poolName    string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Success",
			"pool1",
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},
		{
			"No pool name passed",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.NewS("error: must specify " + cst.DataName),
		},
	}

	for _, tt := range testCases {
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
			viper.Set(cst.DataName, tt.poolName)

			_ = handlePoolRead(vcli, nil)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandlePoolCreateCmd(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		poolName    string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Success",
			[]string{"--name", "pool1"},
			"pool1",
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},
	}

	for _, tt := range testCases {
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
			viper.Set(cst.DataName, tt.poolName)

			_ = handlePoolCreate(vcli, tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandlePoolDeleteCmd(t *testing.T) {
	testCases := []struct {
		name        string
		poolName    string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Success",
			"pool1",
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},
		{
			"No pool name passed",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.NewS("error: must specify " + cst.DataName),
		},
	}

	for _, tt := range testCases {
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
			viper.Set(cst.DataName, tt.poolName)

			_ = handlePoolDelete(vcli, nil)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}
