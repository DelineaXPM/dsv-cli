//go:build windows
// +build windows

package credhelpers

import (
	"bytes"
	"strings"

	winc "github.com/danieljoos/wincred"
)

// Wincred handles secrets using the Windows credential service.
type Wincred struct{}

// GetName gets friendly name of service
func (h Wincred) GetName() string {
	return "windows credential manager"
}

// Add adds new credentials to the windows credentials manager.
func (h Wincred) Add(creds *Credentials) error {
	creds.ServerURL = externalUrlToInternalUrl(creds.ServerURL)
	credsLabels := []byte(CredsLabel)
	g := winc.NewGenericCredential(creds.ServerURL)
	g.UserName = creds.Username
	g.CredentialBlob = []byte(creds.Secret)
	g.Persist = winc.PersistLocalMachine
	g.Attributes = []winc.CredentialAttribute{{Keyword: "label", Value: credsLabels}}

	return g.Write()
}

// Delete removes credentials from the windows credentials manager.
func (h Wincred) Delete(serverURL string) error {
	serverURL = externalUrlToInternalUrl(serverURL)
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
func (h Wincred) Get(serverURL string) (string, string, error) {
	serverURL = externalUrlToInternalUrl(serverURL)
	g, _ := winc.GetGenericCredential(serverURL)
	if g == nil {
		return "", "", NewErrCredentialsNotFound()
	}
	for _, attr := range g.Attributes {
		if strings.Compare(attr.Keyword, "label") == 0 &&
			bytes.Compare(attr.Value, []byte(CredsLabel)) == 0 {

			return g.UserName, string(g.CredentialBlob), nil
		}
	}
	return "", "", NewErrCredentialsNotFound()
}

// List returns the stored URLs and corresponding usernames for a given credentials label.
func (h Wincred) List(prefix string) (map[string]string, error) {
	prefix = externalUrlToInternalUrl(prefix)
	creds, err := winc.List()
	if err != nil {
		return nil, err
	}

	resp := make(map[string]string)
	for i := range creds {
		tName := creds[i].TargetName
		if !strings.HasPrefix(tName, prefix) {
			continue
		}
		tNameExternal := internalUrlToExternalUrl(tName)
		attrs := creds[i].Attributes
		for _, attr := range attrs {
			if strings.Compare(attr.Keyword, "label") == 0 &&
				bytes.Compare(attr.Value, []byte(CredsLabel)) == 0 {

				resp[tNameExternal] = creds[i].UserName
			}
		}

	}

	return resp, nil
}
