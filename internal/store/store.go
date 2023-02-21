package store

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/utils"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ../../tests/fake/fake_store.go . Store

// Store interface provides common methods for storing data.
type Store interface {
	Store(key string, data any) error
	StoreString(key string, data string) error
	Get(key string, out any) error
	Delete(key string) error
	Wipe(prefix string) error
	List(prefix string) ([]string, error)
}

// Supported store types.
const (
	Unset     = ""
	None      = "none"
	PassLinux = "pass_linux"
	WinCred   = "wincred"
	File      = "file"
)

//nolint:gochecknoglobals // these vars are used to ensure a proper initialization.
var (
	store     Store
	storeType string
	once      sync.Once
)

func GetStore(st string) (Store, error) {
	if storeType == Unset {
		if err := ValidateStoreType(st); err != nil {
			return nil, err
		}
	} else if storeType != st {
		return nil, fmt.Errorf("store type cannot be changed during execution (old: %s, new: %s)", storeType, st)
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

func ValidateStoreType(storeType string) error {
	// TODO : support osxkeychain, secretservice
	switch storeType {
	case PassLinux:
		if utils.GetEnvProviderFunc().GetOs() != "linux" {
			return fmt.Errorf("'pass_linux' option for store.type is supported on linux only")
		}
		return nil
	case WinCred:
		if utils.GetEnvProviderFunc().GetOs() != "windows" {
			return fmt.Errorf("'wincred' option for store.type is supported on windows only")
		}
		return nil
	case File, None, Unset:
		return nil
	default:
		return fmt.Errorf(
			"'%s' key store not supported. Please choose from: ['pass_linux','wincred','file', and 'none']",
			storeType,
		)
	}
}

func StoreSecureSetting(key string, val string, sType string) error {
	if key == "" || val == "" {
		return fmt.Errorf("neither key nor value can be empty")
	}

	if sType != PassLinux && sType != WinCred {
		return fmt.Errorf("store.type is not secure store")
	}

	s, err := GetStore(sType)
	if err != nil {
		return err
	}
	keyFull := cst.CliConfigRoot + "-" + strings.ReplaceAll(key, ".", "-")
	return s.Store(keyFull, val)
}

func getSecureSettingForProfile(key string, profile string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key cannot be empty")
	}
	if val := viper.GetString(key); val != "" {
		return val, nil
	}

	if profile == "" {
		return "", fmt.Errorf("profile cannot be empty")
	}
	keyProfile := profile + "-" + key

	storeType := viper.GetString(cst.StoreType)
	s, err := GetStore(storeType)
	if err != nil {
		return "", fmt.Errorf("failed to fetch store: %w", err)
	}
	keyFull := cst.CliConfigRoot + "-" + strings.ReplaceAll(keyProfile, ".", "-")
	var res string
	err = s.Get(keyFull, &res)
	return res, err
}

func GetSecureSetting(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key cannot be empty")
	}
	profile := viper.GetString(cst.Profile)
	if profile == "" {
		profile = cst.DefaultProfile
	}
	s, err := getSecureSettingForProfile(key, profile)
	if err != nil {
		return "", err
	}
	return s, nil
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
