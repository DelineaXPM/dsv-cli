package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

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
			FilePerm:     0o600,
			PathPerm:     0o700,
		}),
	}
}

func (s *fileStore) Store(key string, data any) error {
	marshaled, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.internalStore.Write(key, marshaled)
}

func (s *fileStore) StoreString(key string, data string) error {
	return s.internalStore.Write(key, []byte(data))
}

func (s *fileStore) Get(key string, out any) error {
	if !s.internalStore.Has(key) {
		return nil
	}
	b, err := s.internalStore.Read(key)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func (s *fileStore) Delete(key string) error {
	return s.internalStore.Erase(key)
}

func (s *fileStore) Wipe(prefix string) error {
	keys, err := s.List(prefix)
	if err != nil {
		return err
	}
	for _, k := range keys {
		err = s.internalStore.Erase(k)
		if err != nil {
			return fmt.Errorf("failed to delete key '%s' from disk: %w", k, err)
		}
	}

	if s.internalStore.TempDir != "" {
		err = os.RemoveAll(s.internalStore.TempDir)
		if err != nil {
			return fmt.Errorf("failed to wipe cache: %w", err)
		}
	}

	return nil
}

func (s *fileStore) List(prefix string) ([]string, error) {
	keys := []string{}
	keyChan := s.internalStore.Keys(nil)
	for k := range keyChan {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}
