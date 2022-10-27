package paths

import (
	"fmt"
	"net/url"
	"strings"

	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/errors"

	"github.com/spf13/viper"
)

func GetURIPathFromInternalPath(internalPath string) string {
	path := strings.ReplaceAll(internalPath, ":", "/")
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimPrefix(path, cst.PrefixEntity)
	return path
}

func GetResourceURIFromResourcePath(resourceType string, path string, id string, suffix string, queryTerms map[string]string) (string, *errors.ApiError) {
	if id != "" && path != "" {
		return "", errors.NewS("error: only one of --id and --path (or [path]) may be set")
	}
	if path == "" && id == "" {
		return "", errors.NewS("error: must specify --id or --path (or [path])")
	}
	var resourcePath string
	if path != "" {
		resourcePath = GetURIPathFromInternalPath(path)
	}
	if id != "" {
		queryTerms["id"] = id
	}
	requestURI := CreateResourceURI(resourceType, resourcePath, suffix, true, queryTerms)
	return requestURI, nil
}

func CreateResourceURI(resourceType string, path string, suffix string, trailingSlash bool, queryTerms map[string]string) string {
	var completePath string
	if trailingSlash {
		completePath = fmt.Sprintf("%s/%s%s", resourceType, path, suffix)
	} else {
		completePath = fmt.Sprintf("%s%s%s", resourceType, path, suffix)
	}
	return CreateURI(completePath, queryTerms)
}

func CreateURI(path string, queryTerms map[string]string) string {
	httpScheme := cst.HTTPSchemeSecure
	if httpSchemeOverride := viper.GetString(cst.HTTPSchemeKey); httpSchemeOverride != "" {
		httpScheme = httpSchemeOverride
	}

	domain := GetDomain()
	port := GetPort()
	apiVersion := GetAPIVersion()

	uri := fmt.Sprintf("%s://%s.%s%s/%s/%s", httpScheme, viper.Get(cst.Tenant), domain, port, apiVersion, path)
	if queryTerms != nil {
		first := true
		for k, v := range queryTerms {
			if first {
				first = false
				uri = uri + "?"
			} else {
				uri = uri + "&"
			}
			val := url.QueryEscape(v)
			uri = uri + fmt.Sprintf("%s=%s", k, val)
		}
	}
	return uri
}

// ProcessResource converts a slash-delimited resource path into a colon-delimited resource path.
// The resource is any name, like user name, role name, group name, etc.
func ProcessResource(resource string) string {
	return strings.ReplaceAll(resource, "/", ":")
}

func GetDomain() string {
	domain := cst.Domain
	if domainOverride := viper.GetString(cst.DomainKey); domainOverride != "" {
		domain = domainOverride
	}
	return domain
}

func GetPort() string {
	port := ""
	if portOverride := viper.GetString(cst.PortKey); portOverride != "" {
		port = ":" + portOverride
	}
	return port
}

func GetAPIVersion() string {
	ver := cst.APIVersion
	if verOverride := viper.GetString(cst.APIVersionKey); verOverride != "" {
		ver = verOverride
	}
	return ver
}
