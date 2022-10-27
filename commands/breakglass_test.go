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

func TestGetBreakGlassCmd(t *testing.T) {
	_, err := GetBreakGlassCmd()
	assert.Nil(t, err)
}

func TestGetBreakGlassGetStatusCmd(t *testing.T) {
	_, err := GetBreakGlassGetStatusCmd()
	assert.Nil(t, err)
}

func TestGetBreakGlassGenerateCmd(t *testing.T) {
	_, err := GetBreakGlassGenerateCmd()
	assert.Nil(t, err)
}

func TestGetBreakGlassApplyCmd(t *testing.T) {
	_, err := GetBreakGlassApplyCmd()
	assert.Nil(t, err)
}

func TestHandleBreakGlassGetStatusCmd(t *testing.T) {
	testCase := []struct {
		name                string
		apiOut              []byte
		apiErr              *errors.ApiError
		wantOut             []byte
		wantErr             *errors.ApiError
		wantNotZeroExitCode bool
	}{
		{
			name:    "echo API response",
			apiOut:  []byte(`{"status":"Break Glass feature is set"}`),
			wantOut: []byte(`{"status":"Break Glass feature is set"}`),
		},
		{
			name:                "echo API error",
			apiErr:              errors.NewS(`{"error":"message"}`),
			wantErr:             errors.NewS(`{"error":"message"}`),
			wantNotZeroExitCode: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tc.apiOut, tc.apiErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", rerr)
			}

			code := handleBreakGlassGetStatusCmd(vcli, []string{})

			if tc.wantErr == nil {
				assert.Equal(t, tc.wantOut, data)
			} else {
				assert.Equal(t, tc.wantErr, err)
			}

			if tc.wantNotZeroExitCode {
				assert.NotEqual(t, 0, code)
			} else {
				assert.Equal(t, 0, code)
			}
		})
	}
}

func TestHandleBreakGlassGenerateCmd(t *testing.T) {
	testCase := []struct {
		name               string
		fNewAdmins         string
		fMinNumberOfShares string
		apiOut             []byte
		apiErr             *errors.ApiError
		wantOut            []byte
		wantErr            *errors.ApiError
	}{
		{
			name:               "success",
			fNewAdmins:         "bguser1,bguser2,bguser3",
			fMinNumberOfShares: "2",
			apiOut:             []byte(`{"response":"success"}`),
			wantOut:            []byte(`{"response":"success"}`),
		},
		{
			name:               "API error",
			fNewAdmins:         "bguser1,bguser2,bguser3",
			fMinNumberOfShares: "2",
			apiErr:             errors.NewS(`{"error":"message"}`),
			wantErr:            errors.NewS(`{"error":"message"}`),
		},
		{
			name:               "missing --new-admins",
			fMinNumberOfShares: "2",
			wantErr:            errors.NewS("error: must specify new-admins"),
		},
		{
			name:       "missing --min-number-of-shares",
			fNewAdmins: "bguser1,bguser2,bguser3",
			wantErr:    errors.NewS("error: must specify min-number-of-shares"),
		},
		{
			name:               "invalid --min-number-of-shares",
			fNewAdmins:         "bguser1,bguser2,bguser3",
			fMinNumberOfShares: "aaa",
			wantErr:            errors.NewS("error: minimum number of shares must be a valid integer"),
		},
		{
			name:               "negative --min-number-of-shares",
			fNewAdmins:         "bguser1,bguser2,bguser3",
			fMinNumberOfShares: "-1",
			wantErr:            errors.NewS("error: minimum number of shares must be greater than 1"),
		},
		{
			name:               "zero --min-number-of-shares",
			fNewAdmins:         "bguser1,bguser2,bguser3",
			fMinNumberOfShares: "0",
			wantErr:            errors.NewS("error: minimum number of shares must be greater than 1"),
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tc.apiOut, tc.apiErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", rerr)
			}

			viper.Reset()
			viper.Set(cst.NewAdmins, tc.fNewAdmins)
			viper.Set(cst.MinNumberOfShares, tc.fMinNumberOfShares)

			_ = handleBreakGlassGenerateCmd(vcli, []string{})

			if tc.wantErr == nil {
				assert.Equal(t, tc.wantOut, data)
			} else {
				assert.Equal(t, tc.wantErr, err)
			}
		})
	}
}

func TestHandleBreakGlassApplyCmd(t *testing.T) {
	testCase := []struct {
		name    string
		fShares string
		apiOut  []byte
		apiErr  *errors.ApiError
		wantOut []byte
		wantErr *errors.ApiError
	}{
		{
			name:    "success",
			fShares: "aaaa,bbb",
			apiOut:  []byte(`{"response":"success"}`),
			wantOut: []byte(`{"response":"success"}`),
		},
		{
			name:    "API error",
			fShares: "aaaa,bbb",
			apiErr:  errors.NewS(`{"error":"message"}`),
			wantErr: errors.NewS(`{"error":"message"}`),
		},
		{
			name:    "missing --shares",
			wantErr: errors.NewS("error: must specify shares"),
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tc.apiOut, tc.apiErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", rerr)
			}

			viper.Reset()
			viper.Set(cst.Shares, tc.fShares)

			_ = handleBreakGlassApplyCmd(vcli, []string{})

			if tc.wantErr == nil {
				assert.Equal(t, tc.wantOut, data)
			} else {
				assert.Equal(t, tc.wantErr, err)
			}
		})
	}
}
