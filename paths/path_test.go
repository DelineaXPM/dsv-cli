package paths

import (
	"testing"

	cst "thy/constants"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetAPIVersion(t *testing.T) {
	tests := []struct {
		name           string
		mockAPIVersion interface{}
		expected       string
	}{
		{
			name:     "Default_Path",
			expected: "v1",
		},
		{
			name:           "Happy_Path",
			mockAPIVersion: "v2",
			expected:       "v2",
		},
		{
			name:           "Happy_Path#01",
			mockAPIVersion: 12313,
			expected:       "12313",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			viper.Set(cst.APIVersionKey, test.mockAPIVersion)

			actual := GetAPIVersion()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetPort(t *testing.T) {
	tests := []struct {
		name           string
		mockAPIVersion interface{}
		expected       string
	}{
		{
			name:     "Default_Path",
			expected: "",
		},
		{
			name:           "Happy_Path",
			mockAPIVersion: "8089",
			expected:       ":8089",
		},
		{
			name:           "Happy_Path#01",
			mockAPIVersion: 9192,
			expected:       ":9192",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			viper.Set(cst.PortKey, test.mockAPIVersion)

			actual := GetPort()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetDomain(t *testing.T) {
	tests := []struct {
		name           string
		mockAPIVersion interface{}
		expected       string
	}{
		{
			name:     "Default_Path",
			expected: "secretsvaultcloud.com",
		},
		{
			name:           "Happy_Path",
			mockAPIVersion: "test.com",
			expected:       "test.com",
		},
		{
			name:           "Happy_Path#01",
			mockAPIVersion: 9192,
			expected:       "9192",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			viper.Set(cst.DomainKey, test.mockAPIVersion)

			actual := GetDomain()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetPath(t *testing.T) {
	tests := []struct {
		name           string
		input          []string
		mockAPIVersion interface{}
		expected       string
	}{
		{
			name:     "Default_Path",
			expected: "",
		},
		{
			name:           "Happy_Path",
			mockAPIVersion: "test",
			expected:       "test",
		},
		{
			name:           "Happy_Path#01",
			input:          []string{"com"},
			mockAPIVersion: "test",
			expected:       "com",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			viper.Set(cst.Path, test.mockAPIVersion)

			actual := GetPath(test.input)
			assert.Equal(t, test.expected, actual)
		})
	}
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
