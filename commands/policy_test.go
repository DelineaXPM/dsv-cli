package cmd

import (
	"net/http"
	"testing"

	cst "thy/constants"
	"thy/errors"
	"thy/tests/fake"
	"thy/vaultcli"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetPolicyCmd(t *testing.T) {
	_, err := GetPolicyCmd()
	assert.Nil(t, err)
}

func TestGetPolicyReadCmd(t *testing.T) {
	_, err := GetPolicyReadCmd()
	assert.Nil(t, err)
}

func TestGetPolicyEditCmd(t *testing.T) {
	_, err := GetPolicyEditCmd()
	assert.Nil(t, err)
}

func TestGetPolicyDeleteCmd(t *testing.T) {
	_, err := GetPolicyDeleteCmd()
	assert.Nil(t, err)
}

func TestGetPolicyRestoreCmd(t *testing.T) {
	_, err := GetPolicyRestoreCmd()
	assert.Nil(t, err)
}

func TestGetPolicyCreateCmd(t *testing.T) {
	_, err := GetPolicyCreateCmd()
	assert.Nil(t, err)
}

func TestGetPolicyUpdateCmd(t *testing.T) {
	_, err := GetPolicyUpdateCmd()
	assert.Nil(t, err)
}

func TestGetPolicyRollbackCmd(t *testing.T) {
	_, err := GetPolicyRollbackCmd()
	assert.Nil(t, err)
}

func TestGetPolicySearchCommand(t *testing.T) {
	_, err := GetPolicySearchCmd()
	assert.Nil(t, err)
}

func TestHandlePolicyReadCommand(t *testing.T) {
	testCase := []struct {
		name            string
		fPath           string // flag: --path
		fVersion        string // flag: --version
		args            []string
		apiOut          []byte
		apiErr          *errors.ApiError
		wantOut         []byte
		wantErr         *errors.ApiError
		wantNonZeroCode bool
	}{
		{
			name:    "Only path",
			fPath:   "secrets:databases:postgres58",
			apiOut:  []byte(`{"out":"val"}`),
			wantOut: []byte(`{"out":"val"}`),
		},
		{
			name:     "Path and version",
			fPath:    "secrets:databases:postgres58",
			fVersion: "3",
			apiOut:   []byte(`{"out":"val"}`),
			wantOut:  []byte(`{"out":"val"}`),
		},
		{
			name:    "Path from args",
			args:    []string{"secrets:databases:postgres58"},
			apiOut:  []byte(`{"out":"val"}`),
			wantOut: []byte(`{"out":"val"}`),
		},
		{
			name:            "Missing path",
			wantNonZeroCode: true,
		},
		{
			name:            "API error",
			args:            []string{"secrets:databases:postgres58"},
			apiErr:          errors.NewS("policy doesn't exist"),
			wantErr:         errors.NewS("policy doesn't exist"),
			wantNonZeroCode: true,
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
				return tt.apiOut, tt.apiErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.Path, tt.fPath)
			viper.Set(cst.Version, tt.fVersion)

			code := handlePolicyReadCmd(vcli, tt.args)
			if tt.wantErr == nil {
				assert.Equal(t, tt.wantOut, data)
			} else {
				assert.Equal(t, tt.wantErr, err)
			}
			if tt.wantNonZeroCode {
				assert.NotEqual(t, 0, code)
			}
		})
	}
}

func TestHandlePolicyEditCommand(t *testing.T) {
	testCase := []struct {
		name        string
		args        []string
		out         []byte
		expectedErr *errors.ApiError
		apiError    *errors.ApiError
	}{
		{
			name: "success",
			args: []string{"secrets:databases:postgres58"},
			out:  []byte(`test data`),
		},
		{
			name:        "error-missing-policy",
			args:        []string{"secrets:databases:postgres58"},
			expectedErr: errors.NewS("missing item at path"),
			apiError:    errors.NewS("missing item at path"),
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				if bytes != nil {
					data = bytes
				}
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(method string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				if method == http.MethodGet {
					return tt.out, tt.apiError
				}

				model := i.(*policyUpdateRequest)
				data = []byte(model.Policy)
				return nil, nil
			}

			editFunc := func(data []byte, save vaultcli.SaveFunc) (edited []byte, err *errors.ApiError) {
				return save(data)
			}
			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
				vaultcli.WithEditFunc(editFunc),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			_ = handlePolicyEditCmd(vcli, tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandlePolicyCreateCmd(t *testing.T) {
	testCase := []struct {
		name            string
		fData           string // flag: --data
		fPath           string // flag: --path
		fActions        string // flag: --actions
		fEffect         string // flag: --effect
		fDesc           string // flag: --desc
		fSubjects       string // flag: --subjects
		fResources      string // flag: --resources
		fCIDR           string // flag: --cidr
		args            []string
		apiOut          []byte
		apiErr          *errors.ApiError
		wantOut         []byte
		wantErr         *errors.ApiError
		wantNonZeroCode bool
	}{
		{
			name:            "Missing path",
			wantNonZeroCode: true,
		},
		{
			name:            "Missing actions",
			fPath:           "secrets:databases:postgres58",
			wantErr:         errors.NewS("--actions must be set"),
			wantNonZeroCode: true,
		},
		{
			name:            "Missing subjects",
			fPath:           "secrets:databases:postgres58",
			fActions:        "read,delete",
			wantErr:         errors.NewS("--subjects must be set"),
			wantNonZeroCode: true,
		},
		{
			name:      "All required params are set",
			fPath:     "secrets:databases:postgres58",
			fActions:  "read,delete",
			fSubjects: "groups:g44,groups:g32",
			apiOut:    []byte(`{"code":"success"`),
			wantOut:   []byte(`{"code":"success"`),
		},
		{
			name:      "Read path from args",
			fActions:  "read,delete",
			fSubjects: "groups:g44,groups:g32",
			args:      []string{"secrets:databases:postgres58"},
			apiOut:    []byte(`{"code":"success"`),
			wantOut:   []byte(`{"code":"success"`),
		},
		{
			name:       "All params are set",
			fPath:      "secrets:databases:postgres58",
			fActions:   "read,delete",
			fSubjects:  "groups:g44,groups:g32",
			fEffect:    "deny",
			fDesc:      "policy description",
			fResources: "secrets:databases:postgres58:<.*>",
			fCIDR:      "10.10.10.1/32",
			apiOut:     []byte(`{"code":"success"`),
			wantOut:    []byte(`{"code":"success"`),
		},
		{
			name:    "Use --data flag",
			fPath:   "secrets:databases:postgres58",
			fData:   `{"actions":["read", "delete"],"subjects":["groups:g44","groups:g32"]}`,
			apiOut:  []byte(`{"code":"success"`),
			wantOut: []byte(`{"code":"success"`),
		},
		{
			name:      "Invalid CIR",
			fPath:     "secrets:databases:postgres58",
			fActions:  "read,delete",
			fSubjects: "groups:g44,groups:g32",
			fCIDR:     "10.10.400.1/32",
			wantErr:   errors.NewS("invalid CIDR address: 10.10.400.1/32"),
		},
		{
			name:      "API error",
			fPath:     "secrets:databases:postgres58",
			fActions:  "read,delete",
			fSubjects: "groups:g44,groups:g32",
			apiErr:    errors.NewS(`{"code":"fail"`),
			wantErr:   errors.NewS(`{"code":"fail"`),
		},
		{
			name:      "Create path with supported specific characters",
			fActions:  "create",
			fSubjects: "groups:g44,groups:g32",
			args:      []string{"secrets:fol+der:fol-der/fold@r/fold:er/123/secret"},
			apiOut:    []byte(`{"code":"success"`),
			wantOut:   []byte(`{"code":"success"`),
		},
		{
			name:      "Create path with &",
			fActions:  "create",
			fSubjects: "groups:g44,groups:g32",
			args:      []string{"secrets:foler&:secret"},
			apiOut:    []byte(`{"code":"fail"`),
			wantOut:   []byte(`{"code":"fail"`),
			wantErr:   errors.NewS(`Path "secrets:foler&:secret" is invalid: path may contain only letters, numbers, underscores, dashes, @, pluses and periods separated by colon or slash`),
		},
		{
			name:      "Create path with $",
			fActions:  "create",
			fSubjects: "groups:g44,groups:g32",
			args:      []string{"secrets:foler$:secret"},
			apiOut:    []byte(`{"code":"fail"`),
			wantOut:   []byte(`{"code":"fail"`),
			wantErr:   errors.NewS(`Path "secrets:foler$:secret" is invalid: path may contain only letters, numbers, underscores, dashes, @, pluses and periods separated by colon or slash`),
		},
		{
			name:      "Create path with %",
			fActions:  "create",
			fSubjects: "groups:g44,groups:g32",
			args:      []string{"secrets:foler%:secret"},
			apiOut:    []byte(`{"code":"fail"`),
			wantOut:   []byte(`{"code":"fail"`),
			wantErr:   errors.NewS(`Path "secrets:foler%:secret" is invalid: path may contain only letters, numbers, underscores, dashes, @, pluses and periods separated by colon or slash`),
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
			outClient.FailEStub = func(apiError *errors.ApiError) {
				err = apiError
			}
			outClient.FailFStub = func(format string, args ...interface{}) { err = errors.NewF(format, args...) }

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.apiOut, tt.apiErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.Data, tt.fData)
			viper.Set(cst.Path, tt.fPath)
			viper.Set(cst.DataAction, tt.fActions)
			viper.Set(cst.DataEffect, tt.fEffect)
			viper.Set(cst.DataDescription, tt.fDesc)
			viper.Set(cst.DataSubject, tt.fSubjects)
			viper.Set(cst.DataResource, tt.fResources)
			viper.Set(cst.DataCidr, tt.fCIDR)

			code := handlePolicyCreateCmd(vcli, tt.args)

			t.Log(string(data))
			t.Log(err)
			if tt.wantErr == nil {
				assert.Equal(t, tt.wantOut, data)
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tt.wantErr, err)
			}

			if tt.wantNonZeroCode {
				assert.NotEqual(t, 0, code)
			}
		})
	}
}

func TestHandlePolicyUpdateCmd(t *testing.T) {
	testCase := []struct {
		name            string
		fData           string // flag: --data
		fPath           string // flag: --path
		fActions        string // flag: --actions
		fEffect         string // flag: --effect
		fDesc           string // flag: --desc
		fSubjects       string // flag: --subjects
		fResources      string // flag: --resources
		fCIDR           string // flag: --cidr
		args            []string
		apiOut          []byte
		apiErr          *errors.ApiError
		wantOut         []byte
		wantErr         *errors.ApiError
		wantNonZeroCode bool
	}{
		{
			name:            "Missing path",
			wantNonZeroCode: true,
		},
		{
			name:            "Missing actions",
			fPath:           "secrets:databases:postgres58",
			wantErr:         errors.NewS("--actions must be set"),
			wantNonZeroCode: true,
		},
		{
			name:            "Missing subjects",
			fPath:           "secrets:databases:postgres58",
			fActions:        "read,delete",
			wantErr:         errors.NewS("--subjects must be set"),
			wantNonZeroCode: true,
		},
		{
			name:      "All required params are set",
			fPath:     "secrets:databases:postgres58",
			fActions:  "read,delete",
			fSubjects: "groups:g44,groups:g32",
			apiOut:    []byte(`{"code":"success"`),
			wantOut:   []byte(`{"code":"success"`),
		},
		{
			name:      "Read path from args",
			fActions:  "read,delete",
			fSubjects: "groups:g44,groups:g32",
			args:      []string{"secrets:databases:postgres58"},
			apiOut:    []byte(`{"code":"success"`),
			wantOut:   []byte(`{"code":"success"`),
		},
		{
			name:       "All params are set",
			fPath:      "secrets:databases:postgres58",
			fActions:   "read,delete",
			fSubjects:  "groups:g44,groups:g32",
			fEffect:    "deny",
			fDesc:      "policy description",
			fResources: "secrets:databases:postgres58:<.*>",
			fCIDR:      "10.10.10.1/32",
			apiOut:     []byte(`{"code":"success"`),
			wantOut:    []byte(`{"code":"success"`),
		},
		{
			name:    "Use --data flag",
			fPath:   "secrets:databases:postgres58",
			fData:   `{"actions":["read", "delete"],"subjects":["groups:g44","groups:g32"]}`,
			apiOut:  []byte(`{"code":"success"`),
			wantOut: []byte(`{"code":"success"`),
		},
		{
			name:      "Invalid CIR",
			fPath:     "secrets:databases:postgres58",
			fActions:  "read,delete",
			fSubjects: "groups:g44,groups:g32",
			fCIDR:     "10.10.400.1/32",
			wantErr:   errors.NewS("invalid CIDR address: 10.10.400.1/32"),
		},
		{
			name:      "API error",
			fPath:     "secrets:databases:postgres58",
			fActions:  "read,delete",
			fSubjects: "groups:g44,groups:g32",
			apiErr:    errors.NewS(`{"code":"fail"`),
			wantErr:   errors.NewS(`{"code":"fail"`),
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
			outClient.FailEStub = func(apiError *errors.ApiError) {
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tt.apiOut, tt.apiErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.Data, tt.fData)
			viper.Set(cst.Path, tt.fPath)
			viper.Set(cst.DataAction, tt.fActions)
			viper.Set(cst.DataEffect, tt.fEffect)
			viper.Set(cst.DataDescription, tt.fDesc)
			viper.Set(cst.DataSubject, tt.fSubjects)
			viper.Set(cst.DataResource, tt.fResources)
			viper.Set(cst.DataCidr, tt.fCIDR)

			code := handlePolicyUpdateCmd(vcli, tt.args)

			t.Log(string(data))
			t.Log(err)
			if tt.wantErr == nil {
				assert.Equal(t, tt.wantOut, data)
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tt.wantErr, err)
			}

			if tt.wantNonZeroCode {
				assert.NotEqual(t, 0, code)
			}
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
			name:        "success",
			args:        []string{"secrets/servers/db"},
			apiResponse: []byte(`test`),
			out:         []byte(`test`),
		},
		{
			name:        "validation error",
			args:        []string{"secrets/servers/db/missing-polciy"},
			expectedErr: errors.NewS("item doesn't exist"),
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

			_ = handlePolicyDeleteCmd(vcli, tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandlePolicySearchCmd(t *testing.T) {
	testCase := []struct {
		name     string
		fQuery   string // flag --query
		fLimit   string // flag --limit
		fCursor  string // flag --cursor
		args     []string
		apiOut   []byte
		apiErr   *errors.ApiError
		wantOut  []byte
		wantErr  *errors.ApiError
		wantAddr string
	}{
		{
			name:     "Full search",
			apiOut:   []byte(`response`),
			wantOut:  []byte(`response`),
			wantAddr: "https://test.secretsvaultcloud.com/v1/config/policies",
		},
		{
			name:     "Search with query flag",
			fQuery:   "secrets",
			apiOut:   []byte(`response`),
			wantOut:  []byte(`response`),
			wantAddr: "https://test.secretsvaultcloud.com/v1/config/policies?searchTerm=secrets",
		},
		{
			name:    "Search with all flags",
			fQuery:  "secrets",
			fLimit:  "2",
			fCursor: "000000",
			apiOut:  []byte(`response`),
			wantOut: []byte(`response`),
		},
		{
			name:     "Search with query from args",
			args:     []string{"secrets"},
			apiOut:   []byte(`response`),
			wantOut:  []byte(`response`),
			wantAddr: "https://test.secretsvaultcloud.com/v1/config/policies?searchTerm=secrets",
		},
		{
			name:     "API error",
			apiErr:   errors.NewS("internal 500 error"),
			wantErr:  errors.NewS("internal 500 error"),
			wantAddr: "https://test.secretsvaultcloud.com/v1/config/policies",
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var address string
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				address = s2
				return tt.apiOut, tt.apiErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.Query, tt.fQuery)
			viper.Set(cst.Limit, tt.fLimit)
			viper.Set(cst.Cursor, tt.fCursor)
			viper.Set(cst.Tenant, "test")

			_ = handlePolicySearchCmd(vcli, tt.args)
			if tt.wantErr == nil {
				assert.Equal(t, tt.wantOut, data)
			} else {
				assert.Equal(t, tt.wantErr, err)
			}

			if tt.wantAddr != "" {
				assert.Equal(t, tt.wantAddr, address)
			}
		})
	}
}

func TestHandlePolicyRollbackCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        []string
		apiResponse []byte
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name:        "success (no version passed in)",
			args:        []string{"azure-dev"},
			apiResponse: []byte(`test`),
			out:         []byte(`{"version": "4"}`),
			expectedErr: nil,
		},
		{
			name:        "error (no version passed in)",
			args:        []string{"-q", "what happens when a server error occurs?"},
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

			_ = handlePolicyRollbackCmd(vcli, tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandlePolicyRestoreCmd(t *testing.T) {
	testCase := []struct {
		name            string
		args            string
		out             []byte
		expectedErr     *errors.ApiError
		wantNonZeroCode bool
	}{
		{
			name: "Success restore",
			args: "secrets:databases:postgres58",
			out:  []byte(`test`),
		},
		{
			name:            "API Error",
			args:            "secrets:databases:postgres58",
			expectedErr:     errors.NewS("error"),
			wantNonZeroCode: true,
		},
		{
			name:            "Missing path",
			expectedErr:     errors.NewS("error: must specify " + cst.DataGroupName),
			wantNonZeroCode: true,
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

			code := handlePolicyRestoreCmd(vcli, []string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}

			if tt.wantNonZeroCode {
				assert.NotEqual(t, 0, code)
			}
		})
	}
}
