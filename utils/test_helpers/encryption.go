package test_helpers

import (
	"github.com/DelineaXPM/dsv-cli/auth"
	"github.com/DelineaXPM/dsv-cli/store"
)

// AddEncryptionKey creates a password encryption key in a user's home directory.
func AddEncryptionKey(tenant, username, password string) error {
	filename := auth.GetEncryptionKeyFilename(tenant, username)
	_, key, err := auth.StorePassword(filename, password)
	if err != nil {
		return err
	}

	st, apiError := store.GetStore(string(store.File))
	if apiError != nil {
		return err
	}
	apiError = st.StoreString(filename, key)
	if apiError != nil {
		return err
	}
	return nil
}
