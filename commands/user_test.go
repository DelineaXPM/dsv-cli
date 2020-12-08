package cmd

import (
	e "errors"
	"testing"
	cst "thy/constants"
	"thy/errors"
	"thy/fake"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestHandleUserReadCmd(t *testing.T) {

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
			nil},
		{
			"api Error",
			"user1",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error")),
		},
		{
			"No user passed",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify " + cst.DataUsername)),
		},
	}

	_, err := GetUserReadCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tt := range testCase {

		t.Run(tt.name, func(t *testing.T) {

			acmd := &fake.FakeOutClient{}
			var data []byte
			var err *errors.ApiError
			acmd.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			req := &fake.FakeClient{}
			req.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.out, tt.expectedErr
			}

			u := User{req, acmd}
			_ = u.handleUserReadCmd([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})

	}
}

func TestHandleUserDeleteCmd(t *testing.T) {
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
			nil},
		{
			"api Error",
			"user1",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error")),
		},
		{
			"No DataUsername",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify " + cst.DataUsername)),
		},
	}

	_, err := GetUserDeleteCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tt := range testCase {

		t.Run(tt.name, func(t *testing.T) {
			acmd := &fake.FakeOutClient{}
			var data []byte
			var err *errors.ApiError
			acmd.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			req := &fake.FakeClient{}
			req.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.out, tt.expectedErr
			}

			u := User{req, acmd}
			_ = u.handleUserDeleteCmd([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})

	}
}

func TestHandleUserSearchCmd(t *testing.T) {

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
			nil},
		{
			"api Error",
			"user1",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error")),
		},
		{
			"No Search query",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify " + cst.Query)),
		},
	}

	_, err := GetUserSearchCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tt := range testCase {

		t.Run(tt.name, func(t *testing.T) {
			acmd := &fake.FakeOutClient{}
			var data []byte
			var err *errors.ApiError
			acmd.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			req := &fake.FakeClient{}
			req.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.out, tt.expectedErr
			}

			u := User{req, acmd}
			_ = u.handleUserSearchCmd([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})

	}
}

func TestHandleUserCreateCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        []string
		userName    string
		password    string
		provider    string
		externalID  string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name:        "Successful local user create",
			args:        []string{"--username", "user1", "--password", "password"},
			userName:    "user1",
			password:    "password",
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
		},
		{
			name:        "Create fails no username",
			args:        []string{"--password", "password"},
			password:    "password",
			expectedErr: errors.New(e.New("error: must specify " + cst.DataUsername)),
		},
		{
			name:        "Create fails no password",
			args:        []string{"--username", "user"},
			userName:    "user1",
			expectedErr: errors.New(e.New("error: must specify password for local users")),
		},
		{
			name:        "3rd party provider missing",
			args:        []string{"--username", "user", "--external-id", "1234"},
			userName:    "user1",
			externalID:  "1234",
			expectedErr: errors.New(e.New("error: must specify both provider and external ID for third-party users")),
		},
		{
			name:        "3rd party external ID missing",
			args:        []string{"--username", "user", "--provider", "aws-dev"},
			userName:    "user1",
			provider:    "aws-dev",
			expectedErr: errors.New(e.New("error: must specify both provider and external ID for third-party users")),
		},
		{
			name:       "Successful 3rd party user create",
			args:       []string{"--username", "user", "--provider", "aws-dev", "--external-id", "1234"},
			userName:   "user1",
			provider:   "aws-dev",
			externalID: "1234",
		},
	}

	_, err := GetUserCreateCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tt := range testCase {
		viper.Set(cst.DataUsername, tt.userName)
		viper.Set(cst.DataPassword, tt.password)
		viper.Set(cst.DataProvider, tt.provider)
		viper.Set(cst.DataExternalID, tt.externalID)
		t.Run(tt.name, func(t *testing.T) {
			acmd := &fake.FakeOutClient{}
			var data []byte
			var err *errors.ApiError
			acmd.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			req := &fake.FakeClient{}
			req.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.out, tt.expectedErr
			}

			u := User{req, acmd}
			_ = u.handleUserCreateCmd(tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})

	}
}

func TestHandleUserUpdateCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        []string
		userName    string
		password    string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Happy path",
			[]string{"--username", "user1", "--password", "password"},
			"user1",
			"password",
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},
		{
			"no username",
			[]string{"--password", "password"},
			"",
			"password",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify " + cst.DataUsername)),
		},
		{
			"no password",
			[]string{"--username", "user"},
			"user1",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify " + cst.DataPassword)),
		},
	}

	_, err := GetUserUpdateCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tt := range testCase {
		viper.Set(cst.DataUsername, tt.userName)
		viper.Set(cst.DataPassword, tt.password)
		t.Run(tt.name, func(t *testing.T) {
			acmd := &fake.FakeOutClient{}
			var data []byte
			var err *errors.ApiError
			acmd.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			req := &fake.FakeClient{}
			req.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.out, tt.expectedErr
			}

			u := User{req, acmd}
			_ = u.handleUserUpdateCmd(tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestGetUserCmd(t *testing.T) {
	_, err := GetUserCmd()
	assert.Nil(t, err)
	//cmd.Run([]string{"test"})
}
