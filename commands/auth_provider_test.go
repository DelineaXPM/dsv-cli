package cmd

import (
	e "errors"
	"net/http"
	"testing"
	cst "thy/constants"
	"thy/errors"
	"thy/fake"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestHandleAuthProviderReadCommand(t *testing.T) {
	testCase := []struct {
		name        string
		args        []string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"success",
			[]string{"--name", "aws-dev"},
			[]byte(`test`),
			[]byte(`test`),
			nil},
		{
			"api-error",
			[]string{"--name", "missing"},
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("item doesn't exist")),
		},
	}

	_, err := GetAuthProviderReadCmd()
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

			p := AuthProvider{request: req, outClient: acmd}
			_ = p.handleAuthProviderReadCmd(tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})
	}
}

func TestHandleAuthProviderUpsertCommand(t *testing.T) {
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
			[]string{"--path", "oh-hey"},
			[]byte(`test`),
			[]byte(`test`),
			http.MethodPost,
			[]string{"type", "accountid"},
			nil,
		},

		{
			"success-update",
			[]string{"--name", "oh-hello"},
			[]byte(`test`),
			[]byte(`test`),
			http.MethodPut,
			[]string{"type", "accountid"},
			nil,
		},

		{
			"fail-validation-error",
			[]string{"--name", "azure-demo"},
			[]byte(`test`),
			[]byte(`test`),
			http.MethodPut,
			[]string{},
			errors.New(e.New("--type must be set")),
		},
	}

	_, err := GetAuthProviderCreateCmd()
	assert.Nil(t, err)

	_, err = GetAuthProviderUpdateCmd()
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

			p := &AuthProvider{req, writer, nil}
			viper.Set(cst.LastCommandKey, tt.method)

			if len(tt.dataAction) == 2 {
				viper.Set(cst.DataType, tt.dataAction[0])
				viper.Set(cst.DataAccountID, tt.dataAction[1])
			}

			_ = p.handleAuthProviderUpsert(tt.args)

			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
			viper.Set(cst.DataType, "")
			viper.Set(cst.DataAccountID, "")
		})
	}
}

func TestHandleAuthProviderDeleteCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        []string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"success",
			[]string{"--name", "azure-dev"},
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},
		{
			"validation error",
			[]string{"--name", "missing"},
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("item doesn't exist")),
		},
	}

	_, err := GetAuthProviderDeleteCmd()
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

			p := &AuthProvider{req, writer, nil}
			_ = p.handleAuthProviderDeleteCmd(tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})
	}
}

func TestHandleAuthProviderSearchCommand(t *testing.T) {
	testCase := []struct {
		name        string
		args        []string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"success",
			[]string{"-q", "azure"},
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

	_, err := GetAuthProviderSearchCommand()
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

			p := AuthProvider{request: req, outClient: acmd}
			_ = p.handleAuthProviderSearchCommand(tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})
	}
}

func TestHandleAuthProviderRollbackCommand(t *testing.T) {
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

	_, err := GetAuthProviderRollbackCmd()
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

			p := AuthProvider{request: req, outClient: acmd}
			_ = p.handleAuthProviderRollbackCmd(tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}
