package cmd

import (
	e "errors"
	"testing"
	cst "thy/constants"
	"thy/errors"
	"thy/fake"

	"github.com/stretchr/testify/assert"
	"github.com/thycotic-rd/viper"
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
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
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
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
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
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})

	}
}

func TestHandleUserPostCmd(t *testing.T) {

	testCase := []struct {
		name        string
		args        []string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Happy path",
			[]string{"user1", "password"},
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},

		{
			"no user ",
			[]string{""},
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify " + cst.DataUsername)),
		},
		{
			"no password",
			[]string{"user"},
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify " + cst.DataPassword)),
		},
	}

	_, err := GetUserCreateCmd()
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
			_ = u.handleUserPostCmd(tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})

	}
}

func TestGetUserCmd(t *testing.T) {
	_, err := GetUserCmd()
	assert.Nil(t, err)
	//cmd.Run([]string{"test"})
}
