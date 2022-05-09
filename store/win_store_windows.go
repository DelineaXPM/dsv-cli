//go:build windows
// +build windows

package store

import (
	ch "thy/store/credential-helpers"
)

func NewWinStore() Store {
	return NewSecureStore(&ch.Wincred{})
}
