//go:build windows
// +build windows

package store

import (
	ch "github.com/DelineaXPM/dsv-cli/internal/store/credential-helpers"
)

func NewWinStore() Store {
	return NewSecureStore(&ch.Wincred{})
}
