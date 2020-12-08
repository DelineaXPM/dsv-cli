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
		provider    string
		externalID  string
		args        []string
		apiResponse []byte
		out         []byte
		method      string
		expectedErr *errors.ApiError
	}{
		{
			name:        "Successful create",
			roleName:    "role1",
			args:        []string{"--name", "role1"},
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
			method:      "create",
		},
		{
			name:        "Successful update",
			roleName:    "role1",
			args:        []string{"--name", "role1"},
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
			method:      "update",
		},
		{
			name:        "Create fails no name",
			args:        []string{"--desc", "new role"},
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
			method:      "create",
			expectedErr: errors.New(e.New("error: must specify " + cst.DataName)),
		},
		{
			name:        "Update fails no name",
			args:        []string{"--desc", "updated role"},
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
			method:      "update",
			expectedErr: errors.New(e.New("error: must specify " + cst.DataName)),
		},
		{
			name:        "Create fails external ID is missing",
			roleName:    "role2",
			provider:    "aws-dev",
			args:        []string{"--name", "role2", "--provider", "aws-dev"},
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
			method:      "create",
			expectedErr: errors.New(e.New("error: must specify both provider and external ID for third-party roles")),
		},
		{
			name:        "Create fails provider is missing",
			roleName:    "role2",
			externalID:  "1234",
			args:        []string{"--name", "role2", "--external-id", "1234"},
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
			method:      "create",
			expectedErr: errors.New(e.New("error: must specify both provider and external ID for third-party roles")),
		},
		{
			name:        "Successful 3rd party role create",
			roleName:    "role1",
			provider:    "aws-dev",
			externalID:  "1234",
			args:        []string{"--name", "role2", "--provider", "aws-dev", "--external-id", "1234"},
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
			method:      "create",
		},
	}

	viper.Set(cst.Version, "v1")
	for _, tt := range testCase {
		viper.Set(cst.DataName, tt.roleName)
		viper.Set(cst.DataProvider, tt.provider)
		viper.Set(cst.DataExternalID, tt.externalID)
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
