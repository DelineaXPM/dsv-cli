//go:build windows
// +build windows

package store

import (
	ch "github.com/DelineaXPM/dsv-cli/store/credential-helpers"
)

func NewWinStore() Store {
	return NewSecureStore(&ch.Wincred{})
}
