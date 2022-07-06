package cmd

import (
	"testing"

	cst "thy/constants"
	"thy/errors"
	"thy/tests/fake"
	"thy/vaultcli"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetEngineCmd(t *testing.T) {
	_, err := GetEngineCmd()
	assert.Nil(t, err)
}
func TestGetEngineReadCmd(t *testing.T) {
	_, err := GetEngineReadCmd()
	assert.Nil(t, err)
}
func TestGetEngineListCmd(t *testing.T) {
	_, err := GetEngineListCmd()
	assert.Nil(t, err)
}
func TestGetEngineDeleteCmd(t *testing.T) {
	_, err := GetEngineDeleteCmd()
	assert.Nil(t, err)
}
func TestGetEngineCreateCmd(t *testing.T) {
	_, err := GetEngineCreateCmd()
	assert.Nil(t, err)
}

func TestGetEnginePingCmd(t *testing.T) {
	_, err := GetEnginePingCmd()
	assert.Nil(t, err)
}

func TestHandleEngineReadCmd(t *testing.T) {
	testCases := []struct {
		name        string
		engineName  string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Success",
			"engine1",
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},
		{
			"No engine name passed",
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
			viper.Set(cst.DataName, tt.engineName)

			_ = handleEngineReadCmd(vcli, nil)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleEngineCreateCmd(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		engineName  string
		poolName    string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Success",
			[]string{"--name", "engine1", "--pool-name", "pool1"},
			"engine1",
			"pool1",
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},
		{
			"No engine name passed",
			[]string{"--pool-name", "pool1"},
			"",
			"pool1",
			[]byte(`test`),
			[]byte(`test`),
			errors.NewS("error: must specify engine name and pool name"),
		},
		{
			"No pool name passed",
			[]string{"--name", "engine1"},
			"engine1",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.NewS("error: must specify engine name and pool name"),
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
			viper.Set(cst.DataName, tt.engineName)
			viper.Set(cst.DataPoolName, tt.poolName)

			_ = handleEngineCreateCmd(vcli, tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleEngineDeleteCmd(t *testing.T) {
	testCases := []struct {
		name        string
		engineName  string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Success",
			"engine1",
			[]byte(`test`),
			[]byte(`test`),
			nil},
		{
			"No engine name passed",
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
			viper.Set(cst.DataName, tt.engineName)

			_ = handleEngineDeleteCmd(vcli, nil)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}
