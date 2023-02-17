package vaultcli

import (
	"github.com/DelineaXPM/dsv-cli/auth"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/format"
	"github.com/DelineaXPM/dsv-cli/internal/store"
	"github.com/DelineaXPM/dsv-cli/requests"
)

type CLI interface {
	HTTPClient() requests.Client
	GraphQLClient() requests.GraphClient
	Out() format.OutClient
	Edit(data []byte, saveFunc SaveFunc) (edited []byte, err *errors.ApiError)
	Authenticator() auth.Authenticator
	Store(t string) (store.Store, error)
}

type vaultCLI struct {
	authenticator auth.Authenticator
	httpClient    requests.Client
	graphClient   requests.GraphClient
	outClient     format.OutClient
	store         store.Store
	editFunc      func(data []byte, save SaveFunc) (edited []byte, err *errors.ApiError)
}

func New() CLI {
	return &vaultCLI{}
}

func NewWithOpts(opts ...VaultCLIOption) (CLI, error) {
	vcli := &vaultCLI{}
	for _, op := range opts {
		if err := op(vcli); err != nil {
			return nil, err
		}
	}
	return vcli, nil
}

func (v *vaultCLI) Authenticator() auth.Authenticator {
	if v.authenticator != nil {
		return v.authenticator
	}
	return auth.NewAuthenticatorDefault()
}

func (v *vaultCLI) HTTPClient() requests.Client {
	if v.httpClient == nil {
		v.httpClient = requests.NewHttpClient()
	}
	return v.httpClient
}

func (v *vaultCLI) GraphQLClient() requests.GraphClient {
	if v.graphClient == nil {
		v.graphClient = requests.NewGraphClient()
	}
	return v.graphClient
}

func (v *vaultCLI) Out() format.OutClient {
	if v.outClient == nil {
		v.outClient = format.NewDefaultOutClient()
	}
	return v.outClient
}

func (v *vaultCLI) Store(t string) (store.Store, error) {
	if v.store != nil {
		return v.store, nil
	}
	s, err := store.GetStore(t)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (v *vaultCLI) Edit(data []byte, save SaveFunc) (edited []byte, err *errors.ApiError) {
	if v.editFunc != nil {
		return v.editFunc(data, save)
	}
	return EditData(data, save, nil, false)
}
