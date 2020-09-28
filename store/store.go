package store

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"reflect"
	"strings"
	"sync"
	cst "thy/constants"
	"thy/errors"
	ch "thy/store/credential-helpers"
	"thy/utils"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	store     Store
	storeType StoreType
	once      sync.Once
)

// StoreType is the type of authentication
type StoreType string

// Types of supported stores
const (
	Unset     = StoreType("")
	None      = StoreType("none")
	PassLinux = StoreType("pass_linux")
	WinCred   = StoreType("wincred")
	File      = StoreType("file")
)

// TODO : Move this out into main
func GetStore(stString string) (Store, *errors.ApiError) {
	if storeType == Unset {
		if err := ValidateCredentialStore(stString); err != nil {
			return nil, err
		}
	} else if storeType != StoreType(stString) {
		return nil, errors.NewS("Store type cannot be changed during execution")
	}
	once.Do(func() {
		st := StoreType(stString)
		storeType = st
		f := StoreFactory{}
		store = f.CreateStore(st)
	})

	return store, nil
}

func ValidateCredentialStore(st string) *errors.ApiError {
	switch st {
	// TODO : support osxkeychain, secretservice
	case "pass_linux", "wincred", "file", "none", "":
		osName := utils.GetEnvProviderFunc().GetOs()
		if st == "pass_linux" && osName != "linux" {
			return errors.NewS("'pass_linux' option for store.type is supported on linux only")
		} else if st == "wincred" && osName != "windows" {
			return errors.NewS("'wincred' option for store.type is supported on windows only")
		}
		return nil
	default:
		return errors.NewF("'%s' key store not supported. Please choose from: ['osxkeychain','pass_linux','secretservice','wincred','file', and 'none']", st)
	}
}

// Store interface provides common methods for storing data
type Store interface {
	Store(key string, secret interface{}) *errors.ApiError
	StoreString(key string, data string) *errors.ApiError
	Get(key string, outSecret interface{}) *errors.ApiError
	Delete(key string) *errors.ApiError
	Wipe(prefix string) *errors.ApiError
	List(prefix string) ([]string, *errors.ApiError)
}

type StoreFactory struct{}

func (f *StoreFactory) CreateStore(st StoreType) Store {
	switch st {
	case None:
		return &NoneStore{}
	case PassLinux:
		return f.NewPassStore()
	case WinCred:
		t := reflect.ValueOf(f)
		m := t.MethodByName("NewWinStore")
		sSlice := m.Call([]reflect.Value{})
		return sSlice[0].Interface().(Store)
	default:
		return f.NewFileStore()
	}
}

type NoneStore struct{}

func (s *NoneStore) Store(key string, secret interface{}) *errors.ApiError  { return nil }
func (s *NoneStore) StoreString(key string, data string) *errors.ApiError   { return nil }
func (s *NoneStore) Get(key string, outSecret interface{}) *errors.ApiError { return nil }
func (s *NoneStore) Delete(key string) *errors.ApiError                     { return nil }
func (s *NoneStore) Wipe(prefix string) *errors.ApiError                    { return nil }
func (s *NoneStore) List(prefix string) ([]string, *errors.ApiError)        { return []string{}, nil }

func StoreSecureSetting(key string, val string, st StoreType) *errors.ApiError {
	if key == "" || val == "" {
		return errors.NewS("neither key nor value can be empty")
	}
	isSecure := false
	if st == PassLinux || st == WinCred {
		isSecure = true
	}
	if !isSecure {
		return errors.NewS("store.type is not secure store")
	}

	s, err := GetStore(string(st))
	if err != nil {
		return errors.New(err).Grow("failed to fetch store")
	}
	keyFull := cst.CliConfigRoot + "-" + strings.Replace(key, ".", "-", -1)
	return s.Store(keyFull, val)
}

func dataToCreds(key string, data interface{}) (*ch.Credentials, *errors.ApiError) {
	if marshalled, err := json.Marshal(data); err == nil {
		return &ch.Credentials{
			ServerURL: key,
			Secret:    string(marshalled),
		}, nil
	} else {
		return nil, errors.New(err)
	}
}

func credsToData(creds *ch.Credentials, out interface{}) *errors.ApiError {
	return errors.New(json.Unmarshal([]byte(creds.Secret), out))
}

func secretToData(secret string, out interface{}) *errors.ApiError {
	return errors.New(json.Unmarshal([]byte(secret), out))
}

type Common interface {
	Delete(serverURL string) error
	Get(serverURL string) (string, string, error)
	List(prefix string) (map[string]string, error)
	Add(creds *ch.Credentials) error
	GetName() string
}

func ReadFile(fileName string) (string, error) {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GetDefaultPath retrieves an OS-independent default path for secrets, tokens, encrypted key files. By default, it is /home/{username}/.thy/.
func GetDefaultPath() (string, error) {
	if s, err := homedir.Dir(); err != nil {
		return "", err
	} else {
		return path.Join(s, ".thy"), nil
	}
}

// ReadFileInDefaultPath attempts to read a file in a store path. If the store path is not found,
// then the default thy directory is searched for a given file.
func ReadFileInDefaultPath(fileName string) (string, error) {
	storePath := viper.GetString(cst.StorePath)
	if storePath == "" {
		defaultPath, err := GetDefaultPath()
		if err != nil {
			return "", err
		}
		return ReadFile(path.Join(defaultPath, fileName))
	}
	return ReadFile(path.Join(storePath, fileName))
}
