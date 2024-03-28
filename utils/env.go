package utils

import (
	"runtime"

	"github.com/mitchellh/go-homedir"
)

type EnvProvider interface {
	GetOs() string
	GetHomeDir() string
}
type envProvider struct{}

//nolint:gochecknoglobals // TODO: AB#561862 need to test this in a future PR
var (
	GetEnvProviderFunc func() EnvProvider
	e                  EnvProvider
)

func NewEnvProvider() EnvProvider {
	return &envProvider{}
}

func init() {
	e = NewEnvProvider()
	GetEnvProviderFunc = func() EnvProvider {
		return e
	}
}

func (e *envProvider) GetOs() string {
	return runtime.GOOS
}

type EnvFunc func() string

func (f EnvFunc) GetHomeDir() string {
	return f()
}

func (f EnvFunc) GetOs() string {
	return f()
}

func (e *envProvider) GetHomeDir() string {
	home, _ := homedir.Dir()
	if home == "" {
		home = "~"
	}
	return home
}
