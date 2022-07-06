package cmd

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	cst "thy/constants"
	"thy/errors"
	"thy/tests/fake"
	"thy/vaultcli"
)

func TestGetGroupCmd(t *testing.T) {
	_, err := GetGroupCmd()
	assert.Nil(t, err)
}

func TestGetGroupReadCmd(t *testing.T) {
	_, err := GetGroupReadCmd()
	assert.Nil(t, err)
}

func TestGetGroupCreateCmd(t *testing.T) {
	_, err := GetGroupCreateCmd()
	assert.Nil(t, err)
}

func TestGetGroupDeleteCmd(t *testing.T) {
	_, err := GetGroupDeleteCmd()
	assert.Nil(t, err)
}

func TestGetGroupRestoreCmd(t *testing.T) {
	_, err := GetGroupRestoreCmd()
	assert.Nil(t, err)
}

func TestGetAddMembersCmd(t *testing.T) {
	_, err := GetAddMembersCmd()
	assert.Nil(t, err)
}

func TestGetDeleteMembersCmd(t *testing.T) {
	_, err := GetDeleteMembersCmd()
	assert.Nil(t, err)
}

func TestGetMemberGroupsCmd(t *testing.T) {
	_, err := GetMemberGroupsCmd()
	assert.Nil(t, err)
}

func TestGetGroupSearchCmd(t *testing.T) {
	_, err := GetGroupSearchCmd()
	assert.Nil(t, err)
}

func TestHandleGroupReadCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name:        "Happy path",
			args:        "group1",
			out:         []byte(`test`),
			expectedErr: nil,
		},
		{
			name:        "API Error",
			args:        "group1",
			out:         []byte(`test`),
			expectedErr: errors.NewS("error"),
		},
		{
			name:        "No group passed",
			args:        "",
			out:         []byte(`test`),
			expectedErr: errors.NewS("error: must specify " + cst.DataGroupName),
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

			exitCode := handleGroupReadCmd(vcli, []string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
				assert.Equal(t, 0, exitCode)
			} else {
				assert.Equal(t, tt.expectedErr, err)
				assert.NotEqual(t, 0, exitCode)
			}
		})
	}
}

func TestHandleGroupRestoreCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name:        "Happy path",
			args:        "group1",
			out:         []byte(`test`),
			expectedErr: nil,
		},
		{
			name:        "API Error",
			args:        "group1",
			out:         []byte(`test`),
			expectedErr: errors.NewS("error"),
		},
		{
			name:        "No group passed",
			args:        "",
			out:         []byte(`test`),
			expectedErr: errors.NewS("error: must specify " + cst.DataGroupName),
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

			exitCode := handleGroupRestoreCmd(vcli, []string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
				assert.Equal(t, 0, exitCode)
			} else {
				assert.Equal(t, tt.expectedErr, err)
				assert.NotEqual(t, 0, exitCode)
			}
		})
	}
}

func TestHandleGroupCreateCmd(t *testing.T) {
	testCase := []struct {
		name        string
		fGroupName  string // flag: --group-name
		fMembers    string // flag: --members
		fData       string // flag: --data
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name:       "Happy path (--group-name and --members)",
			fGroupName: "g1",
			fMembers:   "a,b,c",
			out:        []byte(`test`),
		},
		{
			name:  "Happy path (--data)",
			fData: `{"groupName":"g1","members":["a","b","c"]}`,
			out:   []byte(`test`),
		},
		{
			name:  "Happy path (--data)",
			fData: `{"groupName":"g1","memberNames":["a","b","c"]}`,
			out:   []byte(`test`),
		},
		{
			name:        "Missing group name",
			expectedErr: errors.NewS("error: must specify " + cst.DataGroupName),
		},
		{
			name:        "Missing group name in data",
			fData:       `{"memberNames":["a","b","c"]}`,
			expectedErr: errors.NewS("error: missing group name (\"groupName\") field in data"),
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
			viper.Set(cst.DataGroupName, tt.fGroupName)
			viper.Set(cst.Members, tt.fMembers)
			viper.Set(cst.Data, tt.fData)

			_ = handleGroupCreateCmd(vcli, []string{})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleAddMembersCmd(t *testing.T) {
	testCase := []struct {
		name        string
		fGroupName  string // flag: --group-name
		fMembers    string // flag: --members
		fData       string // flag: --data
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name:       "Happy path (--group-name and --members)",
			fGroupName: "g1",
			fMembers:   "user1,user2",
			out:        []byte(`test`),
		},
		{
			name:       "Happy path (--group-name and --data)",
			fGroupName: "g1",
			fData:      `{"memberNames":["user1","user2"]}`,
			out:        []byte(`test`),
		},
		{
			name:       "Happy path (--group-name and --data) 2",
			fGroupName: "g1",
			fData:      `{"members":["user1","user2"]}`,
			out:        []byte(`test`),
		},
		{
			name:        "Missing group name",
			expectedErr: errors.NewS("error: flag --group-name is required"),
		},
		{
			name:        "Missing members",
			fGroupName:  "g1",
			out:         []byte(`test`),
			expectedErr: errors.NewS("error: must specify members"),
		},
		{
			name:        "Missing members in data",
			fGroupName:  "g1",
			fData:       `{"something":"something"}`,
			expectedErr: errors.NewS("error: missing list of members to add"),
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
			if tt.name == "Happy path" {
				viper.Set(cst.DataGroupName, "fakegroup")
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.DataGroupName, tt.fGroupName)
			viper.Set(cst.Members, tt.fMembers)
			viper.Set(cst.Data, tt.fData)

			_ = handleAddMembersCmd(vcli, []string{})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleDeleteMembersCmd(t *testing.T) {
	testCase := []struct {
		name        string
		fGroupName  string // flag: --group-name
		fMembers    string // flag: --members
		fData       string // flag: --data
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name:       "Happy path (--group-name and --members)",
			fGroupName: "g1",
			fMembers:   "user1,user2",
			out:        []byte(`test`),
		},
		{
			name:       "Happy path (--group-name and --data)",
			fGroupName: "g1",
			fData:      `{"memberNames":["user1","user2"]}`,
			out:        []byte(`test`),
		},
		{
			name:       "Happy path (--group-name and --data) 2",
			fGroupName: "g1",
			fData:      `{"members":["user1","user2"]}`,
			out:        []byte(`test`),
		},
		{
			name:        "Missing group name",
			expectedErr: errors.NewS("error: flag --group-name is required"),
		},
		{
			name:        "Missing members",
			fGroupName:  "g1",
			out:         []byte(`test`),
			expectedErr: errors.NewS("error: must specify members"),
		},
		{
			name:        "Missing members in data",
			fGroupName:  "g1",
			fData:       `{"something":"something"}`,
			expectedErr: errors.NewS("error: missing list of members to delete"),
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
			viper.Set(cst.DataGroupName, tt.fGroupName)
			viper.Set(cst.Members, tt.fMembers)
			viper.Set(cst.Data, tt.fData)

			_ = handleDeleteMembersCmd(vcli, []string{})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleGroupDeleteCmd(t *testing.T) {
	testCase := []struct {
		name        string
		fGroupName  string // flag: --group-name
		args        []string
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name:        "Happy path",
			fGroupName:  "group1",
			out:         []byte(`test`),
			expectedErr: nil,
		},
		{
			name:        "Missing group name",
			expectedErr: errors.NewS("error: must specify " + cst.DataGroupName),
		},
		{
			name:        "API error",
			args:        []string{"group1"},
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
			viper.Set(cst.DataGroupName, tt.fGroupName)

			_ = handleGroupDeleteCmd(vcli, tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleMemberGroupCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name:        "Happy path",
			args:        "group1",
			out:         []byte(`test`),
			expectedErr: nil,
		},
		{
			name:        "api Error",
			args:        "group1",
			out:         []byte(`test`),
			expectedErr: errors.NewS("error"),
		},
		{
			name:        "No group passed",
			args:        "",
			out:         []byte(`test`),
			expectedErr: errors.NewS("error: must specify " + cst.DataUsername),
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

			_ = handleUsersGroupReadCmd(vcli, []string{tt.args})
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleGroupSearchCmd(t *testing.T) {
	testCase := []struct {
		name        string
		fQuery      string // flag: --query
		fLimit      string // flag: --limit
		fCursor     string // flag: --cursor
		args        []string
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name:    "Happy path",
			fQuery:  "g",
			fLimit:  "1",
			fCursor: "value",
			out:     []byte(`test`),
		},
		{
			name:    "Happy path",
			fLimit:  "1",
			fCursor: "value",
			args:    []string{"g"},
			out:     []byte(`test`),
		},
		{
			name:        "api Error",
			args:        []string{"g"},
			expectedErr: errors.NewS("error"),
		},
		{
			name:        "No Search query",
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
			viper.Set(cst.Query, tt.fQuery)
			viper.Set(cst.Limit, tt.fLimit)
			viper.Set(cst.Cursor, tt.fCursor)

			_ = handleGroupSearchCmd(vcli, tt.args)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}
