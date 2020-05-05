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

func TestHandlePolicyReadCommand(t *testing.T) {
	testCase := []struct {
		name        string
		args        []string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"success",
			[]string{"--path", "secrets/servers/db"},
			[]byte(`test`),
			[]byte(`test`),
			nil},
		{
			"api Error",
			[]string{"--path", "secrets/servers/none"},
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("policy doesn't exist")),
		},
	}

	_, err := GetPolicyReadCmd()
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

			p := Policy{request: req, outClient: acmd}
			_ = p.handlePolicyReadCmd(tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})
	}
}

func TestHandlePolicyEditCommand(t *testing.T) {
	testCase := []struct {
		name         string
		args         []string
		out          []byte
		editResponse []byte
		expectedErr  *errors.ApiError
		apiError     *errors.ApiError
		editError    *errors.ApiError
	}{
		{
			"success",
			[]string{"--path", "secrets/servers/i-do-exist"},
			[]byte(`test data`),
			[]byte(`test data`),
			nil,
			nil,
			nil,
		},
		{
			"error-missing-policy",
			[]string{"--path", "i-don't-exist"},
			[]byte(`test data`),
			nil,
			errors.New(e.New("missing item at path")),
			errors.New(e.New("missing item at path")),
			nil,
		},
	}

	_, err := GetPolicyEditCmd()
	assert.Nil(t, err)

	for _, tt := range testCase {
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
				return tt.out, tt.apiError
			}

			p := &Policy{request: req, outClient: writer}
			p.edit = func(bytes2 []byte, d dataFunc, apiError *errors.ApiError, retry bool) (bytes []byte, apiError2 *errors.ApiError) {
				_, _ = d([]byte(`config`))
				return tt.editResponse, tt.editError
			}

			_ = p.handlePolicyEditCmd(tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})
	}
}

func TestHandlePolicyUpsertCommand(t *testing.T) {
	testCase := []struct {
		name        string
		args        []string
		apiResponse []byte
		out         []byte
		method      string
		dataAction  []string
		expectedErr *errors.ApiError
	}{
		{
			"success-create",
			[]string{"--path", "secrets/servers/dbs"},
			[]byte(`test`),
			[]byte(`test`),
			"POST",
			[]string{"actions", "subjects", "effect"},
			nil,
		},

		{
			"success-update",
			[]string{"--path", "secrets/servers/dbs"},
			[]byte(`test`),
			[]byte(`test`),
			"PUT",
			[]string{"actions", "subjects", "effect"},
			nil,
		},

		{
			"fail-validation-error",
			[]string{"--path", "secrets/servers/dbs"},
			[]byte(`test`),
			[]byte(`test`),
			"PUT",
			[]string{},
			errors.New(e.New("--actions must be set")),
		},
	}

	_, err := GetPolicyCreateCmd()
	assert.Nil(t, err)

	_, err = GetPolicyUpdateCmd()
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

			req := &fake.FakeClient{}
			req.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.out, tt.expectedErr
			}

			p := &Policy{req, writer, nil}
			viper.Set(cst.LastCommandKey, tt.method)

			if len(tt.dataAction) == 3 {
				viper.Set(cst.DataAction, tt.dataAction[0])
				viper.Set(cst.DataSubject, tt.dataAction[1])
				viper.Set(cst.DataEffect, tt.dataAction[2])
				viper.Set(cst.ID, "ID")
				viper.Set(cst.DataCidr, "135.104.0.0/32")
			} else {
				viper.Set(cst.DataAction, "")
			}

			_ = p.handlePolicyUpsertCmd(tt.args)

			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
			viper.Set(cst.ID, "")
			viper.Set(cst.DataCidr, "")
		})
	}
}

func TestHandlePolicyDeleteCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        []string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"success",
			[]string{"--path", "secrets/servers/db"},
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},
		{
			"validation error",
			[]string{"--path", "secrets/servers/db/missing-polciy"},
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("item doesn't exist")),
		},
	}

	_, err := GetPolicyDeleteCmd()
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

			req := &fake.FakeClient{}
			req.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.out, tt.expectedErr
			}

			p := &Policy{req, writer, nil}
			_ = p.handlePolicyDeleteCmd(tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})
	}
}

func TestHandlePolicySearchCommand(t *testing.T) {
	testCase := []struct {
		name        string
		args        []string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"success",
			[]string{"-q", "servers"},
			[]byte(`test`),
			[]byte(`test`),
			nil},
		{
			"error",
			[]string{"-q", "what happens when a server error occurs?"},
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("internal 500 error")),
		},
	}

	_, err := GetPolicySearchCommand()
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

			p := Policy{request: req, outClient: acmd}
			_ = p.handlePolicySearchCommand(tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})
	}
}

func TestHandlePolicyRollbackCommand(t *testing.T) {
	testCase := []struct {
		name        string
		args        []string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"success (no version passed in)",
			[]string{"azure-dev"},
			[]byte(`test`),
			[]byte(`{"version": "4"}`),
			nil},
		{
			"error (no version passed in)",
			[]string{"-q", "what happens when a server error occurs?"},
			[]byte(`test`),
			[]byte(`{"someData": "hello"}`),
			errors.NewS("version not found"),
		},
	}

	_, err := GetPolicyRollbackCmd()
	assert.Nil(t, err)

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

			p := Policy{request: req, outClient: acmd}
			_ = p.handlePolicyRollbackCmd(tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}
