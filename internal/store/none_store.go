package store

type NoneStore struct{}

func (s *NoneStore) Store(string, any) error          { return nil }
func (s *NoneStore) StoreString(string, string) error { return nil }
func (s *NoneStore) Get(string, any) error            { return nil }
func (s *NoneStore) Delete(string) error              { return nil }
func (s *NoneStore) Wipe(string) error                { return nil }
func (s *NoneStore) List(string) ([]string, error)    { return []string{}, nil }
