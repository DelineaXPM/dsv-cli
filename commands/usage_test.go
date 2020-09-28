package cmd

import (
	e "errors"
	"fmt"
	"testing"
	cst "thy/constants"
	"thy/errors"
	"thy/fake"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestHandleGetUsageCmd(t *testing.T) {
	today := fmt.Sprintf("--%s=%s", cst.StartDate, time.Now().Format("2006-01-02"))
	cases := []struct {
		name        string
		startDate   string
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Valid request",
			today,
			[]byte("requestsUsed"),
			nil,
		},
		{

			"No start date",
			"",
			nil,
			errors.New(e.New("error: must specify " + cst.StartDate)),
		},
	}

	_, err := GetUsageCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set(cst.StartDate, tt.startDate)
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

			u := usage{req, client}
			_ = u.handleGetUsageCmd([]string{tt.startDate})
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})
	}
}
