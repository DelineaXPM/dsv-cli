//go:build windows
// +build windows

package credhelpers

import (
	"bytes"
	"errors"
	"strings"

	winc "github.com/danieljoos/wincred"
)

const (
	credsLabelKey   = "label"
	credsLabelValue = "Vault Token"
)

// Wincred handles secrets using the Windows credential service.
type Wincred struct{}

// GetName gets friendly name of service
func (Wincred) GetName() string { return "windows credential manager" }

// Add adds new credentials to the windows credentials manager.
func (Wincred) Add(creds *Credentials) error {
	creds.ServerURL = externalToInternal(creds.ServerURL)

	g := winc.NewGenericCredential(creds.ServerURL)
	g.UserName = creds.Username
	g.CredentialBlob = []byte(creds.Secret)
	g.Persist = winc.PersistLocalMachine
	g.Attributes = []winc.CredentialAttribute{
		{Keyword: credsLabelKey, Value: []byte(credsLabelValue)},
	}

	return g.Write()
}

// Delete removes credentials from the windows credentials manager.
func (Wincred) Delete(serverURL string) error {
	serverURL = externalToInternal(serverURL)
	g, err := winc.GetGenericCredential(serverURL)
	if g == nil {
		return nil
	}
	if err != nil {
		return err
	}
	return g.Delete()
}

// Get retrieves credentials from the windows credentials manager.
func (Wincred) Get(serverURL string) (string, error) {
	serverURL = externalToInternal(serverURL)
	g, _ := winc.GetGenericCredential(serverURL)
	if g != nil {
		for _, attr := range g.Attributes {
			if attr.Keyword == credsLabelKey && bytes.Compare(attr.Value, []byte(credsLabelValue)) == 0 {
				return string(g.CredentialBlob), nil
			}
		}
	}
	return "", errors.New("credentials not found in native keychain")
}

// List returns the stored URLs and corresponding usernames for a given credentials label.
func (Wincred) List(prefix string) ([]string, error) {
	prefix = externalToInternal(prefix)
	creds, err := winc.List()
	if err != nil {
		return nil, err
	}

	resp := []string{}
	for i := range creds {
		tName := creds[i].TargetName
		if !strings.HasPrefix(tName, prefix) {
			continue
		}
		tNameExternal := internalToExternal(tName)
		attrs := creds[i].Attributes
		for _, attr := range attrs {
			if attr.Keyword == credsLabelKey && bytes.Compare(attr.Value, []byte(credsLabelValue)) == 0 {
				resp = append(resp, tNameExternal)
			}
		}
	}
	return resp, nil
}
