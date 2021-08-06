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

func TestHandlePoolReadCmd(t *testing.T) {
	testCases := []struct {
		name        string
		poolName    string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Success",
			"pool1",
			[]byte(`test`),
			[]byte(`test`),
			nil},
		{
			"No pool name passed",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify " + cst.DataName)),
		},
	}

	_, err := GetPoolReadCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tt := range testCases {
		viper.Set(cst.DataName, tt.poolName)
		t.Run(tt.name, func(t *testing.T) {
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

			p := poolHandler{req, client}
			_ = p.handleRead(nil)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandlePoolCreateCmd(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		poolName    string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Success",
			[]string{"--name", "pool1"},
			"pool1",
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},
	}

	_, err := GetPoolCreateCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tt := range testCases {
		viper.Set(cst.DataName, tt.poolName)
		t.Run(tt.name, func(t *testing.T) {
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

			p := poolHandler{req, client}
			_ = p.handleCreate(tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandlePoolDeleteCmd(t *testing.T) {
	testCases := []struct {
		name        string
		poolName    string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Success",
			"pool1",
			[]byte(`test`),
			[]byte(`test`),
			nil},
		{
			"No pool name passed",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify " + cst.DataName)),
		},
	}

	_, err := GetPoolDeleteCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tt := range testCases {
		viper.Set(cst.DataName, tt.poolName)
		t.Run(tt.name, func(t *testing.T) {
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

			p := poolHandler{req, client}
			_ = p.handleDelete(nil)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}
