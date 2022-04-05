package cmd

import (
	e "errors"
	"testing"
	"time"

	cst "thy/constants"
	"thy/errors"
	"thy/fake"
	"thy/vaultcli"

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

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(req),
				vaultcli.WithOutClient(writer),
				vaultcli.WithStore(st),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			_ = handleSecretDescribeCmd(vcli, cst.NounSecret, tt.arg)
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

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(req),
				vaultcli.WithOutClient(writer),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			_ = handleSecretSearchCmd(vcli, cst.NounSecret, []string{tt.args})
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

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(req),
				vaultcli.WithOutClient(writer),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			_ = handleSecretDeleteCmd(vcli, cst.NounSecret, []string{tt.args})
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

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(req),
				vaultcli.WithOutClient(writer),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			_ = handleSecretRollbackCmd(vcli, cst.NounSecret, []string{tt.args})
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

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(req),
				vaultcli.WithOutClient(writer),
				vaultcli.WithStore(st),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			_ = handleSecretReadCmd(vcli, cst.NounSecret, tt.arg)
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

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(req),
				vaultcli.WithOutClient(writer),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			_ = handleSecretUpsertCmd(vcli, cst.NounSecret, tt.args)

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

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithOutClient(writer),
				vaultcli.WithStore(st),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", rerr)
			}

			_ = handleBustCacheCmd(vcli, tt.arg)
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

			editFunc := func(data []byte, save vaultcli.SaveFunc) (edited []byte, err *errors.ApiError) {
				_, _ = save([]byte(`config`))
				return tt.editResponse, tt.editError
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(req),
				vaultcli.WithOutClient(writer),
				vaultcli.WithStore(st),
				vaultcli.WithEditFunc(editFunc),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", rerr)
			}

			_ = handleSecretEditCmd(vcli, cst.NounSecret, tt.arg)
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
