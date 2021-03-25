package paths

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	cst "thy/constants"
	"thy/errors"

	"github.com/spf13/viper"
)

type SecretType string

func GetInternalPathFromURIPath(uriPath string) string {
	path := strings.Replace(uriPath, "/", ":", -1)
	if strings.Index(path, ":") == 0 {
		path = path[1:]
	}
	if strings.Index(path, cst.PrefixEntityInternal) < 0 {
		path = cst.PrefixEntityInternal + path
	}
	return path
}

func GetURIPathFromInternalPath(internalPath string) string {
	path := strings.Replace(internalPath, ":", "/", -1)
	if strings.Index(path, "/") == 0 {
		path = path[1:]
	}
	if strings.Index(path, cst.PrefixEntity) == 0 {
		path = path[len(cst.PrefixEntity):]
	}
	return path
}

func GetResourceURIFromResourcePath(resourceType string, path string, id string, suffix string, trailingSlash bool, queryTerms map[string]string, pluralize bool) (string, *errors.ApiError) {
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
	requestURI := CreateResourceURI(resourceType, resourcePath, suffix, trailingSlash, queryTerms, pluralize)
	return requestURI, nil
}

func CreateResourceURI(resourceType string, path string, suffix string, trailingSlash bool, queryTerms map[string]string, pluralize bool) string {
	var completePath string
	plural := "s"
	if !pluralize {
		plural = ""
	}
	if trailingSlash {
		completePath = fmt.Sprintf("%s%s/%s%s", resourceType, plural, path, suffix)
	} else {
		completePath = fmt.Sprintf("%s%s%s%s", resourceType, plural, path, suffix)
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
	log.Printf("Request URI is %s\n", uri)
	return uri
}

func GetPath(args []string) string {
	path := viper.GetString(cst.Path)
	if len(args) > 0 {
		path = args[0]
	}
	return path
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

// GetFilenameFromArgs tries to extract a filename from args. If args has a --data or -d flag and
// its value starts with an '@' followed by a filename, the function tries to capture that filename.
func GetFilenameFromArgs(args []string) string {
	var fileName string
	for i := range args {
		if args[i] == "--data" || args[i] == "-d" {
			value := args[i+1]
			if strings.HasPrefix(value, "@") {
				fileName = value[1:]
			}
			break
		}
	}
	return fileName
}
