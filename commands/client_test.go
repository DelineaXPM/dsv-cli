package cmd

import (
	"testing"

	cst "thy/constants"
	"thy/errors"
	"thy/fake"
	"thy/vaultcli"

	"github.com/stretchr/testify/assert"
)

func TestGetClientCmd(t *testing.T) {
	_, err := GetClientCmd()
	assert.Nil(t, err)
}

func TestGetClientReadCmd(t *testing.T) {
	_, err := GetClientReadCmd()
	assert.Nil(t, err)
}

func TestGetClientDeleteCmd(t *testing.T) {
	_, err := GetClientDeleteCmd()
	assert.Nil(t, err)
}

func TestGetClientRestoreCmd(t *testing.T) {
	_, err := GetClientRestoreCmd()
	assert.Nil(t, err)
}

func TestGetClientCreateCmd(t *testing.T) {
	_, err := GetClientCreateCmd()
	assert.Nil(t, err)
}

func TestGetClientSearchCmd(t *testing.T) {
	_, err := GetClientSearchCmd()
	assert.Nil(t, err)
}

func TestHandleClientReadCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Happy path",
			"client1",
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},
		{
			"api Error",
			"client1",
			[]byte(`test`),
			[]byte(`test`),
			errors.NewS("error"),
		},
		{
			"No clientID",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.NewS("error: must specify " + cst.ClientID),
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

			_ = handleClientReadCmd(vcli, []string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})

	}
}

func TestHandleClientDeleteCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Happy path",
			"user1",
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},
		{
			"api Error",
			"user1",
			[]byte(`test`),
			[]byte(`test`),
			errors.NewS("error"),
		},
		{
			"No clientID",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.NewS("error: must specify " + cst.ClientID),
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

			_ = handleClientDeleteCmd(vcli, []string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})

	}
}

func TestHandleClientCreateCmd(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Create client credential",
			[]string{"--role", "gcp-svc-1"},
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

			_ = handleClientCreateCmd(vcli, tt.args)

			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})
	}
}

func TestHandleClientSearchCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Happy path",
			"user1",
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},
		{
			"api Error",
			"user1",
			[]byte(`test`),
			[]byte(`test`),
			errors.NewS("error"),
		},
		{
			"No Search query",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.NewS("error: must specify " + cst.NounRole),
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

			_ = handleClientSearchCmd(vcli, []string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})
	}
}
