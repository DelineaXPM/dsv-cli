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

func TestHandleRoleReadCmd(t *testing.T) {

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
			errors.New(e.New("error: must specify " + cst.DataName)),
		},
	}

	_, err := GetRoleReadCmd()
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

			r := Roles{req, acmd}
			_ = r.handleRoleReadCmd([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})

	}
}

func TestHandleRoleSearchCmd(t *testing.T) {

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

	_, err := GetRoleSearchCmd()
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

			r := Roles{req, acmd}
			_ = r.handleRoleSearchCmd([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})

	}
}

func TestHandleRoleDeleteCmd(t *testing.T) {
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
			errors.New(e.New("error: must specify " + cst.DataName)),
		},
	}

	_, err := GetRoleDeleteCmd()
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

			r := Roles{req, acmd}
			_ = r.handleRoleDeleteCmd([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})

	}
}

func TestHandleRoleUpsertCmd(t *testing.T) {
	testCase := []struct {
		name        string
		roleName    string
		args        []string
		apiResponse []byte
		out         []byte
		method      string
		expectedErr *errors.ApiError
	}{
		{
			"Successful create",
			"role1",
			[]string{"--name", "role1"},
			[]byte(`test`),
			[]byte(`test`),
			"create",
			nil,
		},
		{
			"Successful create",
			"role1",
			[]string{"--name", "role1"},
			[]byte(`test`),
			[]byte(`test`),
			"update",
			nil,
		},
		{
			"Create - no name",
			"",
			[]string{"--desc", "updated role"},
			[]byte(`test`),
			[]byte(`test`),
			"create",
			errors.New(e.New("error: must specify " + cst.DataName)),
		},
		{
			"Update - no name",
			"",
			[]string{"--desc", "updated role"},
			[]byte(`test`),
			[]byte(`test`),
			"update",
			errors.New(e.New("error: must specify " + cst.DataName)),
		},
	}

	viper.Set(cst.Version, "v1")
	for _, tt := range testCase {
		viper.Set(cst.DataName, tt.roleName)
		t.Run(tt.name, func(t *testing.T) {
			writer := &fake.FakeOutClient{}
			var data []byte
			var err *errors.ApiError
			writer.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			req := &fake.FakeClient{}
			req.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.out, tt.expectedErr
			}

			r := &Roles{req, writer}
			viper.Set(cst.LastCommandKey, tt.method)

			_ = r.handleRoleUpsertCmd(tt.args)

			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestGetRoleCmd(t *testing.T) {
	_, err := GetRoleCmd()
	assert.Nil(t, err)
	//cmd.Run([]string{"test"})
}
