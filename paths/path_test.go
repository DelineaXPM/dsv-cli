package paths

import (
	"testing"

	cst "github.com/DelineaXPM/dsv-cli/constants"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetDomain(t *testing.T) {
	f := func(domain interface{}, expected string) {
		t.Helper()
		viper.Set(cst.DomainKey, domain)

		actual := GetDomain()
		assert.Equal(t, expected, actual)
	}
	f(nil, "secretsvaultcloud.com")
	f("test.com", "test.com")
	f(9192, "9192")
}

func TestGetAPIVersion(t *testing.T) {
	f := func(version interface{}, expected string) {
		t.Helper()
		viper.Set(cst.APIVersionKey, version)

		actual := GetAPIVersion()
		assert.Equal(t, expected, actual)
	}
	f(nil, "v1")
	f("v2", "v2")
	f(123, "123")
}

func TestGetPort(t *testing.T) {
	f := func(port interface{}, expected string) {
		t.Helper()
		viper.Set(cst.PortKey, port)

		actual := GetPort()
		assert.Equal(t, expected, actual)
	}
	f(nil, "")
	f("8089", ":8089")
	f(9192, ":9192")
}

func TestCreateURI(t *testing.T) {
	tests := []struct {
		name              string
		path              string
		queryTerms        map[string]string
		mockHTTPSchemeKey interface{}
		mockTenant        interface{}
		mockAPIVersionKey interface{}
		mockPortKey       interface{}
		mockDomainKey     interface{}
		expected          string
	}{
		{
			name:     "Default_Path",
			expected: "https://%!s(<nil>).secretsvaultcloud.com/v1/",
		},
		{
			name:              "HTTPSchemeKey",
			mockHTTPSchemeKey: "http",
			expected:          "http://%!s(<nil>).secretsvaultcloud.com/v1/",
		},
		{
			name:              "APIVersionKey",
			mockHTTPSchemeKey: "http",
			mockAPIVersionKey: "v2",
			expected:          "http://%!s(<nil>).secretsvaultcloud.com/v2/",
		},
		{
			name:              "PortKey",
			mockHTTPSchemeKey: "http",
			mockAPIVersionKey: "v2",
			mockPortKey:       "8080",
			expected:          "http://%!s(<nil>).secretsvaultcloud.com:8080/v2/",
		},
		{
			name:              "DomainKey",
			mockHTTPSchemeKey: "http",
			mockAPIVersionKey: "v2",
			mockPortKey:       "8080",
			mockDomainKey:     "secretsvaultcloud",
			expected:          "http://%!s(<nil>).secretsvaultcloud:8080/v2/",
		},
		{
			name:              "Tenant",
			mockHTTPSchemeKey: "http",
			mockAPIVersionKey: "v2",
			mockPortKey:       "8080",
			mockDomainKey:     "secretsvaultcloud.com",
			mockTenant:        "www",
			expected:          "http://www.secretsvaultcloud.com:8080/v2/",
		},
		{
			name:              "Path",
			path:              "path",
			mockHTTPSchemeKey: "http",
			mockAPIVersionKey: "v2",
			mockPortKey:       "8080",
			mockDomainKey:     "secretsvaultcloud.com",
			mockTenant:        "www",
			expected:          "http://www.secretsvaultcloud.com:8080/v2/path",
		},
		{
			name:              "QueryTerms",
			path:              "path",
			queryTerms:        map[string]string{},
			mockHTTPSchemeKey: "http",
			mockAPIVersionKey: "v2",
			mockPortKey:       "8080",
			mockDomainKey:     "secretsvaultcloud.com",
			mockTenant:        "www",
			expected:          "http://www.secretsvaultcloud.com:8080/v2/path",
		},
		{
			name:              "QueryTerms#01",
			path:              "path",
			queryTerms:        map[string]string{"query1": "query1"},
			mockHTTPSchemeKey: "http",
			mockAPIVersionKey: "v2",
			mockPortKey:       "8080",
			mockDomainKey:     "secretsvaultcloud.com",
			mockTenant:        "www",
			expected:          "http://www.secretsvaultcloud.com:8080/v2/path?query1=query1",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Set(cst.HTTPSchemeKey, test.mockHTTPSchemeKey)
			viper.Set(cst.Tenant, test.mockTenant)
			viper.Set(cst.APIVersionKey, test.mockAPIVersionKey)
			viper.Set(cst.PortKey, test.mockPortKey)
			viper.Set(cst.DomainKey, test.mockDomainKey)

			actual := CreateURI(test.path, test.queryTerms)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetURIPathFromInternalPath(t *testing.T) {
	tests := []struct {
		internalPath string
		expected     string
	}{
		{"", ""},
		{"a", "a"},
		{"a:a", "a/a"},
		{":a:a", "a/a"},
		{"/a:a", "a/a"},
		{"secrets:a:a", "a/a"},
		{":secrets:a:a", "a/a"},
		{"/secrets:a:a", "a/a"},
		{"/secrets/a:a", "a/a"},
	}
	for _, test := range tests {
		t.Run(test.internalPath, func(t *testing.T) {
			actual := GetURIPathFromInternalPath(test.internalPath)
			assert.Equal(t, test.expected, actual)
		})
	}
}
