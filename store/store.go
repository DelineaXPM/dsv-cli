package store

import (
	"os"
	"path"
	"strings"
	"sync"

	cst "thy/constants"
	"thy/errors"
	"thy/utils"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ../tests/fake/fake_store.go . Store

// Store interface provides common methods for storing data.
type Store interface {
	Store(key string, secret interface{}) *errors.ApiError
	StoreString(key string, data string) *errors.ApiError
	Get(key string, outSecret interface{}) *errors.ApiError
	Delete(key string) *errors.ApiError
	Wipe(prefix string) *errors.ApiError
	List(prefix string) ([]string, *errors.ApiError)
}

// Supported store types.
const (
	Unset     = ""
	None      = "none"
	PassLinux = "pass_linux"
	WinCred   = "wincred"
	File      = "file"
)

var (
	store     Store
	storeType string
	once      sync.Once
)

func GetStore(st string) (Store, *errors.ApiError) {
	if storeType == Unset {
		if err := ValidateCredentialStore(st); err != nil {
			return nil, err
		}
	} else if storeType != st {
		return nil, errors.NewS("Store type cannot be changed during execution")
	}

	once.Do(func() {
		storeType = st
		switch storeType {
		case None:
			store = &NoneStore{}
		case PassLinux:
			store = NewPassStore()
		case WinCred:
			store = NewWinStore()
		default:
			store = NewFileStore(viper.GetString(cst.StorePath))
		}
	})

	return store, nil
}

func ValidateCredentialStore(st string) *errors.ApiError {
	// TODO : support osxkeychain, secretservice
	switch st {
	case "pass_linux":
		if utils.GetEnvProviderFunc().GetOs() != "linux" {
			return errors.NewS("'pass_linux' option for store.type is supported on linux only")
		}
		return nil
	case "wincred":
		if utils.GetEnvProviderFunc().GetOs() != "windows" {
			return errors.NewS("'wincred' option for store.type is supported on windows only")
		}
		return nil
	case "file", "none", "":
		return nil
	default:
		return errors.NewF("'%s' key store not supported. Please choose from: ['pass_linux','wincred','file', and 'none']", st)
	}
}

func StoreSecureSetting(key string, val string, sType string) *errors.ApiError {
	if key == "" || val == "" {
		return errors.NewS("neither key nor value can be empty")
	}

	if sType != PassLinux && sType != WinCred {
		return errors.NewS("store.type is not secure store")
	}

	s, err := GetStore(sType)
	if err != nil {
		return errors.New(err).Grow("failed to fetch store")
	}
	keyFull := cst.CliConfigRoot + "-" + strings.Replace(key, ".", "-", -1)
	return s.Store(keyFull, val)
}

func getSecureSettingForProfile(key string, profile string) (string, *errors.ApiError) {
	if key == "" {
		return "", errors.NewS("key cannot be empty")
	}
	if val := viper.GetString(key); val != "" {
		return val, nil
	}

	if profile == "" {
		return "", errors.NewS("profile cannot be empty")
	}
	keyProfile := profile + "-" + key

	storeType := viper.GetString(cst.StoreType)
	s, err := GetStore(storeType)
	if err != nil {
		return "", errors.New(err).Grow("failed to fetch store")
	}
	keyFull := cst.CliConfigRoot + "-" + strings.Replace(keyProfile, ".", "-", -1)
	var res string
	err = s.Get(keyFull, &res)
	return res, err
}

func GetSecureSetting(key string) (string, *errors.ApiError) {
	if key == "" {
		return "", errors.NewS("key cannot be empty")
	}
	profile := viper.GetString(cst.Profile)
	if profile == "" {
		profile = cst.DefaultProfile
	}
	return getSecureSettingForProfile(key, profile)
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
		storePath = defaultPath
	}
	bytes, err := os.ReadFile(path.Join(storePath, fileName))
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
