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

func TestHandleEngineReadCmd(t *testing.T) {
	testCases := []struct {
		name        string
		engineName  string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Success",
			"engine1",
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},
		{
			"No engine name passed",
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
		viper.Set(cst.DataName, tt.engineName)
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

			eh := engineHandler{req, client}
			_ = eh.handleRead(nil)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleEngineCreateCmd(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		engineName  string
		poolName    string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Success",
			[]string{"--name", "engine1", "--pool-name", "pool1"},
			"engine1",
			"pool1",
			[]byte(`test`),
			[]byte(`test`),
			nil,
		},
		{
			"No engine name passed",
			[]string{"--pool-name", "pool1"},
			"",
			"pool1",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify engine name and pool name")),
		},
		{
			"No pool name passed",
			[]string{"--name", "engine1"},
			"engine1",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify engine name and pool name")),
		},
	}

	_, err := GetEngineCreateCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tt := range testCases {
		viper.Set(cst.DataName, tt.engineName)
		viper.Set(cst.DataPoolName, tt.poolName)
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

			eh := engineHandler{req, client}
			_ = eh.handleCreate(tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleEngineDeleteCmd(t *testing.T) {
	testCases := []struct {
		name        string
		engineName  string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Success",
			"engine1",
			[]byte(`test`),
			[]byte(`test`),
			nil},
		{
			"No engine name passed",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify " + cst.DataName)),
		},
	}

	_, err := GetEngineDeleteCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tt := range testCases {
		viper.Set(cst.DataName, tt.engineName)
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

			eh := engineHandler{req, client}
			_ = eh.handleDelete(nil)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}
