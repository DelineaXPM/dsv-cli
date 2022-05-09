package store

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"thy/errors"
	ch "thy/store/credential-helpers"
)

type secureStore struct {
	internalStore ch.StoreHelper
}

func NewSecureStore(internalStore ch.StoreHelper) Store {
	return &secureStore{
		internalStore: internalStore,
	}
}

func (s *secureStore) Store(key string, secret interface{}) *errors.ApiError {
	if creds, err := dataToCreds(key, secret); err == nil {
		return errors.New(s.internalStore.Add(creds)).Grow("Failed to store secret in " + s.internalStore.GetName())
	} else {
		return errors.New(err).Grow("Failed to convert internal secret to credentials format")
	}
}

func (s *secureStore) StoreString(key string, data string) *errors.ApiError {
	if creds, err := dataToCreds(key, data); err == nil {
		return errors.New(s.internalStore.Add(creds)).Grow("Failed to store secret in " + s.internalStore.GetName())
	} else {
		return errors.New(err).Grow("Failed to convert internal secret to credentials format")
	}
}

func (s *secureStore) Get(key string, outSecret interface{}) *errors.ApiError {
	if _, secret, err := s.internalStore.Get(key); err == nil && secret != "" {
		return secretToData(secret, outSecret).Grow("Failed to convert credentials format to internal secret")
	} else if secret == "" || strings.Contains(err.Error(), "credentials not found in native keychain") {
		log.Printf("No entry found in secure storage for key '%s'\n", key)
		return nil
	} else {
		return errors.New(err).Grow("Failed to get secret from " + s.internalStore.GetName())
	}
}

func (s *secureStore) Delete(key string) *errors.ApiError {
	return errors.New(s.internalStore.Delete(key)).Grow("Failed to delete secret from " + s.internalStore.GetName())
}

func (s *secureStore) Wipe(prefix string) *errors.ApiError {
	if urls, err := s.internalStore.List(prefix); err == nil {
		for key := range urls {
			if err := s.internalStore.Delete(key); err != nil {
				return errors.New(err).Grow(fmt.Sprintf("Wipe interrupted; failed to delete credential with key '%s'", key))
			}
		}
	} else {
		return errors.New(err).Grow("Failed to enumerate credentials for deletion")
	}
	return nil
}

func (s *secureStore) List(prefix string) ([]string, *errors.ApiError) {
	if urls, err := s.internalStore.List(prefix); err != nil {
		return []string{}, errors.New(err).Grow("Failed to enumerate keys for deletion")
	} else {
		keys := make([]string, 0, len(urls))
		for k := range urls {
			keys = append(keys, k)
		}
		return keys, nil
	}
}

func dataToCreds(key string, data interface{}) (*ch.Credentials, *errors.ApiError) {
	marshalled, err := json.Marshal(data)
	if err != nil {
		return nil, errors.New(err)
	}
	return &ch.Credentials{
		ServerURL: key,
		Secret:    string(marshalled),
	}, nil
}

func secretToData(secret string, out interface{}) *errors.ApiError {
	return errors.New(json.Unmarshal([]byte(secret), out))
}
