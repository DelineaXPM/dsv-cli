package cmd

import (
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	cst "thy/constants"
	apperrors "thy/errors"
	"thy/format"
	preds "thy/predictors"
	"thy/requests"
	"thy/utils"

	"github.com/posener/complete"
	"github.com/thycotic-rd/cli"
	"github.com/thycotic-rd/viper"
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
		},
		MinNumberArgs: 5,
	})
}

func GetPkiSignCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPki, cst.Sign},
		RunFunc:      pki{requests.NewHttpClient(), nil}.handleSign,
		SynopsisText: "Get a new certificate specified by a CSR and signed by a registered root CA",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myroot --%[4]s @csr.pem --%[5]s google.com,android.com --%[6]s 1000
		`, cst.NounPki, cst.Sign, cst.RootCAPath, cst.CSRPath, cst.SubjectAltNames, cst.TTL),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.CSRPath):         cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.CSRPath, Usage: "Path to a file containing the CSR (required)"}), false},
			preds.LongFlag(cst.RootCAPath):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.RootCAPath, Usage: "Path to a secret which contains the registered root certificate with private key (required)"}), false},
			preds.LongFlag(cst.SubjectAltNames): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.SubjectAltNames, Usage: "List of subject alternative names (domains) for a certificate signed on behalf of the root CA can also be valid"}), false},
			preds.LongFlag(cst.TTL):             cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.TTL, Usage: "Number of hours for which a signed certificate on behalf of the root CA can be valid"}), false},
		},
		MinNumberArgs: 2,
	})
}

func GetPkiLeafCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPki, cst.Leaf},
		RunFunc:      pki{requests.NewHttpClient(), nil}.handleLeaf,
		SynopsisText: "Get a new private key and leaf certificate signed by a registered root CA",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myroot --%[4]s myleafcert --%[5]s thycotic.com --%[6]s Thycotic --%[7]s US --%[8]s DC --%[9]s Washington --%[10]s 100
   • %[1]s %[2]s --%[3]s myroot --%[5]s thycotic.com
		`, cst.NounPki, cst.Leaf, cst.RootCAPath, cst.PkiStorePath, cst.CommonName, cst.Organization, cst.Country, cst.State, cst.Locality, cst.TTL),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.CommonName):   {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.CommonName, Usage: "Domain for which a certificate is generated", Global: false}), false},
			preds.LongFlag(cst.Organization): {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Organization, Usage: "", Global: false}), false},
			preds.LongFlag(cst.Country):      {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Country, Usage: "", Global: false}), false},
			preds.LongFlag(cst.State):        {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.State, Usage: "", Global: false}), false},
			preds.LongFlag(cst.Locality):     {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Locality, Usage: "", Global: false}), false},
			preds.LongFlag(cst.EmailAddress): {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.EmailAddress, Usage: "", Global: false}), false},

			preds.LongFlag(cst.RootCAPath):   cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.RootCAPath, Usage: "Path to a secret which contains the registered root certificate with private key (required)"}), false},
			preds.LongFlag(cst.PkiStorePath): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.PkiStorePath, Usage: "Path to a new secret in which to store the generated certificate with private key"}), false},
			preds.LongFlag(cst.TTL):          cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.TTL, Usage: "Number of hours for which a signed certificate on behalf of the root CA can be valid"}), false},
		},
		MinNumberArgs: 2,
	})
}

func GetPkiGenerateRootCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPki, cst.GenerateRoot},
		RunFunc:      pki{requests.NewHttpClient(), nil}.handleGenerateRoot,
		SynopsisText: "Generate and store a new root certificate with private key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myroot --%[4]s google.org,golang.org --%[5]s thycotic.com --%[6]s Thycotic --%[7]s US --%[8]s DC --%[9]s Washington --%[10]s 1000
   • %[1]s %[2]s --%[3]s myroot --%[4]s google.org,golang.org --%[5]s thycotic.com --%[10]s 1000
		`, cst.NounPki, cst.GenerateRoot, cst.RootCAPath, cst.Domains, cst.CommonName, cst.Organization, cst.Country, cst.State, cst.Locality, cst.MaxTTL),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.CommonName):   {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.CommonName, Usage: "The domain name of the root CA", Global: false}), false},
			preds.LongFlag(cst.Organization): {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Organization, Usage: "", Global: false}), false},
			preds.LongFlag(cst.Country):      {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Country, Usage: "", Global: false}), false},
			preds.LongFlag(cst.State):        {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.State, Usage: "", Global: false}), false},
			preds.LongFlag(cst.Locality):     {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Locality, Usage: "", Global: false}), false},
			preds.LongFlag(cst.EmailAddress): {complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.EmailAddress, Usage: "", Global: false}), false},

			preds.LongFlag(cst.RootCAPath): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.RootCAPath, Usage: "Path to a new secret in which to store the generated root certificate with private key (required)"}), false},
			preds.LongFlag(cst.Domains):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Domains, Usage: "List of domains for which certificates could be signed on behalf of the root CA (required)"}), false},
			preds.LongFlag(cst.MaxTTL):     cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.MaxTTL, Usage: "Number of hours for which a generated root certificate can be valid (required)"}), false},
		},
		MinNumberArgs: 2,
	})
}

func (p pki) handleRegisterRoot([]string) int {
	var apiError *apperrors.ApiError
	var data []byte
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}

	params := make(map[string]string)
	params[cst.CertPath] = viper.GetString(cst.CertPath)
	params[cst.PrivKeyPath] = viper.GetString(cst.PrivKeyPath)
	params[cst.RootCAPath] = viper.GetString(cst.RootCAPath)
	params[cst.Domains] = viper.GetString(cst.Domains)
	params[cst.MaxTTL] = viper.GetString(cst.MaxTTL)

	paramErr := ValidateParams(params, utils.Keys(params))
	if paramErr != nil {
		p.outClient.Fail(paramErr)
		return 1
	}

	_, err := parsePem(params[cst.CertPath])
	if err != nil {
		p.outClient.FailF("certificate data error: %v", err)
		return 1
	}
	_, err = parsePem(params[cst.PrivKeyPath])
	if err != nil {
		p.outClient.FailF("private key data error: %v", err)
		return 1
	}

	maxTTL, err := strconv.Atoi(params[cst.MaxTTL])
	if err != nil {
		p.outClient.FailF("failed to convert %s to a numeric value", params[cst.MaxTTL])
		return 1
	}
	body := rootCASecret{
		RootCAPath: params[cst.RootCAPath],
		PrivateKey: base64Encode(params[cst.PrivKeyPath]),
		Cert:       base64Encode(params[cst.CertPath]),
		Domains:    utils.StringToSlice(params[cst.Domains]),
		MaxTTL:     maxTTL,
	}

	basePath := strings.Join([]string{cst.NounPki, cst.Register}, "/")
	uri := utils.CreateURI(basePath, nil)
	data, apiError = p.request.DoRequest(http.MethodPost, uri, body)
	p.outClient.WriteResponse(data, apiError)
	return utils.GetExecStatus(apiError)
}

func (p pki) handleSign([]string) int {
	var apiError *apperrors.ApiError
	var data []byte
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}

	params := make(map[string]string)
	params[cst.CSRPath] = viper.GetString(cst.CSRPath)
	params[cst.RootCAPath] = viper.GetString(cst.RootCAPath)
	params[cst.SubjectAltNames] = viper.GetString(cst.SubjectAltNames)
	params[cst.TTL] = viper.GetString(cst.TTL)

	paramErr := ValidateParams(params, []string{cst.CSRPath, cst.RootCAPath})
	if paramErr != nil {
		p.outClient.Fail(paramErr)
		return 1
	}

	_, err := parsePem(params[cst.CSRPath])
	if err != nil {
		p.outClient.FailF("certificate signing request data error: %v", err)
		return 1
	}

	body := signingRequest{
		RootCAPath:      params[cst.RootCAPath],
		CSR:             base64Encode(params[cst.CSRPath]),
		SubjectAltNames: utils.StringToSlice(params[cst.SubjectAltNames]),
	}
	if params[cst.TTL] != "" {
		ttl, err := strconv.Atoi(params[cst.TTL])
		if err != nil {
			p.outClient.FailF("failed to convert %s to a numeric value", params[cst.TTL])
			return 1
		}
		body.TTL = ttl
	}

	basePath := strings.Join([]string{cst.NounPki, cst.Sign}, "/")
	uri := utils.CreateURI(basePath, nil)
	data, apiError = p.request.DoRequest(http.MethodPost, uri, body)
	p.outClient.WriteResponse(data, apiError)
	return utils.GetExecStatus(apiError)
}

func (p pki) handleLeaf([]string) int {
	var apiError *apperrors.ApiError
	var data []byte
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

	paramErr := ValidateParams(params, []string{cst.RootCAPath, cst.CommonName})
	if paramErr != nil {
		p.outClient.Fail(paramErr)
		return 1
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
	if params[cst.TTL] != "" {
		ttl, err := strconv.Atoi(params[cst.TTL])
		if err != nil {
			p.outClient.FailF("failed to convert %s to a numeric value", params[cst.TTL])
			return 1
		}
		body.TTL = ttl
	}

	basePath := strings.Join([]string{cst.NounPki, cst.Leaf}, "/")
	uri := utils.CreateURI(basePath, nil)
	data, apiError = p.request.DoRequest(http.MethodPost, uri, body)
	p.outClient.WriteResponse(data, apiError)
	return utils.GetExecStatus(apiError)
}

func (p pki) handleGenerateRoot([]string) int {
	var apiError *apperrors.ApiError
	var data []byte
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

	paramErr := ValidateParams(params, []string{cst.RootCAPath, cst.CommonName, cst.Domains, cst.MaxTTL})
	if paramErr != nil {
		p.outClient.Fail(paramErr)
		return 1
	}

	body := generateRootInformation{
		RootCAPath: params[cst.RootCAPath],
		Domains:    utils.StringToSlice(params[cst.Domains]),
	}
	body.CommonName = params[cst.CommonName]
	body.Organization = params[cst.Organization]
	body.Country = params[cst.Country]
	body.State = params[cst.State]
	body.Locality = params[cst.Locality]
	body.EmailAddress = params[cst.EmailAddress]
	ttl, err := strconv.Atoi(params[cst.MaxTTL])
	if err != nil {
		p.outClient.FailF("failed to convert %s to a numeric value", params[cst.MaxTTL])
		return 1
	}
	body.MaxTTL = ttl

	basePath := strings.Join([]string{cst.NounPki, cst.Root}, "/")
	uri := utils.CreateURI(basePath, nil)
	data, apiError = p.request.DoRequest(http.MethodPost, uri, body)
	p.outClient.WriteResponse(data, apiError)
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
	RootCAPath string   `json:"rootCAPath"`
	PrivateKey string   `json:"privateKey"`
	Cert       string   `json:"cert"`
	Domains    []string `json:"domains"`
	MaxTTL     int      `json:"maxTTL"`
}

type signingRequest struct {
	RootCAPath      string   `json:"rootCAPath"`
	CSR             string   `json:"csr"`
	SubjectAltNames []string `json:"subjectAltNames"`
	TTL             int      `json:"ttl"`
}

type signingRequestInformation struct {
	companyInformation
	RootCAPath string `json:"rootCAPath"`
	StorePath  string `json:"storePath"`
	TTL        int    `json:"ttl"`
}

type generateRootInformation struct {
	companyInformation
	RootCAPath string   `json:"rootCAPath"`
	Domains    []string `json:"domains"`
	MaxTTL     int      `json:"maxTTL"`
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
