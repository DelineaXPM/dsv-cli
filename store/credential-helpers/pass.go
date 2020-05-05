// A `pass` based credential helper. Passwords are stored as arguments to pass
// of the form: "$PASS_FOLDER/base64-url(serverURL)/username". We base64-url
// encode the serverURL, because under the hood pass uses files and folders, so
// /s will get translated into additional folders.
package credhelpers

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	cst "thy/constants"
)

const PASS_FOLDER = cst.CmdRoot

// Pass handles secrets using Linux secret-service as a store.
type Pass struct{}

// GetName gets friendly name of service
func (h Pass) GetName() string {
	return "pass service"
}

// Ideally these would be stored as members of Pass, but since all of Pass's
// methods have value receivers, not pointer receivers, and changing that is
// backwards incompatible, we assume that all Pass instances share the same configuration

// initializationMutex is held while initializing so that only one 'pass'
// round-tripping is done to check pass is functioning.
var initializationMutex sync.Mutex
var passInitialized bool

// CheckInitialized checks whether the password helper can be used. It
// internally caches and so may be safely called multiple times with no impact
// on performance, though the first call may take longer.
func (p Pass) CheckInitialized() bool {
	return p.checkInitialized() == nil
}

func (p Pass) checkInitialized() error {
	initializationMutex.Lock()
	defer initializationMutex.Unlock()
	if passInitialized {
		return nil
	}
	// We just run a `pass ls`, if it fails then pass is not initialized.
	_, err := p.runPassHelper("", "ls")
	if err != nil {
		return fmt.Errorf("pass not initialized: %v", err)
	}
	passInitialized = true
	return nil
}

func (p Pass) runPass(stdinContent string, args ...string) (string, error) {
	if err := p.checkInitialized(); err != nil {
		return "", err
	}
	return p.runPassHelper(stdinContent, args...)
}

func getPathFromUrl(url string) string {
	url = externalUrlToInternalUrl(url)
	path := strings.Split(url, "-")

	// encode identifier as only part that could be unicode
	idIndex := len(path) - 1
	id := path[idIndex]
	idEncoded := base64.URLEncoding.EncodeToString([]byte(id))
	path[idIndex] = idEncoded
	assembled := strings.Join(path, "/")
	return assembled
}
func getUrlFromPath(p string) (string, error) {
	paths := strings.Split(p, "/")

	idIndex := len(paths) - 1
	idEncoded := paths[idIndex]
	idBytes, err := base64.URLEncoding.DecodeString(idEncoded)
	if err != nil {
		return "", err
	}
	id := string(idBytes)
	paths[idIndex] = id
	// drop first item to remove cst.StoreRoot
	assembled := strings.Join(paths, "-")
	assembled = internalUrlToExternalUrl(assembled)
	return assembled, nil
}

func (p Pass) runPassHelper(stdinContent string, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("pass", args...)
	cmd.Stdin = strings.NewReader(stdinContent)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, stderr.String())
	}

	// trim newlines; pass v1.7.1+ includes a newline at the end of `show` output
	return strings.TrimRight(stdout.String(), "\n\r"), nil
}

// Add adds new credentials to the keychain.
func (h Pass) Add(creds *Credentials) error {
	if creds == nil {
		return errors.New("missing credentials")
	} else if creds.ServerURL == "" {
		return errors.New("missing key")
	}

	path := getPathFromUrl(creds.ServerURL)

	_, err := h.runPass(creds.Secret, "insert", "-f", "-m", path)
	return err
}

// Delete removes credentials from the store.
func (h Pass) Delete(serverURL string) error {
	if serverURL == "" {
		return errors.New("missing server url")
	}

	path := getPathFromUrl(serverURL)
	_, err := h.runPass("", "rm", "-rf", path)
	return err
}

func getPassDir() string {
	passDir := "$HOME/.password-store"
	if envDir := os.Getenv("PASSWORD_STORE_DIR"); envDir != "" {
		passDir = envDir
	}
	return os.ExpandEnv(passDir)
}

// listPassDir lists all the contents of a directory in the password store.
// Pass uses fancy unicode to emit stuff to stdout, so rather than try
// and parse this, let's just look at the directory structure instead.
func listPassDir(recurse bool, includeDirs bool, relative bool, args ...string) (paths []string, err error) {
	passDir := getPassDir()
	dirPath := path.Join(append([]string{passDir}, args...)...)
	paths = []string{}
	if !recurse {
		contents, err := ioutil.ReadDir(dirPath)
		if err != nil {
			if os.IsNotExist(err) {
				return []string{}, nil
			}
			return nil, err
		} else {
			for _, f := range contents {
				paths = append(paths, path.Join(dirPath, f.Name()))
			}
		}
	} else {
		if err := filepath.Walk(dirPath, func(p string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			} else if !includeDirs && f.IsDir() {
				return nil
			}
			if relative {
				p = p[len(passDir):]
				if strings.HasPrefix(p, "/") {
					p = p[1:]
				}
			}
			paths = append(paths, p)
			return nil
		}); err != nil {
			if os.IsNotExist(err) {
				return paths, nil
			}
			return nil, err
		} else {
			return paths, err
		}
	}
	return paths, nil
}

// Get returns the username and secret to use for a given registry server URL.
func (h Pass) Get(url string) (string, string, error) {
	if url == "" {
		return "", "", errors.New("missing secret url")
	}

	p := getPathFromUrl(url)
	paths := strings.Split(p, "/")
	pWithoutIdentifier := strings.Join(paths[0:len(paths)-1], "/")
	urlSplit := strings.Split(url, "-")
	identifier := urlSplit[len(urlSplit)-1]

	// All checks for better error message
	if _, err := os.Stat(path.Join(getPassDir(), pWithoutIdentifier)); err != nil {
		if os.IsNotExist(err) {
			log.Println("os not found")
			return "", "", nil
		}
		log.Println("other error")
		return "", "", err
	}
	identifiers, err := listPassDir(false, true, false, pWithoutIdentifier)
	if err != nil {
		log.Println("some other error")
		return "", "", err
	}
	if len(identifiers) < 1 {
		log.Println("no identifiers")
		return "", "", fmt.Errorf("no identifiers for %s", url)
	}

	secret, err := h.runPass("", "show", p)
	return identifier, secret, err
}

// List returns the stored URLs and corresponding usernames for a given credentials label
func (h Pass) List(prefix string) (map[string]string, error) {
	prefix = cst.StoreRoot + "/" + prefix
	paths, err := listPassDir(true, false, true, prefix)
	if err != nil {
		return nil, err
	}

	resp := map[string]string{}

	for _, p := range paths {
		if !strings.HasSuffix(p, ".gpg") {
			//directory
			continue
		}
		p := strings.TrimSuffix(p, ".gpg")
		url, err := getUrlFromPath(p)
		if err != nil {
			continue
		}
		resp[url] = url
	}

	return resp, nil
}
