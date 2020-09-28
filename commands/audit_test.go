package cmd

import (
	e "errors"
	"testing"
	cst "thy/constants"
	"thy/errors"
	"thy/fake"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestHandleSearchAuditCmd(t *testing.T) {
	today := time.Now()
	cases := []struct {
		name        string
		startDate   string
		endDate     string
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Valid request",
			today.Format("2006-01-02"),
			"",
			[]byte("data"),
			nil,
		},
		{
			"No start date",
			"",
			"",
			nil,
			errors.New(e.New("error: must specify " + cst.StartDate)),
		},
		{
			"Start date equals end date",
			today.Format("2006-01-02"),
			today.Format("2006-01-02"),
			[]byte("data"),
			nil,
		},
		{
			"Start date after end date",
			today.Format("2006-01-02"),
			today.AddDate(0, 0, -5).Format("2006-01-02"),
			nil,
			errors.NewS("error: start date cannot be after end date"),
		},
		{
			"End date in the future",
			today.Format("2006-01-02"),
			today.AddDate(0, 0, 5).Format("2006-01-02"),
			nil,
			errors.NewS("error: start date cannot be in the future"),
		},
	}

	_, err := GetAuditSearchCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set(cst.StartDate, tt.startDate)
			viper.Set(cst.EndDate, tt.endDate)
			client := &fake.FakeOutClient{}
			var data []byte
			var err *errors.ApiError
			client.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			req := &fake.FakeClient{}
			req.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.out, tt.expectedErr
			}

			a := audit{req, client}
			_ = a.handleAuditSearch([]string{tt.startDate, tt.endDate})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}
