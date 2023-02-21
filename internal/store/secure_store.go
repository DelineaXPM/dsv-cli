package store

import (
	"encoding/json"
	"fmt"
	"strings"

	ch "github.com/DelineaXPM/dsv-cli/internal/store/credential-helpers"
)

// storeHelper is the interface a credentials store helper must implement.
type storeHelper interface {
	Add(*ch.Credentials) error
	Delete(key string) error
	Get(key string) (string, error)
	List(prefix string) ([]string, error)
	GetName() string
}

type secureStore struct {
	internalStore storeHelper
}

func NewSecureStore(helper storeHelper) Store {
	return &secureStore{internalStore: helper}
}

func (s *secureStore) Store(key string, data any) error {
	marshaled, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("%s marshal '%s': %w", s.internalStore.GetName(), key, err)
	}

	err = s.internalStore.Add(&ch.Credentials{ServerURL: key, Secret: string(marshaled)})
	if err != nil {
		return fmt.Errorf("%s store '%s': %w", s.internalStore.GetName(), key, err)
	}
	return nil
}

func (s *secureStore) StoreString(key string, data string) error {
	return s.Store(key, data)
}

func (s *secureStore) Get(key string, out any) error {
	secret, err := s.internalStore.Get(key)
	if err != nil {
		if strings.Contains(err.Error(), "credentials not found") {
			return nil
		}
		return fmt.Errorf("%s get '%s': %w", s.internalStore.GetName(), key, err)
	}

	if secret == "" {
		return nil
	}

	err = json.Unmarshal([]byte(secret), out)
	if err != nil {
		return fmt.Errorf("%s unmarshal '%s': %w", s.internalStore.GetName(), key, err)
	}
	return nil
}

func (s *secureStore) Delete(key string) error {
	err := s.internalStore.Delete(key)
	if err != nil {
		return fmt.Errorf("%s delete '%s': %w", s.internalStore.GetName(), key, err)
	}
	return nil
}

func (s *secureStore) Wipe(prefix string) error {
	keys, err := s.internalStore.List(prefix)
	if err != nil {
		return fmt.Errorf("%s list prefix '%s': %w", s.internalStore.GetName(), prefix, err)
	}
	for _, key := range keys {
		err = s.internalStore.Delete(key)
		if err != nil {
			return fmt.Errorf("%s delete '%s': %w", s.internalStore.GetName(), key, err)
		}
	}
	return nil
}

func (s *secureStore) List(prefix string) ([]string, error) {
	keys, err := s.internalStore.List(prefix)
	if err != nil {
		return []string{}, fmt.Errorf("%s list prefix '%s': %w", s.internalStore.GetName(), prefix, err)
	}
	return keys, nil
}
