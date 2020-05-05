// +build windows

package store

import (
	ch "thy/store/credential-helpers"
)

func (f *StoreFactory) NewWinStore() Store {
	return f.NewSecureStore(&ch.Wincred{})
}
