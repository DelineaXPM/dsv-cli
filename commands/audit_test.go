package cmd

import (
	"fmt"
	"testing"
	"time"

	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/tests/fake"
	"github.com/DelineaXPM/dsv-cli/vaultcli"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetAuditSearchCmd(t *testing.T) {
	_, err := GetAuditSearchCmd()
	assert.Nil(t, err)
}

func TestHandleAuditSearch(t *testing.T) {
	cases := []struct {
		name        string
		startDate   string
		endDate     string
		apiOut      []byte
		apiErr      *errors.ApiError
		expectedOut []byte
		expectedErr error
	}{
		{
			name:        "Only start date defined",
			startDate:   "2006-01-02",
			endDate:     "",
			apiOut:      []byte("data"),
			expectedOut: []byte("data"),
		},
		{
			name:        "Both start date and end date are defined",
			startDate:   "2006-01-02",
			endDate:     "2006-01-03",
			apiOut:      []byte("data"),
			expectedOut: []byte("data"),
		},
		{
			name:        "Missing start date",
			startDate:   "",
			endDate:     "",
			expectedErr: fmt.Errorf("error: must specify --startdate"),
		},
		{
			name:        "Incorrect start date",
			startDate:   "2006-aaa",
			endDate:     "",
			expectedErr: fmt.Errorf("error: must correctly specify --startdate"),
		},
		{
			name:        "Incorrect end date",
			startDate:   "2006-01-02",
			endDate:     "2006-aaa",
			expectedErr: fmt.Errorf("error: must correctly specify --enddate"),
		},
		{
			name:        "Start date in the future",
			startDate:   time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
			endDate:     "",
			expectedErr: fmt.Errorf("error: start date cannot be in the future"),
		},
		{
			name:        "Start date after end date",
			startDate:   "2006-01-03",
			endDate:     "2006-01-02",
			expectedErr: fmt.Errorf("error: start date cannot be after end date"),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set(cst.StartDate, tt.startDate)
			viper.Set(cst.EndDate, tt.endDate)

			var data []byte

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) ([]byte, *errors.ApiError) {
				return tt.apiOut, tt.apiErr
			}

			vcli, err := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if err != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			err = handleAuditSearch(vcli, []string{tt.startDate, tt.endDate})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.expectedOut, data)
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}
