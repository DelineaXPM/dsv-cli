//go:build !windows
// +build !windows

package store

func NewWinStore() Store {
	panic("Windows Credential Manager is available only on Windows")
}
