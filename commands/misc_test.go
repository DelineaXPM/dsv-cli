package cmd

import (
	"testing"
	cst "thy/constants"
	"thy/errors"
	"thy/fake"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestHandleEvaluateFlag(t *testing.T) {

	testCase := []struct {
		name        string
		args        string
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Happy path",
			"arg1",
			[]byte(`test`),
			nil,
		},
		{
			"Happy path",
			"--arg1",
			[]byte(`test`),
			nil,
		},
	}

	_, err := GetEvaluateFlagCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tt := range testCase {

		t.Run(tt.name, func(t *testing.T) {
			writer := &fake.FakeOutClient{}
			var data []byte
			var err *errors.ApiError
			writer.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			viper.Set("arg1", "test")

			u := Misc{writer}
			_ = u.handleEvaluateFlag([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}

			viper.Set("arg1", "")
		})

	}
}
