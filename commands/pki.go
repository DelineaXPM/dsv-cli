package cmd

import (
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	cst "thy/constants"
	apperrors "thy/errors"
	"thy/internal/predictor"
	"thy/internal/prompt"
	"thy/paths"
	"thy/store"
	"thy/utils"
	"thy/vaultcli"

	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

func GetPkiCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPki},
		SynopsisText: "pki (<action>)",
		HelpText:     "Work with certificates",
		RunFunc: func(args []string) int {
			return cli.RunResultHelp
		},
	})
}

func GetPkiRegisterCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPki, cst.Register},
		SynopsisText: "Register an existing root certificate authority",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s @cert.pem --%[4]s @key.pem --%[5]s myroot --%[6]s google.com,yahoo.com --%[7]s 1000
		`, cst.NounPki, cst.Register, cst.CertPath, cst.PrivKeyPath, cst.RootCAPath, cst.Domains, cst.MaxTTL),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.CertPath, Usage: "Path to a file containing the root certificate (required)"},
			{Name: cst.PrivKeyPath, Usage: "Path to a file containing the private key (required)"},
			{Name: cst.RootCAPath, Usage: "Path to a secret which will contain the registered root certificate with private key (required)"},
			{Name: cst.Domains, Usage: "List of domains for which certificates could be signed on behalf of the root CA (required)"},
			{Name: cst.MaxTTL, Usage: "Maximum number of hours for which a signed certificate on behalf of the root CA can be valid (required)"},
			{Name: cst.CRL, Usage: "URL of the CRL from which the revocation of leaf certificates can be checked"},
		},
		RunFunc: func(args []string) int {
			return handleRegisterRoot(vaultcli.New(), args)
		},
	})
}

func GetPkiSignCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPki, cst.Sign},
		SynopsisText: "Get a new certificate specified by a CSR and signed by a registered root CA",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myroot --%[4]s @csr.pem --%[5]s google.com,android.com --%[6]s 1000h
		`, cst.NounPki, cst.Sign, cst.RootCAPath, cst.CSRPath, cst.SubjectAltNames, cst.TTL),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.CSRPath, Usage: "Path to a file containing the CSR (required)"},
			{Name: cst.RootCAPath, Usage: "Path to a secret which contains the registered root certificate with private key (required)"},
			{Name: cst.SubjectAltNames, Usage: "List of subject alternative names (domains) for a certificate signed on behalf of the root CA can also be valid"},
			{Name: cst.TTL, Usage: "Number of hours for which a signed certificate on behalf of the root CA can be valid"},
			{Name: cst.Chain, Usage: "Include root certificate in response", ValueType: "bool"},
		},
		RunFunc: func(args []string) int {
			return handleSign(vaultcli.New(), args)
		},
	})
}

func GetPkiLeafCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPki, cst.Leaf},
		SynopsisText: "Get a new private key and leaf certificate signed by a registered root CA",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myroot --%[4]s myleafcert --%[5]s thycotic.com --%[6]s Thycotic --%[7]s US --%[8]s DC --%[9]s Washington --%[10]s 100d
   • %[1]s %[2]s --%[3]s myroot --%[5]s thycotic.com
		`, cst.NounPki, cst.Leaf, cst.RootCAPath, cst.PkiStorePath, cst.CommonName, cst.Organization, cst.Country, cst.State, cst.Locality, cst.TTL),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.CommonName, Usage: "Domain for which a certificate is generated (required)"},
			{Name: cst.Organization, Usage: ""},
			{Name: cst.Country, Usage: ""},
			{Name: cst.State, Usage: ""},
			{Name: cst.Locality, Usage: ""},
			{Name: cst.EmailAddress, Usage: ""},
			{Name: cst.Description, Usage: ""},

			{Name: cst.RootCAPath, Usage: "Path to a secret which contains the registered root certificate with private key (required)"},
			{Name: cst.PkiStorePath, Usage: "Path to a new secret in which to store the generated certificate with private key"},
			{Name: cst.TTL, Usage: "Number of hours for which a signed certificate on behalf of the root CA can be valid"},
			{Name: cst.Chain, Usage: "Include root certificate in response", ValueType: "bool"},
		},
		RunFunc: func(args []string) int {
			return handleLeaf(vaultcli.New(), args)
		},
	})
}

func GetPkiGenerateRootCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPki, cst.GenerateRoot},
		SynopsisText: "Generate and store a new root certificates with private key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myroot --%[4]s google.org,golang.org --%[5]s thycotic.com --%[6]s Thycotic --%[7]s US --%[8]s DC --%[9]s Washington --%[10]s 42d
   • %[1]s %[2]s --%[3]s myroot --%[4]s google.org,golang.org --%[5]s thycotic.com --%[10]s 52w
		`, cst.NounPki, cst.GenerateRoot, cst.RootCAPath, cst.Domains, cst.CommonName, cst.Organization, cst.Country, cst.State, cst.Locality, cst.MaxTTL),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.CommonName, Usage: "The domain name of the root CA (required)"},
			{Name: cst.Organization, Usage: ""},
			{Name: cst.Country, Usage: ""},
			{Name: cst.State, Usage: ""},
			{Name: cst.Locality, Usage: ""},
			{Name: cst.EmailAddress, Usage: ""},
			{Name: cst.Description, Usage: ""},

			{Name: cst.RootCAPath, Usage: "Path of a new secret in which to store the generated root certificate with private key (required)"},
			{Name: cst.Domains, Usage: "List of domains for which certificates could be signed on behalf of the root CA (required)"},
			{Name: cst.MaxTTL, Usage: "Number of hours for which a generated root certificate can be valid (required)"},
			{Name: cst.CRL, Usage: "URL of the CRL from which the revocation of leaf certificates can be checked"},
		},
		RunFunc: func(args []string) int {
			return handleGenerateRoot(vaultcli.New(), args)
		},
	})
}

func GetPkiSSHCertCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPki, cst.Sign},
		SynopsisText: "Get a new SSH certificate",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myroot --%[4]s myleaf --%[5]s root,ubuntu --%[6]s 1000
		`, cst.NounPki, cst.SSHCert, cst.RootCAPath, cst.LeafCAPath, cst.Principals, cst.TTL),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.RootCAPath, Usage: "Path to a secret which contains the registered root certificate with private key (required)"},
			{Name: cst.LeafCAPath, Usage: "Path to a secret which contains the leaf certificate with SSH-compatible public key (required)"},
			{Name: cst.Principals, Usage: "List of principals on the certificate (required)"},
			{Name: cst.TTL, Usage: "Number of hours for which a signed certificate can be valid (required)"},
		},
		MinNumberArgs: 8,
		RunFunc: func(args []string) int {
			return handleGetSSHCertificate(vaultcli.New(), args)
		},
	})
}

func handleRegisterRootWorkflow(vcli vaultcli.CLI, args []string) int {
	params := make(map[string]string)
	ui := &PasswordUi{
		cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
	}

	if certPath, err := prompt.Ask(ui, "Path to certificate file:"); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		data, err := store.ReadFile(certPath)
		if err != nil {
			ui.Error(err.Error())
			return utils.GetExecStatus(err)
		}
		params[cst.CertPath] = data
	}

	if privKeyPath, err := prompt.Ask(ui, "Path to private key file:"); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		data, err := store.ReadFile(privKeyPath)
		if err != nil {
			ui.Error(err.Error())
			return utils.GetExecStatus(err)
		}
		params[cst.PrivKeyPath] = data
	}

	if rootCAPath, err := prompt.Ask(ui, "Path to a new secret that will contain root CA registration information:"); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.RootCAPath] = rootCAPath
	}

	if domains, err := prompt.Ask(ui, "List of domains (comma-delimited strings):"); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.Domains] = domains
	}

	if maxTTL, err := prompt.Ask(ui, "Maximum TTL:"); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.MaxTTL] = maxTTL
	}

	if crl, err := prompt.AskDefault(ui, "Certificate Revocation List URL", ""); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.CRL] = crl
	}

	resp, err := submitRoot(vcli, params)
	vcli.Out().WriteResponse(resp, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func submitRoot(vcli vaultcli.CLI, params map[string]string) ([]byte, error) {
	paramErr := ValidateParams(params, []string{cst.RootCAPath, cst.PrivKeyPath, cst.CertPath, cst.Domains, cst.MaxTTL})
	if paramErr != nil {
		return nil, paramErr
	}

	_, err := parsePem(params[cst.CertPath])
	if err != nil {
		return nil, err
	}
	_, err = parsePem(params[cst.PrivKeyPath])
	if err != nil {
		return nil, err
	}

	maxTTL, err := utils.ParseHours(params[cst.MaxTTL])
	if err != nil {
		return nil, err
	}
	body := rootCASecret{
		RootCAPath:  params[cst.RootCAPath],
		PrivateKey:  base64Encode(params[cst.PrivKeyPath]),
		Certificate: base64Encode(params[cst.CertPath]),
		Domains:     utils.StringToSlice(params[cst.Domains]),
		MaxTTL:      maxTTL,
		CRL:         viper.GetString(cst.CRL),
	}

	basePath := strings.Join([]string{cst.NounPki, cst.Register}, "/")
	uri := paths.CreateURI(basePath, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
}

func handleRegisterRoot(vcli vaultcli.CLI, args []string) int {
	if OnlyGlobalArgs(args) {
		return handleRegisterRootWorkflow(vcli, args)
	}
	params := make(map[string]string)
	params[cst.CertPath] = viper.GetString(cst.CertPath)
	params[cst.PrivKeyPath] = viper.GetString(cst.PrivKeyPath)
	params[cst.RootCAPath] = viper.GetString(cst.RootCAPath)
	params[cst.Domains] = viper.GetString(cst.Domains)
	params[cst.MaxTTL] = viper.GetString(cst.MaxTTL)

	data, err := submitRoot(vcli, params)
	vcli.Out().WriteResponse(data, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func handleSignWorkflow(vcli vaultcli.CLI, args []string) int {
	params := make(map[string]string)
	ui := &PasswordUi{
		cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
	}

	if csrPath, err := prompt.Ask(ui, "Path to certificate signing request file:"); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		data, err := store.ReadFile(csrPath)
		if err != nil {
			ui.Error(err.Error())
			return utils.GetExecStatus(err)
		}
		params[cst.CSRPath] = data
	}

	if rootCAPath, err := prompt.Ask(ui, "Path of an existing secret that contains root CA information:"); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.RootCAPath] = rootCAPath
	}

	if subjectAltNames, err := prompt.AskDefault(ui, "List of subject alternative names (comma-delimited strings)", ""); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.SubjectAltNames] = subjectAltNames
	}

	if ttl, err := prompt.Ask(ui, "TTL:"); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.TTL] = ttl
	}

	yes, err := prompt.YesNo(ui, "Chain (optional - include root certificate)", false)
	if err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	}

	if yes {
		params[cst.Chain] = "true"
	} else {
		params[cst.Chain] = "false"
	}

	resp, err := submitSign(vcli, params)
	vcli.Out().WriteResponse(resp, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func handleSign(vcli vaultcli.CLI, args []string) int {
	if OnlyGlobalArgs(args) {
		return handleSignWorkflow(vcli, args)
	}

	params := make(map[string]string)
	params[cst.CSRPath] = viper.GetString(cst.CSRPath)
	params[cst.RootCAPath] = viper.GetString(cst.RootCAPath)
	params[cst.SubjectAltNames] = viper.GetString(cst.SubjectAltNames)
	params[cst.TTL] = viper.GetString(cst.TTL)
	params[cst.Chain] = viper.GetString(cst.Chain)

	data, err := submitSign(vcli, params)
	vcli.Out().WriteResponse(data, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func submitSign(vcli vaultcli.CLI, params map[string]string) ([]byte, error) {
	paramErr := ValidateParams(params, []string{cst.CSRPath, cst.RootCAPath})
	if paramErr != nil {
		return nil, paramErr
	}

	_, err := parsePem(params[cst.CSRPath])
	if err != nil {
		return nil, err
	}

	body := signingRequest{
		RootCAPath:      params[cst.RootCAPath],
		CSR:             base64Encode(params[cst.CSRPath]),
		SubjectAltNames: utils.StringToSlice(params[cst.SubjectAltNames]),
	}
	if params[cst.TTL] != "" {
		ttl, err := utils.ParseHours(params[cst.TTL])
		if err != nil {
			return nil, err
		}
		body.TTL = ttl
	}
	if c, err := strconv.ParseBool(params[cst.Chain]); err == nil {
		body.Chain = c
	}

	basePath := strings.Join([]string{cst.NounPki, cst.Sign}, "/")
	uri := paths.CreateURI(basePath, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
}

func handleLeafWorkflow(vcli vaultcli.CLI, args []string) int {
	params := make(map[string]string)
	ui := &PasswordUi{
		cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
	}

	if rootCAPath, err := prompt.Ask(ui, "Path of an existing secret that contains root CA information:"); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.RootCAPath] = rootCAPath
	}

	if commonName, err := prompt.Ask(ui, "Common name:"); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.CommonName] = commonName
	}

	if organization, err := prompt.AskDefault(ui, "Organization:", ""); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.Organization] = organization
	}

	if country, err := prompt.AskDefault(ui, "Country:", ""); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.Country] = country
	}

	if state, err := prompt.AskDefault(ui, "State:", ""); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.State] = state
	}

	if locality, err := prompt.AskDefault(ui, "Locality:", ""); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.Locality] = locality
	}

	if email, err := prompt.AskDefault(ui, "Email:", ""); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.EmailAddress] = email
	}

	if ttl, err := prompt.AskDefault(ui, "TTL:", ""); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.TTL] = ttl
	}

	if storePath, err := prompt.AskDefault(
		ui, "Path to a new secret in which to store the generated certificate with private key", ""); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.PkiStorePath] = storePath
	}

	yes, err := prompt.YesNo(ui, "Chain (optional - include root certificate)", false)
	if err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	}

	if yes {
		params[cst.Chain] = "true"
	} else {
		params[cst.Chain] = "false"
	}

	resp, err := submitLeaf(vcli, params)
	vcli.Out().WriteResponse(resp, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func submitLeaf(vcli vaultcli.CLI, params map[string]string) ([]byte, error) {
	paramErr := ValidateParams(params, []string{cst.RootCAPath, cst.CommonName})
	if paramErr != nil {
		return nil, paramErr
	}

	body := signingRequestInformation{
		RootCAPath: params[cst.RootCAPath],
		StorePath:  params[cst.PkiStorePath],
	}
	body.CommonName = params[cst.CommonName]
	body.Organization = params[cst.Organization]
	body.Country = params[cst.Country]
	body.State = params[cst.State]
	body.Locality = params[cst.Locality]
	body.EmailAddress = params[cst.EmailAddress]
	body.Description = params[cst.Description]
	if c, err := strconv.ParseBool(params[cst.Chain]); err == nil {
		body.Chain = c
	}
	if params[cst.TTL] != "" {
		ttl, err := utils.ParseHours(params[cst.TTL])
		if err != nil {
			return nil, err
		}
		body.TTL = ttl
	}

	basePath := strings.Join([]string{cst.NounPki, cst.Leaf}, "/")
	uri := paths.CreateURI(basePath, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
}

func handleLeaf(vcli vaultcli.CLI, args []string) int {
	if OnlyGlobalArgs(args) {
		return handleLeafWorkflow(vcli, args)
	}
	params := make(map[string]string)
	params[cst.RootCAPath] = viper.GetString(cst.RootCAPath)
	params[cst.PkiStorePath] = viper.GetString(cst.PkiStorePath)
	params[cst.TTL] = viper.GetString(cst.TTL)

	params[cst.CommonName] = viper.GetString(cst.CommonName)
	params[cst.Organization] = viper.GetString(cst.Organization)
	params[cst.Country] = viper.GetString(cst.Country)
	params[cst.State] = viper.GetString(cst.State)
	params[cst.Locality] = viper.GetString(cst.Locality)
	params[cst.EmailAddress] = viper.GetString(cst.EmailAddress)
	params[cst.Description] = viper.GetString(cst.Description)
	params[cst.Chain] = viper.GetString(cst.Chain)

	data, err := submitLeaf(vcli, params)
	vcli.Out().WriteResponse(data, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func handleGenerateRootWorkflow(vcli vaultcli.CLI, args []string) int {
	params := make(map[string]string)
	ui := &PasswordUi{
		cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
	}

	if rootCAPath, err := prompt.Ask(
		ui, "Path of a new secret in which to store the generated root certificate with private key:"); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.RootCAPath] = rootCAPath
	}

	if commonName, err := prompt.Ask(ui, "Common name:"); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.CommonName] = commonName
	}

	if organization, err := prompt.AskDefault(ui, "Organization:", ""); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.Organization] = organization
	}

	if country, err := prompt.AskDefault(ui, "Country:", ""); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.Country] = country
	}

	if state, err := prompt.AskDefault(ui, "State:", ""); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.State] = state
	}

	if locality, err := prompt.AskDefault(ui, "Locality:", ""); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.Locality] = locality
	}

	if email, err := prompt.AskDefault(ui, "Email:", ""); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.EmailAddress] = email
	}

	if domains, err := prompt.Ask(ui, "List of domains (comma-delimited strings):"); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.Domains] = domains
	}

	if maxTTL, err := prompt.Ask(ui, "Maximum TTL:"); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.MaxTTL] = maxTTL
	}

	if crl, err := prompt.AskDefault(ui, "Certificate Revocation List URL", ""); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.CRL] = crl
	}

	resp, err := submitGenerateRoot(vcli, params)
	vcli.Out().WriteResponse(resp, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func submitGenerateRoot(vcli vaultcli.CLI, params map[string]string) ([]byte, error) {
	paramErr := ValidateParams(params, []string{cst.RootCAPath, cst.CommonName, cst.Domains, cst.MaxTTL})
	if paramErr != nil {
		return nil, paramErr
	}

	body := generateRootInformation{
		RootCAPath: params[cst.RootCAPath],
		Domains:    utils.StringToSlice(params[cst.Domains]),
		CRL:        viper.GetString(cst.CRL),
	}
	body.CommonName = params[cst.CommonName]
	body.Organization = params[cst.Organization]
	body.Country = params[cst.Country]
	body.State = params[cst.State]
	body.Locality = params[cst.Locality]
	body.EmailAddress = params[cst.EmailAddress]
	body.Description = params[cst.Description]
	maxTTL, err := utils.ParseHours(params[cst.MaxTTL])
	if err != nil {
		return nil, err
	}
	body.MaxTTL = maxTTL

	basePath := strings.Join([]string{cst.NounPki, cst.Root}, "/")
	uri := paths.CreateURI(basePath, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
}

func handleGenerateRoot(vcli vaultcli.CLI, args []string) int {
	if OnlyGlobalArgs(args) {
		return handleGenerateRootWorkflow(vcli, args)
	}

	params := make(map[string]string)
	params[cst.RootCAPath] = viper.GetString(cst.RootCAPath)
	params[cst.Domains] = viper.GetString(cst.Domains)
	params[cst.MaxTTL] = viper.GetString(cst.MaxTTL)

	params[cst.CommonName] = viper.GetString(cst.CommonName)
	params[cst.Organization] = viper.GetString(cst.Organization)
	params[cst.Country] = viper.GetString(cst.Country)
	params[cst.State] = viper.GetString(cst.State)
	params[cst.Locality] = viper.GetString(cst.Locality)
	params[cst.EmailAddress] = viper.GetString(cst.EmailAddress)
	params[cst.Description] = viper.GetString(cst.Description)

	data, err := submitGenerateRoot(vcli, params)
	vcli.Out().WriteResponse(data, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func handleGetSSHCertificate(vcli vaultcli.CLI, args []string) int {
	params := make(map[string]string)
	params[cst.RootCAPath] = viper.GetString(cst.RootCAPath)
	params[cst.LeafCAPath] = viper.GetString(cst.LeafCAPath)
	params[cst.Principals] = viper.GetString(cst.Principals)
	params[cst.TTL] = viper.GetString(cst.TTL)

	paramErr := ValidateParams(params, []string{cst.RootCAPath, cst.LeafCAPath, cst.Principals, cst.TTL})
	if paramErr != nil {
		vcli.Out().Fail(paramErr)
		return utils.GetExecStatus(paramErr)
	}

	body := sshCertificateInformation{
		RootCAPath: params[cst.RootCAPath],
		LeafCAPAth: params[cst.LeafCAPath],
		Principals: utils.StringToSlice(params[cst.Principals]),
	}

	if ttl, err := utils.ParseHours(params[cst.TTL]); err != nil {
		vcli.Out().Fail(err)
		return utils.GetExecStatus(err)
	} else {
		body.TTL = ttl
	}

	basePath := strings.Join([]string{cst.NounPki, cst.SSHCert}, "/")
	uri := paths.CreateURI(basePath, nil)
	resp, apiError := vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func parsePem(data string) (*pem.Block, error) {
	block, _ := pem.Decode([]byte(data))
	if block == nil {
		return nil, errors.New("failed to decode the data into PEM format")
	}
	return block, nil
}

func base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

type rootCASecret struct {
	RootCAPath  string   `json:"rootCAPath"`
	PrivateKey  string   `json:"privateKey"`
	Certificate string   `json:"certificate"`
	Domains     []string `json:"domains"`
	MaxTTL      int      `json:"maxTTL"`
	CRL         string   `json:"crl"`
}

type signingRequest struct {
	RootCAPath      string   `json:"rootCAPath"`
	CSR             string   `json:"csr"`
	SubjectAltNames []string `json:"subjectAltNames"`
	TTL             int      `json:"ttl"`
	Chain           bool     `json:"chain"`
}

type signingRequestInformation struct {
	subjectInformation
	RootCAPath string `json:"rootCAPath"`
	StorePath  string `json:"storePath"`
	TTL        int    `json:"ttl"`
	Chain      bool   `json:"chain"`
}

type generateRootInformation struct {
	subjectInformation
	RootCAPath string   `json:"rootCAPath"`
	Domains    []string `json:"domains"`
	MaxTTL     int      `json:"maxTTL"`
	CRL        string   `json:"crl"`
}

type sshCertificateInformation struct {
	RootCAPath string   `json:"rootCAPath"`
	LeafCAPAth string   `json:"leafCAPath"`
	Principals []string `json:"principals"`
	TTL        int      `json:"ttl"`
}

type subjectInformation struct {
	Country            string `json:"country"`
	State              string `json:"state"`
	Locality           string `json:"locality"`
	Organization       string `json:"organization"`
	OrganizationalUnit string `json:"organizationalUnit"`
	CommonName         string `json:"commonName"`
	EmailAddress       string `json:"emailAddress"`
	Description        string `json:"description"`
}
