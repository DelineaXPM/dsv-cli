package cmd

import (
	"testing"
	"time"

	cst "thy/constants"
	"thy/errors"
	"thy/tests/fake"
	"thy/vaultcli"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetSecretCmd(t *testing.T) {
	_, err := GetSecretCmd()
	assert.Nil(t, err)
}

func TestGetSecretReadCmd(t *testing.T) {
	_, err := GetSecretReadCmd()
	assert.Nil(t, err)
}

func TestGetSecretDescribeCmd(t *testing.T) {
	_, err := GetSecretDescribeCmd()
	assert.Nil(t, err)
}

func TestGetSecretDeleteCmd(t *testing.T) {
	_, err := GetSecretDeleteCmd()
	assert.Nil(t, err)
}

func TestGetSecretRestoreCmd(t *testing.T) {
	_, err := GetSecretRestoreCmd()
	assert.Nil(t, err)
}

func TestGetSecretUpdateCmd(t *testing.T) {
	_, err := GetSecretUpdateCmd()
	assert.Nil(t, err)
}

func TestGetSecretRollbackCmd(t *testing.T) {
	_, err := GetSecretRollbackCmd()
	assert.Nil(t, err)
}

func TestGetSecretEditCmd(t *testing.T) {
	_, err := GetSecretEditCmd()
	assert.Nil(t, err)
}

func TestGetSecretCreateCmd(t *testing.T) {
	_, err := GetSecretCreateCmd()
	assert.Nil(t, err)
}

func TestGetSecretBustCacheCmd(t *testing.T) {
	_, err := GetSecretBustCacheCmd()
	assert.Nil(t, err)
}

func TestGetSecretSearchCmd(t *testing.T) {
	_, err := GetSecretSearchCmd()
	assert.Nil(t, err)
}

func TestHandleSecretDescribeCmd(t *testing.T) {
	testCase := []struct {
		name          string
		fID           string // flag: --id
		arg           []string
		cacheStrategy string
		out           []byte
		storeType     string
		expectedErr   *errors.ApiError
		apiError      *errors.ApiError
	}{
		{
			name:          "Happy Path no cacheStrategy",
			arg:           []string{"path1"},
			cacheStrategy: "",
			out:           []byte(`test data`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      nil,
		},
		{
			name:          "Happy Path no cacheStrategy",
			arg:           []string{"140a372c-7d37-11eb-bc08-00155d19ad95"},
			cacheStrategy: "",
			out:           []byte(`test data`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      nil,
		},
		{
			name:          "Happy Path no cacheStrategy",
			fID:           "140a372c-7d37-11eb-bc08-00155d19ad95",
			arg:           []string{""},
			cacheStrategy: "",
			out:           []byte(`test data`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      nil,
		},
		{
			name:          "Happy Path cache.server cacheStrategy",
			arg:           []string{"path1"},
			cacheStrategy: "cache.server",
			out:           []byte(`test data from cache`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      nil,
		},
		{
			name:          "Happy Path cache.server cacheStrategy",
			arg:           []string{"140a372c-7d37-11eb-bc08-00155d19ad95"},
			cacheStrategy: "cache.server",
			out:           []byte(`test data from cache`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      nil,
		},
		{
			name:          "Happy Path cache.server cacheStrategy",
			fID:           "140a372c-7d37-11eb-bc08-00155d19ad95",
			arg:           []string{""},
			cacheStrategy: "cache.server",
			out:           []byte(`test data from cache`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      nil,
		},
		{
			name:          "Happy Path server.cache cacheStrategy",
			arg:           []string{"path1"},
			cacheStrategy: "server.cache",
			out:           []byte(`test data from cache`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      errors.NewS("error"),
		},
		{
			name:          "Happy Path server.cache cacheStrategy",
			arg:           []string{"140a372c-7d37-11eb-bc08-00155d19ad95"},
			cacheStrategy: "server.cache",
			out:           []byte(`test data from cache`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      errors.NewS("error"),
		},
		{
			name:          "Happy Path server.cache cacheStrategy",
			fID:           "140a372c-7d37-11eb-bc08-00155d19ad95",
			arg:           []string{""},
			cacheStrategy: "server.cache",
			out:           []byte(`test data from cache`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      errors.NewS("error"),
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
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

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
				vaultcli.WithStore(st),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.StoreType, tt.storeType)
			viper.Set(cst.CacheStrategy, tt.cacheStrategy)
			viper.Set(cst.ID, tt.fID)

			_ = handleSecretDescribeCmd(vcli, cst.NounSecret, tt.arg)
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
			name:        "Happy path",
			args:        "user1",
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
			expectedErr: nil,
		},
		{
			name:        "api Error",
			args:        "user1",
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
			expectedErr: errors.NewS("error"),
		},
		{
			name:        "No Search query",
			args:        "",
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
			expectedErr: errors.NewS("error: must specify " + cst.Query),
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.out, tt.expectedErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()

			_ = handleSecretSearchCmd(vcli, cst.NounSecret, []string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})
	}
}

func TestHandleSecretDeleteCmd(t *testing.T) {
	testCase := []struct {
		name        string
		fID         string // flag: --id
		args        string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name:        "Happy path",
			args:        "user1",
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
			expectedErr: nil,
		},
		{
			name:        "Happy ID",
			args:        "140a372c-7d37-11eb-bc08-00155d19ad95",
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
			expectedErr: nil,
		},
		{
			name:        "Happy ID",
			fID:         "140a372c-7d37-11eb-bc08-00155d19ad95",
			args:        "",
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
			expectedErr: nil,
		},
		{
			name:        "api Error",
			args:        "user1",
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
			expectedErr: errors.NewS("error"),
		},
		{
			name:        "api Error",
			args:        "140a372c-7d37-11eb-bc08-00155d19ad95",
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
			expectedErr: errors.NewS("error"),
		},
		{
			name:        "api Error",
			fID:         "140a372c-7d37-11eb-bc08-00155d19ad95",
			args:        "",
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
			expectedErr: errors.NewS("error"),
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.out, tt.expectedErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.ID, tt.fID)

			_ = handleSecretDeleteCmd(vcli, cst.NounSecret, []string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})
	}
}

func TestHandleSecretRollbackCmd(t *testing.T) {
	testCase := []struct {
		name        string
		fID         string // flag: --id
		args        string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name:        "success (no version passed in) (ID)",
			args:        "140a372c-7d37-11eb-bc08-00155d19ad95",
			apiResponse: []byte(`test`),
			out:         []byte(`{"version": "4"}`),
			expectedErr: nil,
		},
		{
			name:        "success (no version passed in) (ID)",
			fID:         "140a372c-7d37-11eb-bc08-00155d19ad95",
			args:        "",
			apiResponse: []byte(`test`),
			out:         []byte(`{"version": "4"}`),
			expectedErr: nil,
		},
		{
			name:        "success (no version passed in) (path)",
			args:        "azure-dev",
			apiResponse: []byte(`test`),
			out:         []byte(`{"version": "4"}`),
			expectedErr: nil,
		},
		{
			name:        "error (no version passed in) (ID)",
			args:        "140a372c-7d37-11eb-bc08-00155d19ad95",
			apiResponse: []byte(`test`),
			out:         []byte(`{"someData": "hello"}`),
			expectedErr: errors.NewS("version not found"),
		},
		{
			name:        "error (no version passed in) (ID)",
			fID:         "140a372c-7d37-11eb-bc08-00155d19ad95",
			args:        "",
			apiResponse: []byte(`test`),
			out:         []byte(`{"someData": "hello"}`),
			expectedErr: errors.NewS("version not found"),
		},
		{
			name:        "error (no version passed in) (path)",
			args:        "azure-dev",
			apiResponse: []byte(`test`),
			out:         []byte(`{"someData": "hello"}`),
			expectedErr: errors.NewS("version not found"),
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.out, tt.expectedErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.ID, tt.fID)

			_ = handleSecretRollbackCmd(vcli, cst.NounSecret, []string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleSecretReadCmd(t *testing.T) {
	testCase := []struct {
		name          string
		fID           string // flag: --id
		arg           []string
		cacheStrategy string
		out           []byte
		storeType     string
		expectedErr   *errors.ApiError
		apiError      *errors.ApiError
	}{
		{
			name:          "No cache. ID from args.",
			arg:           []string{"140a372c-7d37-11eb-bc08-00155d19ad95"},
			cacheStrategy: "",
			out:           []byte(`test data`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      nil,
		},
		{
			name:          "No cache. ID from flag.",
			fID:           "140a372c-7d37-11eb-bc08-00155d19ad95",
			arg:           []string{""},
			cacheStrategy: "",
			out:           []byte(`test data`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      nil,
		},
		{
			name:          "No cache. Path from args.",
			arg:           []string{"path1"},
			cacheStrategy: "",
			out:           []byte(`test data`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      nil,
		},
		{
			name:          "Cache then server. Path from args.",
			arg:           []string{"path1"},
			cacheStrategy: "cache.server",
			out:           []byte(`test data from cache`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      nil,
		},
		{
			name:          "Cache then server. ID from flag.",
			fID:           "140a372c-7d37-11eb-bc08-00155d19ad95",
			arg:           []string{""},
			cacheStrategy: "cache.server",
			out:           []byte(`test data from cache`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      nil,
		},
		{
			name:          "Cache then server. ID from args.",
			arg:           []string{"140a372c-7d37-11eb-bc08-00155d19ad95"},
			cacheStrategy: "cache.server",
			out:           []byte(`test data from cache`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      nil,
		},
		{
			name:          "Server then cache. Path from args.",
			arg:           []string{"path1"},
			cacheStrategy: "server.cache",
			out:           []byte(`test data from cache`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      errors.NewS("error"),
		},
		{
			name:          "Server then cache. ID from flag.",
			fID:           "140a372c-7d37-11eb-bc08-00155d19ad95",
			arg:           []string{"path1"},
			cacheStrategy: "server.cache",
			out:           []byte(`test data from cache`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      errors.NewS("error"),
		},
		{
			name:          "Server then cache. ID from args.",
			arg:           []string{"140a372c-7d37-11eb-bc08-00155d19ad95"},
			cacheStrategy: "server.cache",
			out:           []byte(`test data from cache`),
			storeType:     "",
			expectedErr:   nil,
			apiError:      errors.NewS("error"),
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
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

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
				vaultcli.WithStore(st),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.Version, "v1")
			viper.Set(cst.StoreType, tt.storeType)
			viper.Set(cst.CacheStrategy, tt.cacheStrategy)
			viper.Set(cst.ID, tt.fID)

			_ = handleSecretReadCmd(vcli, cst.NounSecret, tt.arg)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})
	}
}

func TestHandleSecretUpsertCmd(t *testing.T) {
	testCases := []struct {
		name        string
		fID         string // flag: --id
		fDesc       string // flag: --desc
		args        []string
		out         []byte
		method      string
		expectedErr *errors.ApiError
	}{
		{
			name:        "Happy path POST",
			fDesc:       "new description",
			args:        []string{"mySecret", "--desc", "new description"},
			out:         []byte(`test`),
			method:      "create",
			expectedErr: nil,
		},
		{
			name:        "Happy path PUT (ID)",
			fDesc:       "new description",
			args:        []string{"140a372c-7d37-11eb-bc08-00155d19ad95", "--desc", "new description"},
			out:         []byte(`test`),
			method:      "update",
			expectedErr: nil,
		},
		{
			name:        "Happy path PUT (ID)",
			fID:         "140a372c-7d37-11eb-bc08-00155d19ad95",
			fDesc:       "new description",
			args:        []string{"--id", "140a372c-7d37-11eb-bc08-00155d19ad95", "--desc", "new description"},
			out:         []byte(`test`),
			method:      "update",
			expectedErr: nil,
		},
		{
			name:        "Happy path PUT (path)",
			fDesc:       "new description",
			args:        []string{"mySecret", "--description", "new description"},
			out:         []byte(`test`),
			method:      "update",
			expectedErr: nil,
		},
		{
			name:        "no path",
			args:        []string{"--description", "new description"},
			out:         []byte(`test`),
			method:      "",
			expectedErr: errors.NewS("error: must specify --id or --path (or [path])"),
		},
		{
			name:        "specific symbols in path are not supported",
			fDesc:       "new description",
			args:        []string{"secret$$", "--description", "new description"},
			out:         []byte(`test`),
			method:      "create",
			expectedErr: errors.NewS(`Path "secret$$" is invalid: path may contain only letters, numbers, underscores, dashes, @, pluses and periods separated by colon or slash`),
		},
		{
			name:        "specific symbols in path should be supported",
			fDesc:       "new description",
			args:        []string{"folder+:filder-:folder@/folder./folder_/secret", "--description", "new description"},
			out:         []byte(`test`),
			method:      "create",
			expectedErr: nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}
			outClient.FailEStub = func(apiError *errors.ApiError) { err = apiError }
			outClient.FailFStub = func(format string, args ...interface{}) { err = errors.NewF(format, args...) }

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.out, tt.expectedErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.ID, tt.fID)
			viper.Set(cst.DataDescription, tt.fDesc)

			_ = handleSecretUpsertCmd(vcli, cst.NounSecret, tt.method, tt.args)

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
			name:        "Happy Path",
			arg:         []string{},
			out:         nil,
			storeType:   "",
			expectedErr: nil,
		},
		{
			name:        "Error",
			arg:         []string{},
			out:         nil,
			storeType:   "",
			expectedErr: errors.NewS("error"),
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			st := &fake.FakeStore{}
			st.WipeStub = func(s string) *errors.ApiError {
				return tt.expectedErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithOutClient(outClient),
				vaultcli.WithStore(st),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", rerr)
			}

			viper.Reset()
			viper.Set(cst.StoreType, tt.storeType)

			_ = handleBustCacheCmd(vcli, tt.arg)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})
	}
}

func TestHandleSecretEditCmd(t *testing.T) {
	testCase := []struct {
		name         string
		fID          string // flag: --id
		arg          []string
		out          []byte
		editResponse []byte
		expectedErr  *errors.ApiError
		apiError     *errors.ApiError
		editError    *errors.ApiError
	}{
		{
			name:         "Happy Path",
			arg:          []string{"path1"},
			out:          []byte(`test data`),
			editResponse: []byte(`test data`),
			expectedErr:  nil,
			apiError:     nil,
			editError:    nil,
		},
		{
			name:         "Happy Path",
			arg:          []string{"140a372c-7d37-11eb-bc08-00155d19ad95"},
			out:          []byte(`test data`),
			editResponse: []byte(`test data`),
			expectedErr:  nil,
			apiError:     nil,
			editError:    nil,
		},
		{
			name:         "Happy Path",
			fID:          "140a372c-7d37-11eb-bc08-00155d19ad95",
			arg:          []string{""},
			out:          []byte(`test data`),
			editResponse: []byte(`test data`),
			expectedErr:  nil,
			apiError:     nil,
			editError:    nil,
		},
		{
			name:         "error get permission",
			arg:          []string{"path1"},
			out:          []byte(`test data`),
			editResponse: nil,
			expectedErr:  errors.NewS("error"),
			apiError:     errors.NewS("error"),
			editError:    nil,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
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

			editFunc := func(data []byte, save vaultcli.SaveFunc) (edited []byte, err *errors.ApiError) {
				_, _ = save([]byte(`config`))
				return tt.editResponse, tt.editError
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
				vaultcli.WithStore(st),
				vaultcli.WithEditFunc(editFunc),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", rerr)
			}

			viper.Reset()
			viper.Set(cst.ID, tt.fID)

			_ = handleSecretEditCmd(vcli, cst.NounSecret, tt.arg)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
		})
	}
}
