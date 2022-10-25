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
	"thy/paths"
	"thy/utils"
	"thy/vaultcli"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

func GetPkiCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPki},
		SynopsisText: "Manage certificates",
		HelpText:     "Work with certificates",
		NoConfigRead: true,
		NoPreAuth:    true,
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
		RunFunc:    handleRegisterRootCmd,
		WizardFunc: handleRegisterRootWizard,
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
		RunFunc:    handleSignCmd,
		WizardFunc: handleSignWizard,
	})
}

func GetPkiLeafCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPki, cst.Leaf},
		SynopsisText: "Get a new private key and leaf certificate signed by a registered root CA",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myroot --%[4]s myleafcert --%[5]s delinea.com --%[6]s Delinea --%[7]s US --%[8]s DC --%[9]s Washington --%[10]s 100d
   • %[1]s %[2]s --%[3]s myroot --%[5]s delinea.com
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
		RunFunc:    handleLeafCmd,
		WizardFunc: handleLeafWizard,
	})
}

func GetPkiGenerateRootCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPki, cst.GenerateRoot},
		SynopsisText: "Generate and store a new root certificates with private key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myroot --%[4]s google.org,golang.org --%[5]s delinea.com --%[6]s Delinea --%[7]s US --%[8]s DC --%[9]s Washington --%[10]s 42d
   • %[1]s %[2]s --%[3]s myroot --%[4]s google.org,golang.org --%[5]s delinea.com --%[10]s 52w
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
		RunFunc:    handleGenerateRootCmd,
		WizardFunc: handleGenerateRootWizard,
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
		RunFunc:       handleGetSSHCertificateCmd,
	})
}

func handleRegisterRootCmd(vcli vaultcli.CLI, args []string) int {
	params := map[string]string{
		cst.CertPath:    viper.GetString(cst.CertPath),
		cst.PrivKeyPath: viper.GetString(cst.PrivKeyPath),
		cst.RootCAPath:  viper.GetString(cst.RootCAPath),
		cst.Domains:     viper.GetString(cst.Domains),
		cst.MaxTTL:      viper.GetString(cst.MaxTTL),
	}
	data, err := submitRoot(vcli, params)
	vcli.Out().WriteResponse(data, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func handleSignCmd(vcli vaultcli.CLI, args []string) int {
	params := map[string]string{
		cst.CSRPath:         viper.GetString(cst.CSRPath),
		cst.RootCAPath:      viper.GetString(cst.RootCAPath),
		cst.SubjectAltNames: viper.GetString(cst.SubjectAltNames),
		cst.TTL:             viper.GetString(cst.TTL),
		cst.Chain:           viper.GetString(cst.Chain),
	}
	data, err := submitSign(vcli, params)
	vcli.Out().WriteResponse(data, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func handleLeafCmd(vcli vaultcli.CLI, args []string) int {
	params := map[string]string{
		cst.RootCAPath:   viper.GetString(cst.RootCAPath),
		cst.PkiStorePath: viper.GetString(cst.PkiStorePath),
		cst.TTL:          viper.GetString(cst.TTL),
		cst.CommonName:   viper.GetString(cst.CommonName),
		cst.Organization: viper.GetString(cst.Organization),
		cst.Country:      viper.GetString(cst.Country),
		cst.State:        viper.GetString(cst.State),
		cst.Locality:     viper.GetString(cst.Locality),
		cst.EmailAddress: viper.GetString(cst.EmailAddress),
		cst.Description:  viper.GetString(cst.Description),
		cst.Chain:        viper.GetString(cst.Chain),
	}
	data, err := submitLeaf(vcli, params)
	vcli.Out().WriteResponse(data, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func handleGenerateRootCmd(vcli vaultcli.CLI, args []string) int {
	params := map[string]string{
		cst.RootCAPath:   viper.GetString(cst.RootCAPath),
		cst.Domains:      viper.GetString(cst.Domains),
		cst.MaxTTL:       viper.GetString(cst.MaxTTL),
		cst.CommonName:   viper.GetString(cst.CommonName),
		cst.Organization: viper.GetString(cst.Organization),
		cst.Country:      viper.GetString(cst.Country),
		cst.State:        viper.GetString(cst.State),
		cst.Locality:     viper.GetString(cst.Locality),
		cst.EmailAddress: viper.GetString(cst.EmailAddress),
		cst.Description:  viper.GetString(cst.Description),
	}
	data, err := submitGenerateRoot(vcli, params)
	vcli.Out().WriteResponse(data, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func handleGetSSHCertificateCmd(vcli vaultcli.CLI, args []string) int {
	params := map[string]string{
		cst.RootCAPath: viper.GetString(cst.RootCAPath),
		cst.LeafCAPath: viper.GetString(cst.LeafCAPath),
		cst.Principals: viper.GetString(cst.Principals),
		cst.TTL:        viper.GetString(cst.TTL),
	}
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

	resp, apiErr := pkiSSHCert(vcli, &body)
	vcli.Out().WriteResponse(resp, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleRegisterRootWizard(vcli vaultcli.CLI) int {
	qs := []*survey.Question{
		{
			Name:      "CertFile",
			Prompt:    &survey.Input{Message: "Path to certificate file:"},
			Validate:  vaultcli.SurveyRequiredFile,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "PrivKeyFile",
			Prompt:    &survey.Input{Message: "Path to private key file:"},
			Validate:  vaultcli.SurveyRequiredFile,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "SecretPath",
			Prompt:    &survey.Input{Message: "Path to a new secret that will contain root CA registration information:"},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "Domains",
			Prompt:    &survey.Input{Message: "List of domains (comma-delimited strings):"},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "MaxTTL",
			Prompt:    &survey.Input{Message: "Maximum TTL:"},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "CRL",
			Prompt:    &survey.Input{Message: "Certificate Revocation List URL:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
	}

	answers := struct {
		CertFile    string
		PrivKeyFile string
		SecretPath  string
		Domains     string
		MaxTTL      string
		CRL         string
	}{}
	survErr := survey.Ask(qs, &answers)
	if survErr != nil {
		vcli.Out().Fail(survErr)
		return utils.GetExecStatus(survErr)
	}

	params := map[string]string{
		cst.RootCAPath: answers.SecretPath,
		cst.Domains:    answers.Domains,
		cst.MaxTTL:     answers.MaxTTL,
		cst.CRL:        answers.CRL,
	}

	data, err := os.ReadFile(answers.CertFile)
	if err != nil {
		vcli.Out().Fail(err)
		return utils.GetExecStatus(err)
	}
	params[cst.CertPath] = string(data)

	data, err = os.ReadFile(answers.PrivKeyFile)
	if err != nil {
		vcli.Out().Fail(err)
		return utils.GetExecStatus(err)
	}
	params[cst.PrivKeyPath] = string(data)

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
	return pkiRegister(vcli, &body)
}

func handleSignWizard(vcli vaultcli.CLI) int {
	qs := []*survey.Question{
		{
			Name:      "CSRPath",
			Prompt:    &survey.Input{Message: "Path to certificate signing request file:"},
			Validate:  vaultcli.SurveyRequiredFile,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "SecretPath",
			Prompt:    &survey.Input{Message: "Path of an existing secret that contains root CA information:"},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "SubjectAltNames",
			Prompt:    &survey.Input{Message: "List of subject alternative names (comma-delimited strings):"},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "TTL",
			Prompt:    &survey.Input{Message: "TTL:"},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:   "Chain",
			Prompt: &survey.Confirm{Message: "Chain (optional - include root certificate):", Default: false},
		},
	}

	answers := struct {
		CSRPath         string
		SecretPath      string
		SubjectAltNames string
		TTL             string
		Chain           bool
	}{}
	survErr := survey.Ask(qs, &answers)
	if survErr != nil {
		vcli.Out().Fail(survErr)
		return utils.GetExecStatus(survErr)
	}

	params := map[string]string{
		cst.RootCAPath:      answers.SecretPath,
		cst.SubjectAltNames: answers.SubjectAltNames,
		cst.TTL:             answers.TTL,
		cst.Chain:           strconv.FormatBool(answers.Chain),
	}

	data, err := os.ReadFile(answers.CSRPath)
	if err != nil {
		vcli.Out().Fail(err)
		return utils.GetExecStatus(err)
	}
	params[cst.CSRPath] = string(data)

	resp, err := submitSign(vcli, params)
	vcli.Out().WriteResponse(resp, apperrors.New(err))
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
	return pkiSign(vcli, &body)
}

func handleLeafWizard(vcli vaultcli.CLI) int {
	qs := []*survey.Question{
		{
			Name:      "RootCAPath",
			Prompt:    &survey.Input{Message: "Path of an existing secret that contains root CA information:"},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "CommonName",
			Prompt:    &survey.Input{Message: "Common name:"},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "Organization",
			Prompt:    &survey.Input{Message: "Organization:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "Country",
			Prompt:    &survey.Input{Message: "Country:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "State",
			Prompt:    &survey.Input{Message: "State:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "Locality",
			Prompt:    &survey.Input{Message: "Locality:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "Email",
			Prompt:    &survey.Input{Message: "Email:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "TTL",
			Prompt:    &survey.Input{Message: "TTL:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "PkiStorePath",
			Prompt:    &survey.Input{Message: "Path to a new secret in which to store the generated certificate with private key:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:   "Chain",
			Prompt: &survey.Confirm{Message: "Chain (optional - include root certificate):", Default: false},
		},
	}

	answers := struct {
		RootCAPath   string
		CommonName   string
		Organization string
		Country      string
		State        string
		Locality     string
		Email        string
		TTL          string
		PkiStorePath string
		Chain        bool
	}{}
	survErr := survey.Ask(qs, &answers)
	if survErr != nil {
		vcli.Out().Fail(survErr)
		return utils.GetExecStatus(survErr)
	}

	params := map[string]string{
		cst.RootCAPath:   answers.RootCAPath,
		cst.CommonName:   answers.CommonName,
		cst.Organization: answers.Organization,
		cst.Country:      answers.Country,
		cst.State:        answers.State,
		cst.Locality:     answers.Locality,
		cst.EmailAddress: answers.Email,
		cst.TTL:          answers.TTL,
		cst.PkiStorePath: answers.PkiStorePath,
		cst.Chain:        strconv.FormatBool(answers.Chain),
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

	return pkiLeaf(vcli, &body)
}

func handleGenerateRootWizard(vcli vaultcli.CLI) int {
	qs := []*survey.Question{
		{
			Name:      "RootCAPath",
			Prompt:    &survey.Input{Message: "Path of a new secret in which to store the generated root certificate with private key:"},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "CommonName",
			Prompt:    &survey.Input{Message: "Common name:"},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "Organization",
			Prompt:    &survey.Input{Message: "Organization:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "Country",
			Prompt:    &survey.Input{Message: "Country:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "State",
			Prompt:    &survey.Input{Message: "State:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "Locality",
			Prompt:    &survey.Input{Message: "Locality:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "Email",
			Prompt:    &survey.Input{Message: "Email:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "Domains",
			Prompt:    &survey.Input{Message: "List of domains (comma-delimited strings):"},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "MaxTTL",
			Prompt:    &survey.Input{Message: "Maximum TTL:"},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "CRL",
			Prompt:    &survey.Input{Message: "Certificate Revocation List URL:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
	}

	answers := struct {
		RootCAPath   string
		CommonName   string
		Organization string
		Country      string
		State        string
		Locality     string
		Email        string
		Domains      string
		MaxTTL       string
		CRL          string
	}{}
	survErr := survey.Ask(qs, &answers)
	if survErr != nil {
		vcli.Out().Fail(survErr)
		return utils.GetExecStatus(survErr)
	}

	params := map[string]string{
		cst.RootCAPath:   answers.RootCAPath,
		cst.CommonName:   answers.CommonName,
		cst.Organization: answers.Organization,
		cst.Country:      answers.Country,
		cst.State:        answers.State,
		cst.Locality:     answers.Locality,
		cst.EmailAddress: answers.Email,
		cst.Domains:      answers.Domains,
		cst.MaxTTL:       answers.MaxTTL,
		cst.CRL:          answers.CRL,
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

	return pkiGenRoot(vcli, &body)
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

// API callers:

type rootCASecret struct {
	RootCAPath  string   `json:"rootCAPath"`
	PrivateKey  string   `json:"privateKey"`
	Certificate string   `json:"certificate"`
	Domains     []string `json:"domains"`
	MaxTTL      int      `json:"maxTTL"`
	CRL         string   `json:"crl"`
}

func pkiRegister(vcli vaultcli.CLI, body *rootCASecret) ([]byte, *apperrors.ApiError) {
	basePath := strings.Join([]string{cst.NounPki, cst.Register}, "/")
	uri := paths.CreateURI(basePath, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
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

type generateRootInformation struct {
	subjectInformation
	RootCAPath string   `json:"rootCAPath"`
	Domains    []string `json:"domains"`
	MaxTTL     int      `json:"maxTTL"`
	CRL        string   `json:"crl"`
}

func pkiGenRoot(vcli vaultcli.CLI, body *generateRootInformation) ([]byte, *apperrors.ApiError) {
	basePath := strings.Join([]string{cst.NounPki, cst.Root}, "/")
	uri := paths.CreateURI(basePath, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
}

type signingRequestInformation struct {
	subjectInformation
	RootCAPath string `json:"rootCAPath"`
	StorePath  string `json:"storePath"`
	TTL        int    `json:"ttl"`
	Chain      bool   `json:"chain"`
}

func pkiLeaf(vcli vaultcli.CLI, body *signingRequestInformation) ([]byte, *apperrors.ApiError) {
	basePath := strings.Join([]string{cst.NounPki, cst.Leaf}, "/")
	uri := paths.CreateURI(basePath, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
}

type sshCertificateInformation struct {
	RootCAPath string   `json:"rootCAPath"`
	LeafCAPAth string   `json:"leafCAPath"`
	Principals []string `json:"principals"`
	TTL        int      `json:"ttl"`
}

func pkiSSHCert(vcli vaultcli.CLI, body *sshCertificateInformation) ([]byte, *apperrors.ApiError) {
	basePath := strings.Join([]string{cst.NounPki, cst.SSHCert}, "/")
	uri := paths.CreateURI(basePath, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
}

type signingRequest struct {
	RootCAPath      string   `json:"rootCAPath"`
	CSR             string   `json:"csr"`
	SubjectAltNames []string `json:"subjectAltNames"`
	TTL             int      `json:"ttl"`
	Chain           bool     `json:"chain"`
}

func pkiSign(vcli vaultcli.CLI, body *signingRequest) ([]byte, *apperrors.ApiError) {
	basePath := strings.Join([]string{cst.NounPki, cst.Sign}, "/")
	uri := paths.CreateURI(basePath, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
}
