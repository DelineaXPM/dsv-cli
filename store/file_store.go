package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"thy/errors"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/peterbourgon/diskv/v3"
)

const (
	defaultBasePath = "~/.thy/"
	defaultBaseDir  = ".thy"
)

type fileStore struct {
	internalStore *diskv.Diskv
}

func NewFileStore(basePath string) Store {
	if basePath == "" {
		if home, err := homedir.Dir(); err == nil {
			basePath = path.Join(home, defaultBaseDir)
		} else {
			basePath = defaultBasePath
		}
	}
	return &fileStore{
		internalStore: diskv.New(diskv.Options{
			BasePath:     basePath,
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
	keyChan := s.internalStore.Keys(nil)
	for k := range keyChan {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}
