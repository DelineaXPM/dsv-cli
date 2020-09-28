package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	cst "thy/constants"
	"thy/errors"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/peterbourgon/diskv"
	"github.com/spf13/viper"
)

// FileStore is a file-backed store
type fileStore struct {
	internalStore *diskv.Diskv
}

const (
	basePath = "~/.thy/"
)

func (f *StoreFactory) NewFileStore() Store {
	p := basePath
	if viper.GetString(cst.StorePath) != "" {
		p = viper.GetString(cst.StorePath)
	} else if home, err := homedir.Dir(); err == nil {
		p = path.Join(home, ".thy")
	}
	return &fileStore{
		internalStore: diskv.New(diskv.Options{
			BasePath:     p,
			CacheSizeMax: 1024 * 1024,
			FilePerm:     0600,
			PathPerm:     0700,
		}),
	}
}

func (s *fileStore) Store(key string, secret interface{}) *errors.ApiError {
	if marshalled, err := json.Marshal(secret); err == nil {
		return errors.New(s.internalStore.Write(key, marshalled))
	} else {
		return errors.New(err)
	}
}

func (s *fileStore) StoreString(key string, data string) *errors.ApiError {
	return errors.New(s.internalStore.Write(key, []byte(data)))
}

func (s *fileStore) Get(key string, outSecret interface{}) *errors.ApiError {
	if !s.internalStore.Has(key) {
		return nil
	}
	if b, err := s.internalStore.Read(key); err == nil {
		return errors.New(json.Unmarshal(b, outSecret))
	} else {
		return errors.New(err)
	}
}

func (s *fileStore) Delete(key string) *errors.ApiError {
	return errors.New(s.internalStore.Erase(key))
}

func (s *fileStore) Wipe(prefix string) *errors.ApiError {
	if keys, err := s.List(prefix); err != nil {
		return err
	} else {
		for _, k := range keys {
			if e := s.internalStore.Erase(k); e != nil {
				return errors.New(e).Grow(fmt.Sprintf("Wipe interrupted. Failure to delete key '%s' from disk.", k))
			}
			if s.internalStore.TempDir != "" {
				return errors.New(os.RemoveAll(s.internalStore.TempDir)).Grow("Failed to wipe cache")
			}
		}
		return nil
	}
}

func (s *fileStore) List(prefix string) ([]string, *errors.ApiError) {
	keys := []string{}
	//cChan := make(chan struct{})
	keyChan := s.internalStore.Keys(nil)
	for k := range keyChan {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}
