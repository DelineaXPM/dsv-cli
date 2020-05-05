package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	cst "thy/constants"
	"thy/errors"
	"thy/format"
	preds "thy/predictors"
	"thy/requests"
	"thy/utils"
	"thy/version"

	"github.com/posener/complete"
	"github.com/thycotic-rd/cli"
	"github.com/thycotic-rd/viper"
)

type AuthProvider struct {
	request   requests.Client
	outClient format.OutClient
}

func (p AuthProvider) Run(args []string) int {
	ap := cli.NewCLI(fmt.Sprintf("%s %s %s", cst.CmdRoot, cst.NounConfig, cst.NounAuthProvider), version.Version)
	ap.Args = args
	ap.Commands = map[string]cli.CommandFactory{
		"read":    GetAuthProviderReadCmd,
		"search":  GetAuthProviderSearchCommand,
		"delete":  GetAuthProviderDeleteCmd,
		"restore": GetAuthProviderRestoreCmd,
		"create":  GetAuthProviderCreateCmd,
		"update":  GetAuthProviderUpdateCmd,
	}

	exitStatus, err := ap.Run()
	if err != nil {
		return utils.GetExecStatus(err)
	}
	return exitStatus
}

func GetNoDataOpAuthProviderWrappers() cli.PredictorWrappers {
	return cli.PredictorWrappers{
		preds.LongFlag(cst.DataName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataName, Shorthand: "n", Usage: fmt.Sprintf("Target %s to an %s", cst.Path, cst.NounAuthProvider)}), false},
		preds.LongFlag(cst.Version):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Version, Shorthand: "", Usage: "List the current and last (n) versions"}), false},
	}
}

func GetAuthProviderCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounConfig, cst.NounAuthProvider},
		RunFunc: func(args []string) int {
			name := viper.GetString(cst.DataName)
			if name == "" && len(args) > 0 {
				name = args[0]
			}
			if name == "" {
				return cli.RunResultHelp
			}
			return AuthProvider{requests.NewHttpClient(), nil}.handleAuthProviderReadCmd(args)
		},
		SynopsisText:   "manage 3rd party authentication providers",
		HelpText:       fmt.Sprintf("Execute an action on an %s from %s", cst.NounAuthProvider, cst.ProductName),
		FlagsPredictor: GetNoDataOpAuthProviderWrappers(),
		MinNumberArgs:  1,
	})
}

func GetAuthProviderReadCmd() (cli.Command, error) {
	forcePlain := viper.GetBool(cst.Plain)
	if !forcePlain {
		viper.Set(cst.Beautify, true)
	}
	return NewCommand(CommandArgs{
		Path:         []string{cst.Config, cst.NounAuthProvider, cst.Read},
		RunFunc:      AuthProvider{request: requests.NewHttpClient(), outClient: nil}.handleAuthProviderReadCmd,
		SynopsisText: fmt.Sprintf("%s %s %s (<name> | --name|-n)", cst.NounConfig, cst.NounAuthProvider, cst.Read),
		HelpText: fmt.Sprintf(`Read a %[1]s

Usage:
   • %[1]s %[2]s %[4]s %[3]s
   • %[1]s %[2]s %[4]s --name %[3]s 
		`, cst.NounConfig, cst.NounAuthProvider, cst.ExampleAuthProviderName, cst.Read),
		FlagsPredictor: GetNoDataOpAuthProviderWrappers(),
		MinNumberArgs:  1,
	})
}

func GetAuthProviderDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.Config, cst.NounAuthProvider, cst.Delete},
		RunFunc:      AuthProvider{request: requests.NewHttpClient(), outClient: nil}.handleAuthProviderDeleteCmd,
		SynopsisText: fmt.Sprintf("%s %s %s (<name> | --name|-n)", cst.Config, cst.NounAuthProvider, cst.Delete),
		HelpText: fmt.Sprintf(`Delete %[1]s

Usage:
  • %[1]s %[2]s %[3]s %[4]s --all
  • %[1]s %[2]s %[3]s --name %[4]s --force
		`, cst.NounConfig, cst.NounAuthProvider, cst.Delete, cst.ExampleAuthProviderName),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataName, Shorthand: "n", Usage: fmt.Sprintf("Target %s to an %s", cst.Path, cst.NounAuthProvider)}), false},
			preds.LongFlag(cst.Force):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Force, Shorthand: "", Usage: fmt.Sprintf("Immediately delete %s", cst.NounAuthProvider), Global: false, ValueType: "bool"}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetAuthProviderRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.Config, cst.NounAuthProvider, cst.Restore},
		RunFunc:      AuthProvider{request: requests.NewHttpClient(), outClient: nil}.handleAuthProviderRestoreCmd,
		SynopsisText: fmt.Sprintf("%s %s %s (<name> | --name|-n)", cst.Config, cst.NounAuthProvider, cst.Restore),
		HelpText: fmt.Sprintf(`Restore %[1]s

Usage:
  • %[1]s %[2]s %[3]s %[4]s --all
  • %[1]s %[2]s %[3]s --name %[4]s
		`, cst.NounConfig, cst.NounAuthProvider, cst.Restore, cst.ExampleAuthProviderName),
		FlagsPredictor: GetNoDataOpAuthProviderWrappers(),
		MinNumberArgs:  1,
	})
}

func GetAuthProviderCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounAuthProvider, cst.Create},
		RunFunc:      AuthProvider{request: requests.NewHttpClient(), outClient: nil}.handleAuthProviderUpsert,
		SynopsisText: fmt.Sprintf("%s %s %s (<name> | --name|-n) (--type) ((--data|-d) | --aws-account-id | --azure-tenant-id)", cst.NounConfig, cst.NounAuthProvider, cst.Create),
		HelpText: fmt.Sprintf(`Add %[1]s provider

Usage:
  • %[1]s %[2]s %[4]s %[3]s --aws-account-id 11652944433808  --type aws
  • %[1]s %[2]s %[4]s --name azure-prod --azure-tenant-id 164543 --type azure
  • %[1]s %[2]s %[4]s --data %[5]s
		`, cst.NounConfig, cst.NounAuthProvider, cst.ExampleAuthProviderName, cst.Create, cst.ExampleDataPath),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Data):          cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("%s to be stored in an auth provider. Prefix with '@' to denote filepath", strings.Title(cst.Data))}), false},
			preds.LongFlag(cst.DataType):      cli.PredictorWrapper{preds.AuthTypePredictor{}, preds.NewFlagValue(preds.Params{Name: cst.DataType, Usage: fmt.Sprintf("Auth provider type (azure,aws)")}), false},
			preds.LongFlag(cst.DataName):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataName, Shorthand: "n", Usage: fmt.Sprintf("Auth provider friendly name")}), false},
			preds.LongFlag(cst.DataTenantID):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataTenantID, Usage: fmt.Sprintf("Azure Tenant ID")}), false},
			preds.LongFlag(cst.DataAccountID): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataAccountID, Usage: fmt.Sprintf("AWS Account ID")}), false},
		},
		MinNumberArgs: 2,
	})
}

func GetAuthProviderUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.Config, cst.NounAuthProvider, cst.Update},
		RunFunc:      AuthProvider{request: requests.NewHttpClient(), outClient: nil}.handleAuthProviderUpsert,
		SynopsisText: fmt.Sprintf("%s %s %s (<name> | --name|-n) (--type) ((--data|-d) | --aws-account-id | --azure-tenant-id)", cst.NounConfig, cst.NounAuthProvider, cst.Update),
		HelpText: fmt.Sprintf(`Update %[1]s properties

Usage:
  • %[1]s %[2]s %[4]s %[3]s --aws-account-id 11652944433808  --type aws
  • %[1]s %[2]s --name azure-prod --azure-tenant-id 164543 --type azure
  • %[1]s %[2]s %[4]s --data %[5]s
		`, cst.NounConfig, cst.NounAuthProvider, cst.ExampleAuthProviderName, cst.Update, cst.ExampleDataPath),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Data):          cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("%s to be stored in an auth provider. Prefix with '@' to denote filepath", strings.Title(cst.Data))}), false},
			preds.LongFlag(cst.DataType):      cli.PredictorWrapper{preds.AuthTypePredictor{}, preds.NewFlagValue(preds.Params{Name: cst.DataType, Usage: fmt.Sprintf("Auth provider type (azure,aws)")}), false},
			preds.LongFlag(cst.DataName):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataName, Shorthand: "n", Usage: fmt.Sprintf("Auth provider friendly name")}), false},
			preds.LongFlag(cst.DataTenantID):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataTenantID, Usage: fmt.Sprintf("Azure Tenant ID")}), false},
			preds.LongFlag(cst.DataAccountID): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataAccountID, Usage: fmt.Sprintf("AWS Account ID")}), false},
		},
		MinNumberArgs: 2,
	})
}

func GetAuthProviderRollbackCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.Config, cst.NounAuthProvider, cst.Rollback},
		RunFunc:      AuthProvider{request: requests.NewHttpClient(), outClient: nil}.handleAuthProviderRollbackCmd,
		SynopsisText: fmt.Sprintf("%s %s %s (<name> | --name|-n)", cst.NounConfig, cst.NounAuthProvider, cst.Rollback),
		HelpText: fmt.Sprintf(`Rollback %[1]s properties

Usage:
  • %[1]s %[2]s %[4]s %[3]s
  • %[1]s %[2]s %[4]s --version %[5]s 1
		`, cst.NounConfig, cst.NounAuthProvider, cst.ExampleAuthProviderName, cst.Rollback, cst.Version),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataName, Shorthand: "n", Usage: fmt.Sprintf("Target %s to an %s", cst.Path, cst.NounAuthProvider)}), false},
			preds.LongFlag(cst.Version):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Version, Shorthand: "", Usage: "The version to which to rollback"}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetAuthProviderSearchCommand() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.Config, cst.NounAuthProvider, cst.Search},
		RunFunc:      AuthProvider{request: requests.NewHttpClient(), outClient: nil}.handleAuthProviderSearchCommand,
		SynopsisText: fmt.Sprintf("%s %s %s (<query> | --query)", cst.NounConfig, cst.NounAuthProvider, cst.Search),
		HelpText: fmt.Sprintf(`Search for a %[1]s

		Usage:
		• %[1]s %[2]s %[3]s %[4]s
		• %[1]s %[2]s %[3]s --query %[4]s
				`, cst.NounConfig, cst.NounAuthProvider, cst.Search, cst.ExampleAuthProviderName),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Query):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Query, Shorthand: "q", Usage: fmt.Sprintf("Filter %s of items to fetch (required)", strings.Title(cst.Query))}), false},
			preds.LongFlag(cst.Limit):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Limit, Shorthand: "l", Usage: fmt.Sprint("Maximum number of results per cursor (optional)")}), false},
			preds.LongFlag(cst.Cursor): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Cursor, Usage: fmt.Sprint("Next cursor for additional results (optional)")}), false},
		},
		MinNumberArgs: 0,
	})
}

func (p AuthProvider) handleAuthProviderReadCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	path, status := getAuthProviderParams(args)
	if status != 0 {
		return status
	}
	ver := viper.GetString(cst.Version)
	if strings.TrimSpace(ver) != "" {
		path = fmt.Sprint(path, "/", cst.Version, "/", ver)
	}

	baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")
	uri := utils.CreateResourceURI(baseType, path, "", true, nil, false)

	data, err = p.request.DoRequest("GET", uri, nil)

	outClient := p.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (p AuthProvider) handleAuthProviderDeleteCmd(args []string) int {
	var err *errors.ApiError
	var resp []byte
	path, status := getAuthProviderParams(args)
	if status != 0 {
		return status
	}
	force := viper.GetBool(cst.Force)

	baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")
	query := map[string]string{"force": strconv.FormatBool(force)}
	uri := utils.CreateResourceURI(baseType, path, "", true, query, false)

	resp, err = p.request.DoRequest("DELETE", uri, nil)

	outClient := p.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func (p AuthProvider) handleAuthProviderRestoreCmd(args []string) int {
	var err *errors.ApiError
	var resp []byte
	path, status := getAuthProviderParams(args)
	if status != 0 {
		return status
	}

	baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")
	uri := utils.CreateResourceURI(baseType, path, "", true, nil, false)
	uri += "/restore"

	resp, err = p.request.DoRequest("PUT", uri, nil)

	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}
	p.outClient.WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func (p AuthProvider) handleAuthProviderUpsert(args []string) int {
	params := map[string]string{}
	var resp []byte
	var err *errors.ApiError

	name, status := getAuthProviderParams(args)
	if status != 0 {
		return status
	}
	params[cst.DataName] = name

	data := viper.GetString(cst.Data)
	if data == "" {
		params[cst.DataName] = name
		params[cst.DataType] = viper.GetString(cst.DataType)
		params[cst.DataAccountID] = viper.GetString(cst.DataAccountID)
		params[cst.DataTenantID] = viper.GetString(cst.DataTenantID)
	}

	var postData interface{}
	if data != "" {
		if err := json.Unmarshal([]byte(data), &postData); err != nil {
			postData = data
		}
	} else {
		postData = map[string]interface{}{
			"name": params[cst.DataName],
			"type": params[cst.DataType],
			"properties": map[string]string{
				"accountId": params[cst.DataAccountID],
				"tenantId":  params[cst.DataTenantID],
			},
		}
	}

	baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")
	var uri string
	reqMethod := strings.ToLower(viper.GetString(cst.LastCommandKey))
	if reqMethod == cst.Create {
		reqMethod = "POST"
		uri = utils.CreateResourceURI(baseType, "", "", true, nil, false)
	} else {
		reqMethod = "PUT"
		uri = utils.CreateResourceURI(baseType, params[cst.DataName], "", true, nil, false)
	}
	resp, err = p.request.DoRequest(reqMethod, uri, postData)

	outClient := p.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func (p AuthProvider) handleAuthProviderRollbackCmd(args []string) int {
	var apiError *errors.ApiError
	var resp []byte
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}
	baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")

	path, status := getAuthProviderParams(args)
	if status != 0 {
		return status
	}
	version := viper.GetString(cst.Version)

	// If version is not provided, get the current auth-provider item and parse the version from it.
	// Submit a request for a version that's previous relative to the one found.
	if version == "" {
		uri := utils.CreateResourceURI(baseType, path, "", true, nil, false)
		resp, apiError = p.request.DoRequest("GET", uri, nil)
		if apiError != nil {
			p.outClient.WriteResponse(resp, apiError)
			return utils.GetExecStatus(apiError)
		}

		v, err := utils.GetPreviousVersion(resp)
		if err != nil {
			p.outClient.Fail(err)
			return 1
		}
		version = v
	}

	if strings.TrimSpace(version) != "" {
		path = fmt.Sprint(path, "/rollback/", version)
	}
	uri := utils.CreateResourceURI(baseType, path, "", true, nil, false)
	resp, apiError = p.request.DoRequest("PUT", uri, nil)

	p.outClient.WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func (p AuthProvider) handleAuthProviderSearchCommand(args []string) int {
	baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")
	data, err := handleSearch(args, baseType, p.request)
	outClient := p.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func getAuthProviderParams(args []string) (name string, status int) {
	status = 0
	name = viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 {
		name = args[0]
	}
	if name == "" {
		status = cli.RunResultHelp
	}
	return name, status
}
