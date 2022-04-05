package cmd

import (
	"testing"
	"time"

	cst "thy/constants"
	"thy/errors"
	"thy/fake"
	"thy/vaultcli"

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
		expectedErr *errors.ApiError
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
			expectedErr: errors.NewS("error: must specify " + cst.StartDate),
		},
		{
			name:        "Incorrect start date",
			startDate:   "2006-aaa",
			endDate:     "",
			expectedErr: errors.NewS("error: must correctly specify " + cst.StartDate),
		},
		{
			name:        "Incorrect end date",
			startDate:   "2006-01-02",
			endDate:     "2006-aaa",
			expectedErr: errors.NewS("error: must correctly specify " + cst.EndDate),
		},
		{
			name:        "Start date in the future",
			startDate:   time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
			endDate:     "",
			expectedErr: errors.NewS("error: start date cannot be in the future"),
		},
		{
			name:        "Start date after end date",
			startDate:   "2006-01-03",
			endDate:     "2006-01-02",
			expectedErr: errors.NewS("error: start date cannot be after end date"),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set(cst.StartDate, tt.startDate)
			viper.Set(cst.EndDate, tt.endDate)

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

			_ = handleAuditSearch(vcli, []string{tt.startDate, tt.endDate})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.expectedOut, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}
