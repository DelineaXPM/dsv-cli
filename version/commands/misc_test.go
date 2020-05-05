package cmd

import (
	"testing"
	cst "thy/constants"
	"thy/errors"
	"thy/fake"

	"github.com/stretchr/testify/assert"
	"github.com/thycotic-rd/viper"
)

func TestHandleWhoAmICmd(t *testing.T) {

	testCase := []struct {
		name        string
		args        string
		userType    string
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Happy path",
			"user1",
			"auth.username",
			[]byte(`auth.username`),
			nil,
		},
		{
			"Happy path",
			"user1",
			"auth.awsprofile",
			[]byte(`auth.awsprofile`),
			nil,
		},
		{
			"Happy path",
			"user1",
			"auth.client.id",
			[]byte(`auth.client.id`),
			nil,
		},
		{
			"Happy path",
			"user1",
			"auth.type",
			[]byte(`whoami does not yet support displaying an ID for a user authenticated with Azure`),
			nil,
		},
	}

	_, err := GetWhoAmICmd()
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

			if tt.userType == "auth.username" {
				viper.Set(cst.Username, tt.userType)
			} else if tt.userType == "auth.awsprofile" {
				viper.Set(cst.AwsProfile, tt.userType)
			} else if tt.userType == "auth.client.id" {
				viper.Set(cst.AuthClientID, tt.userType)
			} else {
				viper.Set(cst.AuthType, "azure")
			}

			u := Misc{writer}
			_ = u.handleWhoAmICmd([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)

			} else {
				assert.Equal(t, err, tt.expectedErr)
			}

			//clean up
			viper.Set(cst.Username, "")
			viper.Set(cst.AwsProfile, "")
			viper.Set(cst.AuthClientID, "")
			viper.Set(cst.AuthType, "")
		})

	}
}

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
