package vaultcli

import (
	"thy/errors"

	"thy/auth"
	"thy/format"
	"thy/requests"
	"thy/store"
)

type VaultCLIOption func(vcli CLI) error

func WithAuthenticator(a auth.Authenticator) VaultCLIOption {
	return func(vcli CLI) error {
		internal, ok := vcli.(*vaultCLI)
		if !ok {
			return errors.NewS("unknown VaultCLI implementation")
		}
		internal.authenticator = a
		return nil
	}
}

func WithHTTPClient(c requests.Client) VaultCLIOption {
	return func(vcli CLI) error {
		internal, ok := vcli.(*vaultCLI)
		if !ok {
			return errors.NewS("unknown VaultCLI implementation")
		}
		internal.httpClient = c
		return nil
	}
}

func WithGraphQLClient(c requests.GraphClient) VaultCLIOption {
	return func(vcli CLI) error {
		internal, ok := vcli.(*vaultCLI)
		if !ok {
			return errors.NewS("unknown VaultCLI implementation")
		}
		internal.graphClient = c
		return nil
	}
}

func WithOutClient(c format.OutClient) VaultCLIOption {
	return func(vcli CLI) error {
		internal, ok := vcli.(*vaultCLI)
		if !ok {
			return errors.NewS("unknown VaultCLI implementation")
		}
		internal.outClient = c
		return nil
	}
}

func WithStore(s store.Store) VaultCLIOption {
	return func(vcli CLI) error {
		internal, ok := vcli.(*vaultCLI)
		if !ok {
			return errors.NewS("unknown VaultCLI implementation")
		}
		internal.store = s
		return nil
	}
}

func WithEditFunc(f func(data []byte, save SaveFunc) (edited []byte, err *errors.ApiError)) VaultCLIOption {
	return func(vcli CLI) error {
		internal, ok := vcli.(*vaultCLI)
		if !ok {
			return errors.NewS("unknown VaultCLI implementation")
		}
		internal.editFunc = f
		return nil
	}
}
