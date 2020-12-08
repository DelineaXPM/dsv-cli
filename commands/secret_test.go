package cmd

import (
	e "errors"
	"testing"
	cst "thy/constants"
	"thy/errors"
	"thy/fake"
	"thy/store"
	"time"

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
	}{
		{
			"Happy Path no cacheStrategy",
			[]string{"path1"},
			"",
			[]byte(`test data`),
			"",
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
		},
		{
			"Happy Path server.cache cacheStrategy",
			[]string{"path1"},
			"server.cache",
			[]byte(`test data from cache`),
			"",
			nil,
			errors.New(e.New("error")),
		},
	}

	_, err := GetDescribeCmd()
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

	_, err := GetSecretSearchCmd()
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

			sec := &Secret{req, writer, store.GetStore, nil, cst.NounSecret}
			_ = sec.handleSecretSearchCmd([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
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
	}

	_, err := GetDeleteCmd()
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

			sec := &Secret{req, writer, store.GetStore, nil, cst.NounSecret}
			_ = sec.handleDeleteCmd([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
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
	}{
		{
			"success (no version passed in)",
			"azure-dev",
			[]byte(`test`),
			[]byte(`{"version": "4"}`),
			nil},
		{
			"error (no version passed in)",
			"azure-dev",
			[]byte(`test`),
			[]byte(`{"someData": "hello"}`),
			errors.NewS("version not found"),
		},
	}

	_, err := GetRollbackCmd()
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
				return tt.out, tt.expectedErr
			}

			sec := &Secret{req, writer, store.GetStore, nil, cst.NounSecret}
			_ = sec.handleRollbackCmd([]string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
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
	}{
		{
			"Happy Path no cacheStrategy",
			[]string{"path1"},
			"",
			[]byte(`test data`),
			"",
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
		},
		{
			"Happy Path server.cache cacheStrategy",
			[]string{"path1"},
			"server.cache",
			[]byte(`test data from cache`),
			"",
			nil,
			errors.New(e.New("error")),
		},
	}

	_, err := GetReadCmd()
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
	}{
		{
			"Happy path POST",
			[]string{"mySecret"},
			[]byte(`test`),
			"create",
			nil,
		},
		{
			"Happy path PUT",
			[]string{"mySecret"},
			[]byte(`test`),
			"update",
			nil,
		},
		{
			"no path",
			[]string{""},
			[]byte(`test`),
			"",
			errors.New(e.New("error: must specify --id or --path (or [path])")),
		},
	}

	_, err := GetUpdateCmd()
	assert.Nil(t, err)

	_, err = GetEditCmd()
	assert.Nil(t, err)

	_, err = GetCreateCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	viper.Set(cst.DataDescription, "new description")
	for _, tt := range testCases {
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

			sec := &Secret{req, writer, store.GetStore, nil, cst.NounSecret}
			viper.Set(cst.LastCommandKey, tt.method)

			_ = sec.handleUpsertCmd(tt.args)

			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
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
	}{
		{
			"Happy Path",
			[]string{},
			nil,
			"",
			nil,
		},
		{
			"Error",
			[]string{},
			nil,
			"",
			errors.New(e.New("error")),
		},
	}

	_, err := GetBustCacheCmd()
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
	}{
		{
			"Happy Path",
			[]string{"path1"},
			[]byte(`test data`),
			[]byte(`test data`),
			nil,
			nil,
			nil,
		},
		{
			"error get permission",
			[]string{"path1"},
			[]byte(`test data`),
			nil,
			errors.New(e.New("error")),
			errors.New(e.New("error")),
			nil,
		},
	}

	_, err := GetEditCmd()
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
		})
	}
}

func TestGetSecretCmd(t *testing.T) {
	_, err := GetSecretCmd()
	assert.Nil(t, err)
}
