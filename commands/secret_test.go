package cmd

import (
	e "errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	cst "thy/constants"
	"thy/errors"
	"thy/fake"
	"thy/store"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestHandleDescribeCmd(t *testing.T) {
	testCase := []struct {
		name          string
		arg           []string
		cacheStrategy string
		out           []byte
		storeType     string
		expectedErr   *errors.ApiError
		apiError      *errors.ApiError
		flags         []struct {
			flag  string
			value string
		}
	}{
		{
			"Happy Path no cacheStrategy",
			[]string{"path1"},
			"",
			[]byte(`test data`),
			"",
			nil,
			nil,
			nil,
		},
		{
			"Happy Path no cacheStrategy",
			[]string{"140a372c-7d37-11eb-bc08-00155d19ad95"},
			"",
			[]byte(`test data`),
			"",
			nil,
			nil,
			nil,
		},
		{
			"Happy Path no cacheStrategy",
			[]string{""},
			"",
			[]byte(`test data`),
			"",
			nil,
			nil,
			[]struct {
				flag  string
				value string
			}{
				{
					cst.ID,
					"140a372c-7d37-11eb-bc08-00155d19ad95",
				},
			},
		},
		{
			"Happy Path cache.server cacheStrategy",
			[]string{"path1"},
			"cache.server",
			[]byte(`test data from cache`),
			"",
			nil,
			nil,
			nil,
		},
		{
			"Happy Path cache.server cacheStrategy",
			[]string{"140a372c-7d37-11eb-bc08-00155d19ad95"},
			"cache.server",
			[]byte(`test data from cache`),
			"",
			nil,
			nil,
			nil,
		},
		{
			"Happy Path cache.server cacheStrategy",
			[]string{""},
			"cache.server",
			[]byte(`test data from cache`),
			"",
			nil,
			nil,
			[]struct {
				flag  string
				value string
			}{
				{
					cst.ID,
					"140a372c-7d37-11eb-bc08-00155d19ad95",
				},
			},
		},
		{
			"Happy Path server.cache cacheStrategy",
			[]string{"path1"},
			"server.cache",
			[]byte(`test data from cache`),
			"",
			nil,
			errors.New(e.New("error")),
			nil,
		},
		{
			"Happy Path server.cache cacheStrategy",
			[]string{"140a372c-7d37-11eb-bc08-00155d19ad95"},
			"server.cache",
			[]byte(`test data from cache`),
			"",
			nil,
			errors.New(e.New("error")),
			nil,
		},
		{
			"Happy Path server.cache cacheStrategy",
			[]string{""},
			"server.cache",
			[]byte(`test data from cache`),
			"",
			nil,
			errors.New(e.New("error")),
			[]struct {
				flag  string
				value string
			}{
				{
					cst.ID,
					"140a372c-7d37-11eb-bc08-00155d19ad95",
				},
			},
		},
	}

	_, err := GetDescribeCmd()
	assert.Nil(t, err)

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			for _, f := range tt.flags {
				viper.Set(f.flag, f.value)
			}

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

			st := &fake.FakeStore{}

			st.GetStub = func(s string, d interface{}) *errors.ApiError {
				sData, ok := d.(*secretData)
				if ok {
					sData.Date = time.Now().Add(60 * time.Minute)
					sData.Data = tt.out
				}
				return tt.expectedErr
			}

			st.StoreStub = func(s string, i interface{}) *errors.ApiError {
				return tt.expectedErr
			}

			viper.Set(cst.StoreType, tt.storeType)
			viper.Set(cst.CacheStrategy, tt.cacheStrategy)
			sec := &Secret{req, writer, store.GetStore, nil, cst.NounSecret}
			sec.getStore = func(stString string) (i store.Store, apiError *errors.ApiError) {
				return st, nil
			}
			_ = sec.handleDescribeCmd(tt.arg)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}

			for _, f := range tt.flags {
				viper.Set(f.flag, "")
			}
		})
	}
}

func TestHandleSecretSearchCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
		flags       []struct {
			flag  string
			value string
		}
	}{
		{
			"Happy path",
			"user1",
			[]byte(`test`),
			[]byte(`test`),
			nil,
			nil,
		},
		{
			"api Error",
			"user1",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error")),
			nil,
		},
		{
			"No Search query",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error: must specify " + cst.Query)),
			nil,
		},
	}

	_, err := GetSecretSearchCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tt := range testCase {

		t.Run(tt.name, func(t *testing.T) {
			for _, f := range tt.flags {
				viper.Set(f.flag, f.value)
			}

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

			sec := &Secret{req, writer, store.GetStore, nil, cst.NounSecret}
			_ = sec.handleSecretSearchCmd([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}

			for _, f := range tt.flags {
				viper.Set(f.flag, "")
			}
		})

	}
}

func TestHandleDeleteCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
		flags       []struct {
			flag  string
			value string
		}
	}{
		{
			"Happy path",
			"user1",
			[]byte(`test`),
			[]byte(`test`),
			nil,
			nil,
		},
		{
			"Happy ID",
			"140a372c-7d37-11eb-bc08-00155d19ad95",
			[]byte(`test`),
			[]byte(`test`),
			nil,
			nil,
		},
		{
			"Happy ID",
			"",
			[]byte(`test`),
			[]byte(`test`),
			nil,
			[]struct {
				flag  string
				value string
			}{
				{
					cst.ID,
					"140a372c-7d37-11eb-bc08-00155d19ad95",
				},
			},
		},
		{
			"api Error",
			"user1",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error")),
			nil,
		},
		{
			"api Error",
			"140a372c-7d37-11eb-bc08-00155d19ad95",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error")),
			nil,
		},
		{
			"api Error",
			"",
			[]byte(`test`),
			[]byte(`test`),
			errors.New(e.New("error")),
			[]struct {
				flag  string
				value string
			}{
				{
					cst.ID,
					"140a372c-7d37-11eb-bc08-00155d19ad95",
				},
			},
		},
	}

	_, err := GetDeleteCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			for _, f := range tt.flags {
				viper.Set(f.flag, f.value)
			}

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

			sec := &Secret{req, writer, store.GetStore, nil, cst.NounSecret}
			_ = sec.handleDeleteCmd([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}

			for _, f := range tt.flags {
				viper.Set(f.flag, "")
			}
		})
	}
}

func TestHandleRollbackCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
		flags       []struct {
			flag  string
			value string
		}
	}{
		{
			"success (no version passed in) (ID)",
			"140a372c-7d37-11eb-bc08-00155d19ad95",
			[]byte(`test`),
			[]byte(`{"version": "4"}`),
			nil,
			nil,
		},
		{
			"success (no version passed in) (ID)",
			"",
			[]byte(`test`),
			[]byte(`{"version": "4"}`),
			nil,
			[]struct {
				flag  string
				value string
			}{
				{
					cst.ID,
					"140a372c-7d37-11eb-bc08-00155d19ad95",
				},
			},
		},
		{
			"success (no version passed in) (path)",
			"azure-dev",
			[]byte(`test`),
			[]byte(`{"version": "4"}`),
			nil,
			nil,
		},
		{
			"error (no version passed in) (ID)",
			"140a372c-7d37-11eb-bc08-00155d19ad95",
			[]byte(`test`),
			[]byte(`{"someData": "hello"}`),
			errors.NewS("version not found"),
			nil,
		},
		{
			"error (no version passed in) (ID)",
			"",
			[]byte(`test`),
			[]byte(`{"someData": "hello"}`),
			errors.NewS("version not found"),
			[]struct {
				flag  string
				value string
			}{
				{
					cst.ID,
					"140a372c-7d37-11eb-bc08-00155d19ad95",
				},
			},
		},
		{
			"error (no version passed in) (path)",
			"azure-dev",
			[]byte(`test`),
			[]byte(`{"someData": "hello"}`),
			errors.NewS("version not found"),
			nil,
		},
	}

	_, err := GetRollbackCmd()
	assert.Nil(t, err)

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			for _, f := range tt.flags {
				viper.Set(f.flag, f.value)
			}

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

			sec := &Secret{req, writer, store.GetStore, nil, cst.NounSecret}
			_ = sec.handleRollbackCmd([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}

			for _, f := range tt.flags {
				viper.Set(f.flag, "")
			}
		})
	}
}

func TestHandleReadCmd(t *testing.T) {
	testCase := []struct {
		name          string
		arg           []string
		cacheStrategy string
		out           []byte
		storeType     string
		expectedErr   *errors.ApiError
		apiError      *errors.ApiError
		flags         []struct {
			flag  string
			value string
		}
	}{
		{
			"Happy Path no cacheStrategy (ID)",
			[]string{"140a372c-7d37-11eb-bc08-00155d19ad95"},
			"",
			[]byte(`test data`),
			"",
			nil,
			nil,
			nil,
		},
		{
			"Happy Path no cacheStrategy (ID)",
			[]string{""},
			"",
			[]byte(`test data`),
			"",
			nil,
			nil,
			[]struct {
				flag  string
				value string
			}{
				{
					cst.ID,
					"140a372c-7d37-11eb-bc08-00155d19ad95",
				},
			},
		},
		{
			"Happy Path no cacheStrategy",
			[]string{"path1"},
			"",
			[]byte(`test data`),
			"",
			nil,
			nil,
			nil,
		},
		{
			"Happy Path cache.server cacheStrategy",
			[]string{"path1"},
			"cache.server",
			[]byte(`test data from cache`),
			"",
			nil,
			nil,
			nil,
		},
		{
			"Happy Path cache.server cacheStrategy",
			[]string{""},
			"cache.server",
			[]byte(`test data from cache`),
			"",
			nil,
			nil,
			[]struct {
				flag  string
				value string
			}{
				{
					cst.ID,
					"140a372c-7d37-11eb-bc08-00155d19ad95",
				},
			},
		},
		{
			"Happy Path cache.server cacheStrategy",
			[]string{"140a372c-7d37-11eb-bc08-00155d19ad95"},
			"cache.server",
			[]byte(`test data from cache`),
			"",
			nil,
			nil,
			nil,
		},
		{
			"Happy Path server.cache cacheStrategy",
			[]string{"path1"},
			"server.cache",
			[]byte(`test data from cache`),
			"",
			nil,
			errors.New(e.New("error")),
			nil,
		},
		{
			"Happy Path server.cache cacheStrategy",
			[]string{"path1"},
			"server.cache",
			[]byte(`test data from cache`),
			"",
			nil,
			errors.New(e.New("error")),
			[]struct {
				flag  string
				value string
			}{
				{
					cst.ID,
					"140a372c-7d37-11eb-bc08-00155d19ad95",
				},
			},
		},
		{
			"Happy Path server.cache cacheStrategy",
			[]string{"140a372c-7d37-11eb-bc08-00155d19ad95"},
			"server.cache",
			[]byte(`test data from cache`),
			"",
			nil,
			errors.New(e.New("error")),
			nil,
		},
	}

	_, err := GetReadCmd()
	assert.Nil(t, err)

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			for _, f := range tt.flags {
				viper.Set(f.flag, f.value)
			}

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

			st := &fake.FakeStore{}

			st.GetStub = func(s string, d interface{}) *errors.ApiError {
				sData, ok := d.(*secretData)
				if ok {
					sData.Date = time.Now().Add(60 * time.Minute)
					sData.Data = tt.out
				}
				return tt.expectedErr
			}

			st.StoreStub = func(s string, i interface{}) *errors.ApiError {
				return tt.expectedErr
			}
			viper.Set(cst.Version, "v1")
			viper.Set(cst.StoreType, tt.storeType)
			viper.Set(cst.CacheStrategy, tt.cacheStrategy)
			sec := &Secret{req, writer, store.GetStore, nil, cst.NounSecret}
			sec.getStore = func(stString string) (i store.Store, apiError *errors.ApiError) {
				return st, nil
			}
			_ = sec.handleReadCmd(tt.arg)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
			viper.Set(cst.StoreType, "")
			viper.Set(cst.CacheStrategy, "")

			for _, f := range tt.flags {
				viper.Set(f.flag, "")
			}
		})
	}
}

func TestHandleUpsertCmd(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		out         []byte
		method      string
		expectedErr *errors.ApiError
		flags       []struct {
			flag  string
			value string
		}
	}{
		{
			"Happy path POST",
			[]string{"mySecret", "--desc", "new description"},
			[]byte(`test`),
			"create",
			nil,
			[]struct {
				flag  string
				value string
			}{
				{cst.DataDescription, "new description"},
			},
		},
		{
			"Happy path PUT (ID)",
			[]string{"140a372c-7d37-11eb-bc08-00155d19ad95", "--desc", "new description"},
			[]byte(`test`),
			"update",
			nil,
			[]struct {
				flag  string
				value string
			}{
				{cst.DataDescription, "new description"},
			},
		},
		{
			"Happy path PUT (ID)",
			[]string{"--id", "140a372c-7d37-11eb-bc08-00155d19ad95", "--desc", "new description"},
			[]byte(`test`),
			"update",
			nil,
			[]struct {
				flag  string
				value string
			}{
				{cst.ID, "140a372c-7d37-11eb-bc08-00155d19ad95"},
				{cst.DataDescription, "new description"},
			},
		},
		{
			"Happy path PUT (path)",
			[]string{"mySecret", "--description", "new description"},
			[]byte(`test`),
			"update",
			nil,
			[]struct {
				flag  string
				value string
			}{
				{cst.DataDescription, "new description"},
			},
		},
		{
			"no path",
			[]string{"--description", "new description"},
			[]byte(`test`),
			"",
			errors.New(e.New("error: must specify --id or --path (or [path])")),
			nil,
		},
	}

	_, err := GetUpdateCmd()
	assert.Nil(t, err)

	_, err = GetEditCmd()
	assert.Nil(t, err)

	_, err = GetCreateCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			for _, f := range tt.flags {
				viper.Set(f.flag, f.value)
			}
			viper.Set(cst.LastCommandKey, tt.method)

			var data []byte
			var err *errors.ApiError

			writer := &fake.FakeOutClient{}
			writer.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}
			writer.FailEStub = func(apiError *errors.ApiError) { err = apiError }

			req := &fake.FakeClient{}
			req.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.out, tt.expectedErr
			}

			sec := &Secret{req, writer, store.GetStore, nil, cst.NounSecret}
			_ = sec.handleUpsertCmd(tt.args)

			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
			for _, f := range tt.flags {
				viper.Set(f.flag, "")
			}
			viper.Set(cst.LastCommandKey, "")
		})
	}
}

// UIMock is a cli.UI mock which will go somewhere else later.
type UIMock struct {
	stub func(string) (string, error)
}

func (u *UIMock) Ask(s string) (string, error)       { return u.stub(s) }
func (u *UIMock) AskSecret(s string) (string, error) { return u.stub(s) }
func (u *UIMock) Output(string)                      {}
func (u *UIMock) Info(string)                        {}
func (u *UIMock) Error(string)                       {}
func (u *UIMock) Warn(string)                        {}

func TestHandleCreateWorkflow(t *testing.T) {
	const tenantName = "createworkflowtest"

	// uiStep contains prefix of the expected line and answer to it.
	type uiStep struct {
		inPrefix string
		out      string
	}

	cases := []struct {
		name         string
		steps        []*uiStep
		expectedURI  string
		expectedBody *secretUpsertBody
	}{
		{
			name: "Add description and k/v attributes, but skip data",
			steps: []*uiStep{
				{"Path", "key1"},
				{"Description", "some key"},
				{"Add Attributes", "2"},
				{"Key", "aa"},
				{"Value", "11"},
				{"Add more?", "no"},
				{"Add Data", "1"},
			},
			expectedURI: "https://createworkflowtest.secretsvaultcloud.com/v1/secrets/key1",
			expectedBody: &secretUpsertBody{
				Description: "some key",
				Data:        map[string]interface{}{},
				Attributes: map[string]interface{}{
					"aa": "11",
				},
			},
		},
		{
			name: "Add description and json attributes, but skip data",
			steps: []*uiStep{
				{"Path", "key1"},
				{"Description", "some key"},
				{"Add Attributes", "3"},
				{"Attributes", "{\"attr1\":\"value1\"}"},
				{"Add Data", "1"},
			},
			expectedURI: "https://createworkflowtest.secretsvaultcloud.com/v1/secrets/key1",
			expectedBody: &secretUpsertBody{
				Description: "some key",
				Data:        map[string]interface{}{},
				Attributes: map[string]interface{}{
					"attr1": "value1",
				},
			},
		},
		{
			name: "Add description and k/v data, but skip attributes",
			steps: []*uiStep{
				{"Path", "key1"},
				{"Description", "some key"},
				{"Add Attributes", "1"},
				{"Add Data", "2"},
				{"Key", "apikey"},
				{"Value", "testtest"},
				{"Add more?", "no"},
			},
			expectedURI: "https://createworkflowtest.secretsvaultcloud.com/v1/secrets/key1",
			expectedBody: &secretUpsertBody{
				Description: "some key",
				Data: map[string]interface{}{
					"apikey": "testtest",
				},
				Attributes: map[string]interface{}{},
			},
		},
		{
			name: "Add description and json data, but skip attributes",
			steps: []*uiStep{
				{"Path", "key1"},
				{"Description", "some key"},
				{"Add Attributes", "1"},
				{"Add Data", "3"},
				{"Data", "{\"apikey\":\"testtesttest\"}"},
			},
			expectedURI: "https://createworkflowtest.secretsvaultcloud.com/v1/secrets/key1",
			expectedBody: &secretUpsertBody{
				Description: "some key",
				Data: map[string]interface{}{
					"apikey": "testtesttest",
				},
				Attributes: map[string]interface{}{},
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set(cst.Tenant, tenantName)

			cnt := 0
			uimock := &UIMock{
				stub: func(s string) (string, error) {
					step := tt.steps[cnt]
					cnt++

					if strings.HasPrefix(s, step.inPrefix) {
						return step.out, nil
					}
					return "", fmt.Errorf("unexpected line: %s", s)
				},
			}

			var (
				reqMethod string
				reqURI    string
				reqData   interface{}
			)
			req := &fake.FakeClient{}
			req.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				reqMethod = s
				reqURI = s2
				reqData = i
				return nil, nil
			}

			sec := &Secret{request: req, secretType: cst.NounSecret}
			_, apiErr := sec.handleCreateWorkflow(uimock)

			viper.Set(cst.Tenant, "")

			assert.Nil(t, apiErr)
			assert.Equal(t, "POST", reqMethod)
			assert.Equal(t, tt.expectedURI, reqURI)
			assert.IsType(t, &secretUpsertBody{}, reqData)
			assert.Equal(t, tt.expectedBody, reqData)
		})
	}
}

func TestHandleUpdateWorkflow(t *testing.T) {
	const tenantName = "updateworkflowtest"

	// uiStep contains prefix of the expected line and answer to it.
	type uiStep struct {
		inPrefix string
		out      string
	}

	cases := []struct {
		name            string
		steps           []*uiStep
		getSecretAPIErr *errors.ApiError
		expectedURI     string
		expectedBody    *secretUpsertBody
		shouldFail      bool
	}{
		{
			name: "Update description only",
			steps: []*uiStep{
				{"Path", "key1"},
				{"Update description", "yes"},
				{"Description", "new description"},
				{"Overwrite existing attributes and data?", "no"},
				{"Update attributes?", "1"},
				{"Update data?", "1"},
			},
			expectedURI: "https://updateworkflowtest.secretsvaultcloud.com/v1/secrets/key1",
			expectedBody: &secretUpsertBody{
				Description: "new description",
				Data:        map[string]interface{}{},
				Attributes:  map[string]interface{}{},
				Overwrite:   false,
			},
			shouldFail: false,
		},
		{
			name: "Secret does not exist",
			steps: []*uiStep{
				{"Path", "key1"},
			},
			getSecretAPIErr: errors.NewS("some message").WithResponse(&http.Response{StatusCode: http.StatusNotFound}),
			shouldFail:      true,
		},
		{
			name: "No permission to read secret",
			steps: []*uiStep{
				{"Path", "key1"},
				{"You are not allowed to read secret under that path. Do you want to continue?", "yes"},
				{"Update description", "yes"},
				{"Description", "new description"},
				{"Overwrite existing attributes and data?", "no"},
				{"Update attributes?", "1"},
				{"Update data?", "1"},
			},
			getSecretAPIErr: errors.NewS("some message").WithResponse(&http.Response{StatusCode: http.StatusForbidden}),
			expectedURI:     "https://updateworkflowtest.secretsvaultcloud.com/v1/secrets/key1",
			expectedBody: &secretUpsertBody{
				Description: "new description",
				Data:        map[string]interface{}{},
				Attributes:  map[string]interface{}{},
				Overwrite:   false,
			},
			shouldFail: false,
		},
		{
			name: "Nothing to update",
			steps: []*uiStep{
				{"Path", "key1"},
				{"Update description", "no"},
				{"Overwrite existing attributes and data?", "no"},
				{"Update attributes?", "1"},
				{"Update data?", "1"},
			},
			expectedURI:  "",
			expectedBody: nil,
			shouldFail:   false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set(cst.Tenant, tenantName)

			cnt := 0
			uimock := &UIMock{
				stub: func(s string) (string, error) {
					if len(tt.steps) <= cnt {
						return "", fmt.Errorf("unexpected line: %s", s)
					}
					step := tt.steps[cnt]
					cnt++

					if strings.HasPrefix(s, step.inPrefix) {
						return step.out, nil
					}
					return "", fmt.Errorf("unexpected line: %s", s)
				},
			}

			var (
				reqMethod string
				reqURI    string
				reqData   interface{}
			)
			req := &fake.FakeClient{}
			req.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				if s == http.MethodGet {
					if tt.getSecretAPIErr != nil {
						return nil, tt.getSecretAPIErr
					}
					return nil, nil
				}

				reqMethod = s
				reqURI = s2
				reqData = i
				return nil, nil
			}

			sec := &Secret{request: req, secretType: cst.NounSecret}
			_, apiErr := sec.handleUpdateWorkflow(uimock)

			viper.Set(cst.Tenant, "")

			if tt.shouldFail {
				assert.NotNil(t, apiErr)
				return
			}
			assert.Nil(t, apiErr)

			if tt.expectedBody != nil {
				assert.Equal(t, "PUT", reqMethod)
				assert.Equal(t, tt.expectedURI, reqURI)
				assert.IsType(t, &secretUpsertBody{}, reqData)
				assert.Equal(t, tt.expectedBody, reqData)
			}
		})
	}
}

func TestHandleBustCacheCmd(t *testing.T) {
	testCase := []struct {
		name        string
		arg         []string
		out         []byte
		storeType   string
		expectedErr *errors.ApiError
		flags       []struct {
			flag  string
			value string
		}
	}{
		{
			"Happy Path",
			[]string{},
			nil,
			"",
			nil,
			nil,
		},
		{
			"Error",
			[]string{},
			nil,
			"",
			errors.New(e.New("error")),
			nil,
		},
	}

	_, err := GetBustCacheCmd()
	assert.Nil(t, err)

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			for _, f := range tt.flags {
				viper.Set(f.flag, f.value)
			}

			writer := &fake.FakeOutClient{}
			var data []byte
			var err *errors.ApiError
			writer.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			st := &fake.FakeStore{}

			st.WipeStub = func(s string) *errors.ApiError {
				return tt.expectedErr
			}

			viper.Set(cst.StoreType, tt.storeType)
			sec := &Secret{nil, writer, store.GetStore, nil, cst.NounSecret}
			sec.getStore = func(stString string) (i store.Store, apiError *errors.ApiError) {
				return st, nil
			}
			_ = sec.handleBustCacheCmd(tt.arg)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}

			for _, f := range tt.flags {
				viper.Set(f.flag, "")
			}
		})
	}
}

func TestHandleEditCmd(t *testing.T) {
	testCase := []struct {
		name         string
		arg          []string
		out          []byte
		editResponse []byte
		expectedErr  *errors.ApiError
		apiError     *errors.ApiError
		editError    *errors.ApiError
		flags        []struct {
			flag  string
			value string
		}
	}{
		{
			"Happy Path",
			[]string{"path1"},
			[]byte(`test data`),
			[]byte(`test data`),
			nil,
			nil,
			nil,
			nil,
		},
		{
			"Happy Path",
			[]string{"140a372c-7d37-11eb-bc08-00155d19ad95"},
			[]byte(`test data`),
			[]byte(`test data`),
			nil,
			nil,
			nil,
			nil,
		},
		{
			"Happy Path",
			[]string{""},
			[]byte(`test data`),
			[]byte(`test data`),
			nil,
			nil,
			nil,
			[]struct {
				flag  string
				value string
			}{
				{
					cst.ID,
					"140a372c-7d37-11eb-bc08-00155d19ad95",
				},
			},
		},
		{
			"error get permission",
			[]string{"path1"},
			[]byte(`test data`),
			nil,
			errors.New(e.New("error")),
			errors.New(e.New("error")),
			nil,
			nil,
		},
	}

	_, err := GetEditCmd()
	assert.Nil(t, err)

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			for _, f := range tt.flags {
				viper.Set(f.flag, f.value)
			}

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

			st := &fake.FakeStore{}

			st.GetStub = func(s string, d interface{}) *errors.ApiError {
				sData, ok := d.(*secretData)
				if ok {
					sData.Date = time.Now().Add(60 * time.Minute)
					sData.Data = tt.out
				}
				return tt.expectedErr
			}

			st.StoreStub = func(s string, i interface{}) *errors.ApiError {
				return tt.expectedErr
			}

			viper.Set(cst.StoreType, "")
			viper.Set(cst.CacheStrategy, "")

			s := &Secret{request: req, outClient: writer}
			s.edit = func(bytes2 []byte, d dataFunc, apiError *errors.ApiError, retry bool) (bytes []byte, apiError2 *errors.ApiError) {
				_, _ = d([]byte(`config`))
				return tt.editResponse, tt.editError
			}

			s.getStore = func(stString string) (i store.Store, apiError *errors.ApiError) {
				return st, nil
			}
			_ = s.handleEditCmd(tt.arg)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}

			for _, f := range tt.flags {
				viper.Set(f.flag, "")
			}
		})
	}
}

func TestGetSecretCmd(t *testing.T) {
	_, err := GetSecretCmd()
	assert.Nil(t, err)
}
