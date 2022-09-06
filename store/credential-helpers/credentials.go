package credhelpers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Credentials holds the information shared between docker and the credentials store.
type Credentials struct {
	ServerURL string
	Username  string
	Secret    string
}

// isValid checks the integrity of Credentials object such that no credentials lack
// a server URL or a username.
// It returns whether the credentials are valid and the error if it isn't.
// error values can be errCredentialsMissingServerURL or errCredentialsMissingUsername
func (c *Credentials) isValid() (bool, error) {
	if len(c.ServerURL) == 0 {
		return false, NewErrCredentialsMissingServerURL()
	}

	if len(c.Username) == 0 {
		return false, NewErrCredentialsMissingUsername()
	}

	return true, nil
}

// CredsLabel holds the way Docker credentials should be labeled as such in credentials stores that allow labelling.
// That label allows to filter out non-Docker credentials too at lookup/search in macOS keychain,
// Windows credentials manager and Linux libsecret. Default value is "Vault Token"
var CredsLabel = "Vault Token"

// SetCredsLabel is a simple setter for CredsLabel
func SetCredsLabel(label string) {
	CredsLabel = label
}

// Store uses a helper and an input reader to save credentials.
// The reader must contain the JSON serialization of a Credentials struct.
func Store(helper StoreHelper, reader io.Reader) error {
	scanner := bufio.NewScanner(reader)

	buffer := new(bytes.Buffer)
	for scanner.Scan() {
		buffer.Write(scanner.Bytes())
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return err
	}

	var creds Credentials
	if err := json.NewDecoder(buffer).Decode(&creds); err != nil {
		return err
	}

	if ok, err := creds.isValid(); !ok {
		return err
	}

	return helper.Add(&creds)
}

// Get retrieves the credentials for a given server url.
// The reader must contain the server URL to search.
// The writer is used to write the JSON serialization of the credentials.
func Get(helper StoreHelper, reader io.Reader, writer io.Writer) error {
	scanner := bufio.NewScanner(reader)

	buffer := new(bytes.Buffer)
	for scanner.Scan() {
		buffer.Write(scanner.Bytes())
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return err
	}

	serverURL := strings.TrimSpace(buffer.String())
	if len(serverURL) == 0 {
		return NewErrCredentialsMissingServerURL()
	}

	username, secret, err := helper.Get(serverURL)
	if err != nil {
		return err
	}

	resp := Credentials{
		ServerURL: serverURL,
		Username:  username,
		Secret:    secret,
	}

	buffer.Reset()
	if err := json.NewEncoder(buffer).Encode(resp); err != nil {
		return err
	}

	fmt.Fprint(writer, buffer.String())
	return nil
}

// Erase removes credentials from the store.
// The reader must contain the server URL to remove.
func Erase(helper StoreHelper, reader io.Reader) error {
	scanner := bufio.NewScanner(reader)

	buffer := new(bytes.Buffer)
	for scanner.Scan() {
		buffer.Write(scanner.Bytes())
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return err
	}

	serverURL := strings.TrimSpace(buffer.String())
	if len(serverURL) == 0 {
		return NewErrCredentialsMissingServerURL()
	}

	return helper.Delete(serverURL)
}

// List returns all the serverURLs of keys in
// the OS store as a list of strings
func List(helper StoreHelper, writer io.Writer) error {
	accts, err := helper.List("")
	if err != nil {
		return err
	}
	return json.NewEncoder(writer).Encode(accts)
}
