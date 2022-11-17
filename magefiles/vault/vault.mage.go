// Package vault provides a simple interface to the Vault API, allowing the secrets retrieval for running integration tests.
package vault

import (
	"github.com/DelineaXPM/dsv-sdk-go/v2/vault"
	env "github.com/caarlos0/env/v6"
)

// Vault is the mage namespace for tasks related to DelineaXPM vault.
type Vault mg.Namespace

type Config struct {
	// DSV SPECIFIC ENV VARIABLES.
	DomainEnv       string `env:"DSV_DOMAIN,notEmpty"`                 // DomainEnv is the tenant domain name (e.g. example.secretsvaultcloud.com).
	ClientIDEnv     string `env:"DSV_CLIENT_ID,notEmpty"`              // ClientIDEnv for client based authentication.
	TenantIDEnv     string `env:"DSV_TENANT_ID,notEmpty"`              // TenantIDEnv is the DSV Tenant name. This is just the `example` of `example.secretsvaultcloud.com`.
	TLDEnv          string `env:"DSV_TLD" envDefault:"com"`            // TLDEnv is the DSV top level domain. Default of `com`.
	ClientSecretEnv string `json:"-" env:"DSV_CLIENT_SECRET,notEmpty"` // ClientSecretEnv is the client secret token for authentication.
}

func ParseDSVConfig() (Config, error) {
	cfg := Config{}
	cfg.configureLogging()
	if err := env.Parse(&cfg, env.Options{
		// Prefix: "DSV_",.
	}); err != nil {
		pterm.Error.Printfln("env.Parse() %+v", err)
		return Config{}, fmt.Errorf("unable to parse env vars: %w", err)
	}
	pterm.Success.Println("parsed environment variables")
	return cfg, nil
}

func newClient(cfg *Config) (*Vault, error) {
	clientVault, err := vault.New(vault.Configuration{
		Credentials: vault.ClientCredential{
			ClientID:     cfg.ClientIDEnv,
			ClientSecret: cfg.ClientSecretEnv,
		},
		Tenant: cfg.TenantIDEnv,
		TLD:    os.Getenv("DSV_TLD"),
	})
	if err != nil {
		return &Vault{}, fmt.Errorf("unable to create vault client: %w", err)
	}
	return clientVault, nil
}
