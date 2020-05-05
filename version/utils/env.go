package utils

import (
	"github.com/mitchellh/go-homedir"
	"runtime"
)

type EnvProvider interface {
	GetOs() string
	GetHomeDir() string
}
type envProvider struct{}

var GetEnvProviderFunc func() EnvProvider
var e EnvProvider

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
