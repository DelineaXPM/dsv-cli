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

func TestHandleGroupReadCmd(t *testing.T) {

	testCase := []struct {
		name        string
		args        string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Happy path",
			"group1",
			[]byte(`test`),
			[]byte(`test`),
			nil},
		{
			"api Error",
			"group1",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error")),
		},
		{
			"No group passed",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify " + cst.DataGroupName)),
		},
	}

	_, err := GetGroupReadCmd()
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

			u := Group{req, acmd}
			_ = u.handleGroupReadCmd([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})

	}
}

func TestHandleGroupCreateCmd(t *testing.T) {

	testCase := []struct {
		name        string
		args        []string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Happy path",
			[]string{"groupName"},
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},

		{
			"no group ",
			[]string{""},
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify " + cst.DataGroupName)),
		},
	}

	_, err := GetGroupCreateCmd()
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

			u := Group{req, acmd}
			_ = u.handleCreateCmd(tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})

	}
}

func TestHandleAddMemberGroupCmd(t *testing.T) {

	testCase := []struct {
		name        string
		args        []string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Happy path",
			[]string{"groupName"},
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},

		{
			"no group ",
			[]string{""},
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("--group-name required")),
		},
	}

	_, err := GetMemberGroupsCmd()
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
			if tt.name == "Happy path" {
				viper.Set(cst.DataGroupName, "fakegroup")
			}
			u := Group{req, acmd}
			_ = u.handleAddMembersCmd(tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, string(data), string(tt.out))
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})

	}
}

func TestHandleDeleteMemberGroupCmd(t *testing.T) {

	testCase := []struct {
		name        string
		args        []string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Happy path",
			[]string{"groupName"},
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},

		{
			"no group ",
			[]string{""},
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify " + cst.DataGroupName)),
		},
	}

	_, err := GetDeleteMembersCmd()
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

			u := Group{req, acmd}
			_ = u.handleDeleteMemberCmd(tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})

	}
}

func TestHandleGroupDeleteCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Happy path",
			"group1",
			[]byte(`test`),
			[]byte(`test`),
			nil},
		{
			"api Error",
			"group1",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error")),
		},
		{
			"No DataUsername",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify " + cst.DataGroupName)),
		},
	}

	_, err := GetGroupDeleteCmd()
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

			u := Group{req, acmd}
			_ = u.handleGroupDeleteCmd([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})

	}
}

func TestHandleMemberGroupCmd(t *testing.T) {

	testCase := []struct {
		name        string
		args        string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Happy path",
			"group1",
			[]byte(`test`),
			[]byte(`test`),
			nil},
		{
			"api Error",
			"group1",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error")),
		},
		{
			"No group passed",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify " + cst.DataUsername)),
		},
	}

	_, err := GetMemberGroupsCmd()
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

			u := Group{req, acmd}
			_ = u.handleUsersGroupReadCmd([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})

	}
}

func TestHandleGroupSearchCmd(t *testing.T) {

	testCase := []struct {
		name        string
		args        string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Happy path",
			"group1",
			[]byte(`test`),
			[]byte(`test`),
			nil},
		{
			"api Error",
			"group1",
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

	_, err := GetGroupSearchCmd()
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

			u := Group{req, acmd}
			_ = u.handleGroupSearchCmd([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})

	}
}
