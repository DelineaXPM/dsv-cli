package store

import "github.com/DelineaXPM/dsv-cli/errors"

type NoneStore struct{}

func (s *NoneStore) Store(key string, secret interface{}) *errors.ApiError  { return nil }
func (s *NoneStore) StoreString(key string, data string) *errors.ApiError   { return nil }
func (s *NoneStore) Get(key string, outSecret interface{}) *errors.ApiError { return nil }
func (s *NoneStore) Delete(key string) *errors.ApiError                     { return nil }
func (s *NoneStore) Wipe(prefix string) *errors.ApiError                    { return nil }
func (s *NoneStore) List(prefix string) ([]string, *errors.ApiError)        { return []string{}, nil }
