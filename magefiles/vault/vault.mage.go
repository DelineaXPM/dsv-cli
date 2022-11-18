// Package vault provides a simple interface to the Vault API, allowing the secrets retrieval for running integration tests.
package vault

import (
	"fmt"

	"github.com/DelineaXPM/dsv-sdk-go/v2/vault"
	env "github.com/caarlos0/env/v6"
	"github.com/magefile/mage/mg"

	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"
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

// GetSecrets retrieves the desired secrets from the DelineaXPM vault.
func (Vault) GetSecrets() error {
	magetoolsutils.CheckPtermDebug()
	pterm.Info.Println("(Vault) GetSecrets()")
	cfg, err := ParseDSVConfig()
	if err != nil {
		return err
	}
	pterm.Debug.Printfln("Config: %+v", cfg)

	clientVault, err := newClient(cfg)
	if err != nil {
		pterm.Error.Printfln("newClient unable to create vault client: %+v", err)
		return err
	}

	secret1, err := clientVault.Secret("test-secret")
	if err != nil {
		pterm.Error.Printfln("unable to retrieve secret: %+v", err)
		return err
	}
	pterm.Debug.Printfln("secretkey: %v", secret1.Path)
	return nil
}

// parseDSVConfig parses the DelineaXPM vault configuration from the environment variables, and returns the configuration for usage in setting up client credentials.
func ParseDSVConfig() (*Config, error) {
	magetoolsutils.CheckPtermDebug()
	pterm.Info.Println("ParseDSVConfig()")

	cfg := Config{}
	if err := env.Parse(&cfg, env.Options{
		// Prefix: "DSV_",.
	}); err != nil {
		pterm.Error.Printfln("env.Parse() %+v", err)
		return &Config{}, fmt.Errorf("unable to parse env vars: %w", err)
	}
	pterm.Success.Println("parsed environment variables")
	return &cfg, nil
}

// newClient creates a new DelineaXPM vault client and returns for usage in retrieving and setting secrets.
func newClient(cfg *Config) (*vault.Vault, error) {
	magetoolsutils.CheckPtermDebug()
	pterm.Info.Println("newClient")
	clientVault, err := vault.New(vault.Configuration{
		Credentials: vault.ClientCredential{
			ClientID:     cfg.ClientIDEnv,
			ClientSecret: cfg.ClientSecretEnv,
		},
		Tenant: cfg.TenantIDEnv,
		TLD:    cfg.TLDEnv,
	})
	if err != nil {
		return &vault.Vault{}, fmt.Errorf("unable to create vault client: %w", err)
	}
	return clientVault, nil
}
