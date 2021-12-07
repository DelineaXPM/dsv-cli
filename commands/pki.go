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
	"thy/format"
	"thy/paths"
	preds "thy/predictors"
	"thy/requests"
	"thy/store"
	"thy/utils"

	"github.com/posener/complete"
	"github.com/spf13/viper"
	"github.com/thycotic-rd/cli"
)

type pki struct {
	request   requests.Client
	outClient format.OutClient
}

func GetPkiCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounPki},
		RunFunc: func(args []string) int {
			return cli.RunResultHelp
		},
		SynopsisText:  "pki (<action>)",
		HelpText:      "Work with certificates",
		MinNumberArgs: 0,
	})
}

func GetPkiRegisterCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPki, cst.Register},
		RunFunc:      pki{requests.NewHttpClient(), nil}.handleRegisterRoot,
		SynopsisText: "Register an existing root certificate authority",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s @cert.pem --%[4]s @key.pem --%[5]s myroot --%[6]s google.com,yahoo.com --%[7]s 1000
		`, cst.NounPki, cst.Register, cst.CertPath, cst.PrivKeyPath, cst.RootCAPath, cst.Domains, cst.MaxTTL),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.CertPath):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.CertPath, Usage: "Path to a file containing the root certificate (required)"}), false},
			preds.LongFlag(cst.PrivKeyPath): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.PrivKeyPath, Usage: "Path to a file containing the private key (required)"}), false},
			preds.LongFlag(cst.RootCAPath):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.RootCAPath, Usage: "Path to a secret which will contain the registered root certificate with private key (required)"}), false},
			preds.LongFlag(cst.Domains):     cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Domains, Usage: "List of domains for which certificates could be signed on behalf of the root CA (required)"}), false},
			preds.LongFlag(cst.MaxTTL):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.MaxTTL, Usage: "Maximum number of hours for which a signed certificate on behalf of the root CA can be valid (required)"}), false},
			preds.LongFlag(cst.CRL):         cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.CRL, Usage: "URL of the CRL from which the revocation of leaf certificates can be checked"}), false},
		},
		MinNumberArgs: 0,
	})
}

func GetPkiSignCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPki, cst.Sign},
		RunFunc:      pki{requests.NewHttpClient(), nil}.handleSign,
		SynopsisText: "Get a new certificate specified by a CSR and signed by a registered root CA",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myroot --%[4]s @csr.pem --%[5]s google.com,android.com --%[6]s 1000h
		`, cst.NounPki, cst.Sign, cst.RootCAPath, cst.CSRPath, cst.SubjectAltNames, cst.TTL),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.CSRPath):         cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.CSRPath, Usage: "Path to a file containing the CSR (required)"}), false},
			preds.LongFlag(cst.RootCAPath):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.RootCAPath, Usage: "Path to a secret which contains the registered root certificate with private key (required)"}), false},
			preds.LongFlag(cst.SubjectAltNames): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.SubjectAltNames, Usage: "List of subject alternative names (domains) for a certificate signed on behalf of the root CA can also be valid"}), false},
			preds.LongFlag(cst.TTL):             cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.TTL, Usage: "Number of hours for which a signed certificate on behalf of the root CA can be valid"}), false},
			preds.LongFlag(cst.Chain):           cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Chain, Usage: "Include root certificate in response", ValueType: "bool"}), false},
		},
		MinNumberArgs: 0,
	})
}

func GetPkiLeafCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPki, cst.Leaf},
		RunFunc:      pki{requests.NewHttpClient(), nil}.handleLeaf,
		SynopsisText: "Get a new private key and leaf certificate signed by a registered root CA",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myroot --%[4]s myleafcert --%[5]s thycotic.com --%[6]s Thycotic --%[7]s US --%[8]s DC --%[9]s Washington --%[10]s 100d
   • %[1]s %[2]s --%[3]s myroot --%[5]s thycotic.com
		`, cst.NounPki, cst.Leaf, cst.RootCAPath, cst.PkiStorePath, cst.CommonName, cst.Organization, cst.Country, cst.State, cst.Locality, cst.TTL),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.CommonName):   {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.CommonName, Usage: "Domain for which a certificate is generated (required)", Global: false}), false},
			preds.LongFlag(cst.Organization): {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Organization, Usage: "", Global: false}), false},
			preds.LongFlag(cst.Country):      {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Country, Usage: "", Global: false}), false},
			preds.LongFlag(cst.State):        {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.State, Usage: "", Global: false}), false},
			preds.LongFlag(cst.Locality):     {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Locality, Usage: "", Global: false}), false},
			preds.LongFlag(cst.EmailAddress): {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.EmailAddress, Usage: "", Global: false}), false},

			preds.LongFlag(cst.RootCAPath):   cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.RootCAPath, Usage: "Path to a secret which contains the registered root certificate with private key (required)"}), false},
			preds.LongFlag(cst.PkiStorePath): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.PkiStorePath, Usage: "Path to a new secret in which to store the generated certificate with private key"}), false},
			preds.LongFlag(cst.TTL):          cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.TTL, Usage: "Number of hours for which a signed certificate on behalf of the root CA can be valid"}), false},
			preds.LongFlag(cst.Chain):        cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Chain, Usage: "Include root certificate in response", ValueType: "bool"}), false},
		},
		MinNumberArgs: 0,
	})
}

func GetPkiGenerateRootCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPki, cst.GenerateRoot},
		RunFunc:      pki{requests.NewHttpClient(), nil}.handleGenerateRoot,
		SynopsisText: "Generate and store a new root certificate with private key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myroot --%[4]s google.org,golang.org --%[5]s thycotic.com --%[6]s Thycotic --%[7]s US --%[8]s DC --%[9]s Washington --%[10]s 42d
   • %[1]s %[2]s --%[3]s myroot --%[4]s google.org,golang.org --%[5]s thycotic.com --%[10]s 52w
		`, cst.NounPki, cst.GenerateRoot, cst.RootCAPath, cst.Domains, cst.CommonName, cst.Organization, cst.Country, cst.State, cst.Locality, cst.MaxTTL),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.CommonName):   {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.CommonName, Usage: "The domain name of the root CA (required)", Global: false}), false},
			preds.LongFlag(cst.Organization): {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Organization, Usage: "", Global: false}), false},
			preds.LongFlag(cst.Country):      {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Country, Usage: "", Global: false}), false},
			preds.LongFlag(cst.State):        {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.State, Usage: "", Global: false}), false},
			preds.LongFlag(cst.Locality):     {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Locality, Usage: "", Global: false}), false},
			preds.LongFlag(cst.EmailAddress): {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.EmailAddress, Usage: "", Global: false}), false},

			preds.LongFlag(cst.RootCAPath): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.RootCAPath, Usage: "Path of a new secret in which to store the generated root certificate with private key (required)"}), false},
			preds.LongFlag(cst.Domains):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Domains, Usage: "List of domains for which certificates could be signed on behalf of the root CA (required)"}), false},
			preds.LongFlag(cst.MaxTTL):     cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.MaxTTL, Usage: "Number of hours for which a generated root certificate can be valid (required)"}), false},
			preds.LongFlag(cst.CRL):        cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.CRL, Usage: "URL of the CRL from which the revocation of leaf certificates can be checked"}), false},
		},
		MinNumberArgs: 0,
	})
}

func GetPkiSSHCertCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPki, cst.Sign},
		RunFunc:      pki{requests.NewHttpClient(), nil}.handleGetSSHCertificate,
		SynopsisText: "Get a new SSH certificate",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myroot --%[4]s myleaf --%[5]s root,ubuntu --%[6]s 1000
		`, cst.NounPki, cst.SSHCert, cst.RootCAPath, cst.LeafCAPath, cst.Principals, cst.TTL),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.RootCAPath): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.RootCAPath, Usage: "Path to a secret which contains the registered root certificate with private key (required)"}), false},
			preds.LongFlag(cst.LeafCAPath): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.LeafCAPath, Usage: "Path to a secret which contains the leaf certificate with SSH-compatible public key (required)"}), false},
			preds.LongFlag(cst.Principals): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Principals, Usage: "List of principals on the certificate (required)"}), false},
			preds.LongFlag(cst.TTL):        cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.TTL, Usage: "Number of hours for which a signed certificate can be valid (required)"}), false},
		},
		MinNumberArgs: 8,
	})
}

func (p pki) handleRegisterRootWorkflow(args []string) int {
	params := make(map[string]string)
	ui := &PasswordUi{
		cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
	}

	if certPath, err := getStringAndValidate(
		ui, "Path to certificate file:", false, nil, false, false); err != nil {
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

	if privKeyPath, err := getStringAndValidate(
		ui, "Path to private key file:", false, nil, false, false); err != nil {
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

	if rootCAPath, err := getStringAndValidate(
		ui, "Path to a new secret that will contain root CA registration information:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.RootCAPath] = rootCAPath
	}

	if domains, err := getStringAndValidate(
		ui, "List of domains (comma-delimited strings):", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.Domains] = domains
	}

	if maxTTL, err := getStringAndValidate(
		ui, "Maximum TTL:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.MaxTTL] = maxTTL
	}

	if crl, err := getStringAndValidate(
		ui, "Certificate Revocation List URL (optional):", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.CRL] = crl
	}

	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}
	resp, err := p.submitRoot(params)
	p.outClient.WriteResponse(resp, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func (p pki) submitRoot(params map[string]string) ([]byte, error) {
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
	return p.request.DoRequest(http.MethodPost, uri, body)
}

func (p pki) handleRegisterRoot(args []string) int {
	if OnlyGlobalArgs(args) {
		return p.handleRegisterRootWorkflow(args)
	}
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}

	params := make(map[string]string)
	params[cst.CertPath] = viper.GetString(cst.CertPath)
	params[cst.PrivKeyPath] = viper.GetString(cst.PrivKeyPath)
	params[cst.RootCAPath] = viper.GetString(cst.RootCAPath)
	params[cst.Domains] = viper.GetString(cst.Domains)
	params[cst.MaxTTL] = viper.GetString(cst.MaxTTL)

	data, err := p.submitRoot(params)
	p.outClient.WriteResponse(data, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func (p pki) handleSignWorkflow(args []string) int {
	params := make(map[string]string)
	ui := &PasswordUi{
		cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
	}

	if csrPath, err := getStringAndValidate(
		ui, "Path to certificate signing request file:", false, nil, false, false); err != nil {
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

	if rootCAPath, err := getStringAndValidate(
		ui, "Path of an existing secret that contains root CA information:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.RootCAPath] = rootCAPath
	}

	if subjectAltNames, err := getStringAndValidate(
		ui, "List of subject alternative names (comma-delimited strings, optional):", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.SubjectAltNames] = subjectAltNames
	}

	if ttl, err := getStringAndValidate(
		ui, "TTL:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.TTL] = ttl
	}

	if resp, err := getStringAndValidateDefault(
		ui, "Chain (optional - include root certificate) [y/N]:", "N", false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		if isYes(resp, true) {
			params[cst.Chain] = "true"
		} else {
			params[cst.Chain] = "false"
		}
	}

	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}
	resp, err := p.submitSign(params)
	p.outClient.WriteResponse(resp, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func (p pki) handleSign(args []string) int {
	if OnlyGlobalArgs(args) {
		return p.handleSignWorkflow(args)
	}
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}

	params := make(map[string]string)
	params[cst.CSRPath] = viper.GetString(cst.CSRPath)
	params[cst.RootCAPath] = viper.GetString(cst.RootCAPath)
	params[cst.SubjectAltNames] = viper.GetString(cst.SubjectAltNames)
	params[cst.TTL] = viper.GetString(cst.TTL)
	params[cst.Chain] = viper.GetString(cst.Chain)

	data, err := p.submitSign(params)
	p.outClient.WriteResponse(data, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func (p pki) submitSign(params map[string]string) ([]byte, error) {
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
	return p.request.DoRequest(http.MethodPost, uri, body)
}

func (p pki) handleLeafWorkflow(args []string) int {
	params := make(map[string]string)
	ui := &PasswordUi{
		cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
	}

	if rootCAPath, err := getStringAndValidate(
		ui, "Path of an existing secret that contains root CA information:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.RootCAPath] = rootCAPath
	}

	if commonName, err := getStringAndValidate(
		ui, "Common name:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.CommonName] = commonName
	}

	if organization, err := getStringAndValidate(
		ui, "Organization:", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.Organization] = organization
	}

	if country, err := getStringAndValidate(
		ui, "Country:", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.Country] = country
	}

	if state, err := getStringAndValidate(
		ui, "State:", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.State] = state
	}

	if locality, err := getStringAndValidate(
		ui, "Locality:", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.Locality] = locality
	}

	if email, err := getStringAndValidate(
		ui, "Email:", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.EmailAddress] = email
	}

	if ttl, err := getStringAndValidate(
		ui, "TTL:", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.TTL] = ttl
	}

	if storePath, err := getStringAndValidate(
		ui, "Path to a new secret in which to store the generated certificate with private key (optional):", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.PkiStorePath] = storePath
	}

	if resp, err := getStringAndValidateDefault(
		ui, "Chain (optional - include root certificate) [y/N]:", "N", false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		if isYes(resp, true) {
			params[cst.Chain] = "true"
		} else {
			params[cst.Chain] = "false"
		}
	}

	resp, err := p.submitLeaf(params)
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}
	p.outClient.WriteResponse(resp, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func (p pki) submitLeaf(params map[string]string) ([]byte, error) {
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
	return p.request.DoRequest(http.MethodPost, uri, body)
}

func (p pki) handleLeaf(args []string) int {
	if OnlyGlobalArgs(args) {
		return p.handleLeafWorkflow(args)
	}
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
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
	params[cst.Chain] = viper.GetString(cst.Chain)

	data, err := p.submitLeaf(params)
	p.outClient.WriteResponse(data, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func (p pki) handleGenerateRootWorkflow(args []string) int {
	params := make(map[string]string)
	ui := &PasswordUi{
		cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
	}

	if rootCAPath, err := getStringAndValidate(
		ui, "Path of a new secret in which to store the generated root certificate with private key:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.RootCAPath] = rootCAPath
	}

	if commonName, err := getStringAndValidate(
		ui, "Common name:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.CommonName] = commonName
	}

	if organization, err := getStringAndValidate(
		ui, "Organization:", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.Organization] = organization
	}

	if country, err := getStringAndValidate(
		ui, "Country:", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.Country] = country
	}

	if state, err := getStringAndValidate(
		ui, "State:", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.State] = state
	}

	if locality, err := getStringAndValidate(
		ui, "Locality:", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.Locality] = locality
	}

	if email, err := getStringAndValidate(
		ui, "Email:", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.EmailAddress] = email
	}

	if domains, err := getStringAndValidate(
		ui, "List of domains (comma-delimited strings):", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.Domains] = domains
	}

	if maxTTL, err := getStringAndValidate(
		ui, "Maximum TTL:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.MaxTTL] = maxTTL
	}

	if crl, err := getStringAndValidate(
		ui, "Certificate Revocation List URL (optional):", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.CRL] = crl
	}

	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}
	resp, err := p.submitGenerateRoot(params)
	p.outClient.WriteResponse(resp, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func (p pki) submitGenerateRoot(params map[string]string) ([]byte, error) {
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
	maxTTL, err := utils.ParseHours(params[cst.MaxTTL])
	if err != nil {
		return nil, err
	}
	body.MaxTTL = maxTTL

	basePath := strings.Join([]string{cst.NounPki, cst.Root}, "/")
	uri := paths.CreateURI(basePath, nil)
	return p.request.DoRequest(http.MethodPost, uri, body)
}

func (p pki) handleGenerateRoot(args []string) int {
	if OnlyGlobalArgs(args) {
		return p.handleGenerateRootWorkflow(args)
	}
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
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

	data, err := p.submitGenerateRoot(params)
	p.outClient.WriteResponse(data, apperrors.New(err))
	return utils.GetExecStatus(err)
}

func (p pki) handleGetSSHCertificate(args []string) int {
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}
	params := make(map[string]string)
	params[cst.RootCAPath] = viper.GetString(cst.RootCAPath)
	params[cst.LeafCAPath] = viper.GetString(cst.LeafCAPath)
	params[cst.Principals] = viper.GetString(cst.Principals)
	params[cst.TTL] = viper.GetString(cst.TTL)

	paramErr := ValidateParams(params, []string{cst.RootCAPath, cst.LeafCAPath, cst.Principals, cst.TTL})
	if paramErr != nil {
		p.outClient.Fail(paramErr)
		return utils.GetExecStatus(paramErr)
	}

	body := sshCertificateInformation{
		RootCAPath: params[cst.RootCAPath],
		LeafCAPAth: params[cst.LeafCAPath],
		Principals: utils.StringToSlice(params[cst.Principals]),
	}

	if ttl, err := utils.ParseHours(params[cst.TTL]); err != nil {
		p.outClient.Fail(err)
		return utils.GetExecStatus(err)
	} else {
		body.TTL = ttl
	}

	basePath := strings.Join([]string{cst.NounPki, cst.SSHCert}, "/")
	uri := paths.CreateURI(basePath, nil)
	resp, apiError := p.request.DoRequest(http.MethodPost, uri, body)
	p.outClient.WriteResponse(resp, apiError)
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
	companyInformation
	RootCAPath string `json:"rootCAPath"`
	StorePath  string `json:"storePath"`
	TTL        int    `json:"ttl"`
	Chain      bool   `json:"chain"`
}

type generateRootInformation struct {
	companyInformation
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

type companyInformation struct {
	Country            string `json:"country"`
	State              string `json:"state"`
	Locality           string `json:"locality"`
	Organization       string `json:"organization"`
	OrganizationalUnit string `json:"organizationalUnit"`
	CommonName         string `json:"commonName"`
	EmailAddress       string `json:"emailAddress"`
}
