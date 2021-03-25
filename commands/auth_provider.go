package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"thy/constants"
	cst "thy/constants"
	"thy/errors"
	"thy/format"
	"thy/paths"
	preds "thy/predictors"
	"thy/requests"
	"thy/store"
	"thy/utils"
	"thy/version"

	"github.com/mitchellh/mapstructure"
	"github.com/posener/complete"
	"github.com/spf13/viper"
	"github.com/thycotic-rd/cli"
)

type AuthProvider struct {
	request   requests.Client
	outClient format.OutClient
	edit      func([]byte, dataFunc, *errors.ApiError, bool) ([]byte, *errors.ApiError)
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
			return AuthProvider{requests.NewHttpClient(), nil, nil}.handleAuthProviderReadCmd(args)
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
		SynopsisText: fmt.Sprintf("%s %s %s (<name> | --name|-n) (--type) ((--data|-d) | --aws-account-id | --azure-tenant-id | --gcp-project-id)", cst.NounConfig, cst.NounAuthProvider, cst.Create),
		HelpText: fmt.Sprintf(`Add %[1]s provider

Usage:
  • %[1]s %[2]s %[4]s %[3]s --aws-account-id 11652944433808  --type aws
  • %[1]s %[2]s %[4]s --name azure-prod --azure-tenant-id 164543 --type azure
  • %[1]s %[2]s %[4]s --name GCP-prod --gcp-project-id test-proj --type gcp
  • %[1]s %[2]s %[4]s --data %[5]s

 %[6]s
		`, cst.NounConfig, cst.NounAuthProvider, cst.ExampleAuthProviderName, cst.Create, cst.ExampleDataPath, cst.GCPNote),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Data):          cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("%s to be stored in an auth provider. Prefix with '@' to denote filepath", strings.Title(cst.Data))}), false},
			preds.LongFlag(cst.DataType):      cli.PredictorWrapper{preds.AuthTypePredictor{}, preds.NewFlagValue(preds.Params{Name: cst.DataType, Usage: fmt.Sprintf("Auth provider type (azure,aws,gcp,thycoticone)")}), false},
			preds.LongFlag(cst.DataName):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataName, Shorthand: "n", Usage: fmt.Sprintf("Auth provider friendly name")}), false},
			preds.LongFlag(cst.DataTenantID):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataTenantID, Usage: fmt.Sprintf("Azure Tenant ID")}), false},
			preds.LongFlag(cst.DataAccountID): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataAccountID, Usage: fmt.Sprintf("AWS Account ID")}), false}, preds.LongFlag(cst.DataAccountID): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataAccountID, Usage: fmt.Sprintf("AWS Account ID")}), false},
			preds.LongFlag(cst.DataProjectID):           cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataProjectID, Usage: fmt.Sprintf("GCP Project ID")}), false},
			preds.LongFlag(cst.ThyOneAuthClientBaseUri): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.ThyOneAuthClientBaseUri, Usage: fmt.Sprintf("Thycotic One base URI")}), false},
			preds.LongFlag(cst.ThyOneAuthClientID):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.ThyOneAuthClientID, Usage: fmt.Sprintf("Thycotic One client ID")}), false},
			preds.LongFlag(cst.ThyOneAuthClientSecret):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.ThyOneAuthClientSecret, Usage: fmt.Sprintf("Thycotic One client secret")}), false},
			preds.LongFlag(cst.SendWelcomeEmail):        cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.SendWelcomeEmail, Usage: fmt.Sprintf("Whether to send welcome email for thycotic-one users linked to the auth provider (true or false)")}), false},
		},
		MinNumberArgs: 0,
	})
}

func GetAuthProviderUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.Config, cst.NounAuthProvider, cst.Update},
		RunFunc:      AuthProvider{request: requests.NewHttpClient(), outClient: nil}.handleAuthProviderUpsert,
		SynopsisText: fmt.Sprintf("%s %s %s (<name> | --name|-n) (--type) ((--data|-d) | --aws-account-id | --azure-tenant-id | --gcp-project-id)", cst.NounConfig, cst.NounAuthProvider, cst.Update),
		HelpText: fmt.Sprintf(`Update %[1]s properties

Usage:
  • %[1]s %[2]s %[4]s %[3]s --aws-account-id 11652944433808  --type aws
  • %[1]s %[2]s %[4]s --name azure-prod --azure-tenant-id 164543 --type azure
  • %[1]s %[2]s %[4]s --name GCP-prod --gcp-project-id test-proj --type gcp
  • %[1]s %[2]s %[4]s --data %[5]s
		`, cst.NounConfig, cst.NounAuthProvider, cst.ExampleAuthProviderName, cst.Update, cst.ExampleDataPath),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Data):                    cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("%s to be stored in an auth provider. Prefix with '@' to denote filepath", strings.Title(cst.Data))}), false},
			preds.LongFlag(cst.DataType):                cli.PredictorWrapper{preds.AuthTypePredictor{}, preds.NewFlagValue(preds.Params{Name: cst.DataType, Usage: fmt.Sprintf("Auth provider type (azure,aws,gcp,thycoticone)")}), false},
			preds.LongFlag(cst.DataName):                cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataName, Shorthand: "n", Usage: fmt.Sprintf("Auth provider friendly name")}), false},
			preds.LongFlag(cst.DataTenantID):            cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataTenantID, Usage: fmt.Sprintf("Azure Tenant ID")}), false},
			preds.LongFlag(cst.DataAccountID):           cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataAccountID, Usage: fmt.Sprintf("AWS Account ID")}), false},
			preds.LongFlag(cst.DataProjectID):           cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataProjectID, Usage: fmt.Sprintf("GCP Project ID")}), false},
			preds.LongFlag(cst.ThyOneAuthClientBaseUri): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.ThyOneAuthClientBaseUri, Usage: fmt.Sprintf("Thycotic One base URI")}), false},
			preds.LongFlag(cst.ThyOneAuthClientID):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.ThyOneAuthClientID, Usage: fmt.Sprintf("Thycotic One client ID")}), false},
			preds.LongFlag(cst.ThyOneAuthClientSecret):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.ThyOneAuthClientSecret, Usage: fmt.Sprintf("Thycotic One client secret")}), false},
			preds.LongFlag(cst.SendWelcomeEmail):        cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.SendWelcomeEmail, Usage: fmt.Sprintf("Whether to send welcome email for thycotic-one users linked to the auth provider (true or false)")}), false},
		},
		MinNumberArgs: 0,
	})
}

func GetAuthProviderEditCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.Config, cst.NounAuthProvider, cst.Edit},
		RunFunc:      AuthProvider{request: requests.NewHttpClient(), outClient: nil, edit: EditData}.handleAuthProviderEdit,
		SynopsisText: fmt.Sprintf("%s %s %s (<name> | --name|-n)", cst.NounConfig, cst.NounAuthProvider, cst.Edit),
		HelpText: fmt.Sprintf(`Edit an auth provider

Usage:
   • %[1]s %[2]s %[4]s %[3]s
   • %[1]s %[2]s %[4]s --name %[3]s
		`, cst.NounConfig, cst.NounAuthProvider, cst.ExampleAuthProviderName, cst.Edit),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataName, Shorthand: "n", Usage: fmt.Sprintf("Target %s to an %s", cst.Path, cst.NounAuthProvider)}), false},
		},
		MinNumberArgs: 1,
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
			preds.LongFlag(cst.Cursor): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Cursor, Usage: constants.CursorHelpMessage}), false},
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
	path = paths.ProcessResource(path)
	ver := viper.GetString(cst.Version)
	if strings.TrimSpace(ver) != "" {
		path = fmt.Sprint(path, "/", cst.Version, "/", ver)
	}

	uri := p.makeReadUrl(path)
	data, err = p.request.DoRequest("GET", uri, nil)

	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}
	p.outClient.WriteResponse(data, err)
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
	uri := paths.CreateResourceURI(baseType, path, "", true, query, false)

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
	uri := paths.CreateResourceURI(baseType, path, "/restore", true, nil, false)

	resp, err = p.request.DoRequest("PUT", uri, nil)

	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}
	p.outClient.WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func (p AuthProvider) handleAuthProviderUpsertWorkflow(args []string) int {
	isUpdate := viper.GetString(cst.LastCommandKey) == cst.Update
	var params authProvider
	ui := &PasswordUi{
		cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
	}

	name, err := getStringAndValidate(
		ui, "Auth provider name:", false, nil, false, false)
	if err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params.Name = paths.ProcessResource(name)
	}

	if isUpdate {
		model, err := p.readAuthProvider(name)
		if err != nil {
			ui.Error(err.Error())
			return utils.GetExecStatus(err)
		} else {
			params.Type = model.Type
		}
	} else {
		if providerType, err := getStringAndValidate(
			ui, "Auth provider type:", true, []option{
				{"aws", "AWS"},
				{"azure", "Azure"},
				{"gcp", "GCP"},
				{cst.ThyOne, "Thycotic One"},
			}, false, false); err != nil {
			ui.Error(err.Error())
			return utils.GetExecStatus(err)
		} else {
			params.Type = providerType
		}
	}

	switch params.Type {
	case "aws":
		if awsAccountId, err := getStringAndValidate(
			ui, "AWS account ID:", false, nil, false, false); err != nil {
			ui.Error(err.Error())
			return utils.GetExecStatus(err)
		} else {
			params.Properties.AccountID = awsAccountId
		}
	case "azure":
		if azureTenantId, err := getStringAndValidate(
			ui, "Azure tenant ID:", false, nil, false, false); err != nil {
			ui.Error(err.Error())
			return utils.GetExecStatus(err)
		} else {
			params.Properties.TenantID = azureTenantId
		}
	case "gcp":
		if gcpProjectID, err := getStringAndValidate(
			ui, "GCP project ID:", false, nil, false, false); err == nil && gcpProjectID != "" {
			params.Properties.ProjectID = gcpProjectID
		} else if dataPath, err := getStringAndValidate(
			ui, "Path to data file with provider properties:", false, nil, false, false); err != nil {
			ui.Error(err.Error())
			return utils.GetExecStatus(err)
		} else {
			data, err := store.ReadFile(dataPath)
			if err != nil {
				ui.Error(err.Error())
				return utils.GetExecStatus(err)
			}
			var dataMap map[string]interface{}
			err = json.Unmarshal([]byte(data), &dataMap)
			if err != nil {
				ui.Error("Failed to parse properties file.")
				return utils.GetExecStatus(err)
			}
			var props Properties
			if properties, ok := dataMap["properties"]; !ok {
				ui.Error("No properties in data file.")
				return utils.GetExecStatus(err)
			} else {
				mapstructure.Decode(properties, &props)
			}
			params.Properties = props
		}
	case cst.ThyOne:
		if url, err := getStringAndValidate(
			ui, "Base URL:", false, nil, false, false); err != nil {
			ui.Error(err.Error())
			return utils.GetExecStatus(err)
		} else {
			params.Properties.BaseURI = url
		}
		if clientId, err := getStringAndValidate(
			ui, "Client ID:", false, nil, false, false); err != nil {
			ui.Error(err.Error())
			return utils.GetExecStatus(err)
		} else {
			params.Properties.ClientID = clientId
		}
		if clientSecret, err := getStringAndValidate(
			ui, "Client secret:", false, nil, false, false); err != nil {
			ui.Error(err.Error())
			return utils.GetExecStatus(err)
		} else {
			params.Properties.ClientSecret = clientSecret
		}
		if sendWelcomeEmail, err := getStringAndValidate(
			ui, "Send welcome email (true or false):", true, nil, false, false); err != nil {
			ui.Error(err.Error())
			return utils.GetExecStatus(err)
		} else {
			sendWelcomeEmail, parseErr := strconv.ParseBool(sendWelcomeEmail)
			if parseErr == nil {
				params.Properties.SendWelcomeEmail = &sendWelcomeEmail
			}
		}
	default:
		ui.Error("Unsupported auth provider type.")
		return 1
	}

	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}

	resp, apiErr := p.submitAuthProvider(params)
	p.outClient.WriteResponse(resp, apiErr)
	return utils.GetExecStatus(apiErr)
}

func (p AuthProvider) handleAuthProviderUpsert(args []string) int {
	if OnlyGlobalArgs(args) {
		return p.handleAuthProviderUpsertWorkflow(args)
	}
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}
	params := map[string]string{}
	var resp []byte
	var err *errors.ApiError

	name, status := getAuthProviderParams(args)
	if status != 0 {
		return status
	}
	params[cst.DataName] = name

	data := viper.GetString(cst.Data)
	var model authProvider
	if data != "" {
		if err := json.Unmarshal([]byte(data), &model); err != nil {
			p.outClient.Fail(err)
			return utils.GetExecStatus(err)
		}
	} else {
		model = authProvider{
			Name: params[cst.DataName],
			Type: viper.GetString(cst.DataType),
			Properties: Properties{
				AccountID:    viper.GetString(cst.DataAccountID),
				TenantID:     viper.GetString(cst.DataTenantID),
				ProjectID:    viper.GetString(cst.DataProjectID),
				BaseURI:      viper.GetString(cst.ThyOneAuthClientBaseUri),
				ClientID:     viper.GetString(cst.ThyOneAuthClientID),
				ClientSecret: viper.GetString(cst.ThyOneAuthClientSecret),
			},
		}
	}
	sendWelcomeEmail, parseErr := strconv.ParseBool(viper.GetString(cst.SendWelcomeEmail))
	if parseErr == nil && model.Type == cst.ThyOne {
		model.Properties.SendWelcomeEmail = &sendWelcomeEmail
	}

	resp, apiErr := p.submitAuthProvider(model)
	p.outClient.WriteResponse(resp, apiErr)
	return utils.GetExecStatus(err)
}

func (p AuthProvider) submitAuthProvider(model authProvider) ([]byte, *errors.ApiError) {
	baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")
	var uri string
	reqMethod := strings.ToLower(viper.GetString(cst.LastCommandKey))
	if reqMethod == cst.Create {
		reqMethod = "POST"
		uri = paths.CreateResourceURI(baseType, "", "", true, nil, false)
	} else {
		reqMethod = "PUT"
		uri = paths.CreateResourceURI(baseType, model.Name, "", true, nil, false)
	}
	return p.request.DoRequest(reqMethod, uri, model)
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
		uri := paths.CreateResourceURI(baseType, path, "", true, nil, false)
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
	uri := paths.CreateResourceURI(baseType, path, "", true, nil, false)
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

func (p AuthProvider) handleAuthProviderEdit(args []string) int {
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}

	var err *errors.ApiError
	var resp []byte

	path, status := getAuthProviderParams(args)
	if status != 0 {
		return status
	}
	baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")
	uri := paths.CreateResourceURI(baseType, paths.ProcessResource(path), "", true, nil, false)

	resp, err = p.request.DoRequest("GET", uri, nil)
	if err != nil {
		p.outClient.WriteResponse(resp, err)
		return utils.GetExecStatus(err)
	}

	saveFunc := dataFunc(func(data []byte) (resp []byte, err *errors.ApiError) {
		var model authProvider
		if mErr := json.Unmarshal(data, &model); mErr != nil {
			return nil, errors.New(mErr).Grow("invalid format for auth provider")
		}
		_, err = p.request.DoRequest("PUT", uri, &model)
		return nil, err
	})
	resp, err = p.edit(resp, saveFunc, nil, false)
	p.outClient.WriteResponse(resp, err)
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

func (p AuthProvider) readAuthProvider(name string) (*authProvider, error) {
	uri := p.makeReadUrl(name)
	data, apiError := p.request.DoRequest("GET", uri, nil)
	if apiError != nil {
		return nil, apiError
	}
	var m authProvider
	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (p AuthProvider) makeReadUrl(name string) string {
	baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")
	return paths.CreateResourceURI(baseType, name, "", true, nil, false)
}

type authProvider struct {
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	Properties Properties `json:"properties"`
}

type Properties struct {
	AccountID        string `json:"accountId,omitempty"`
	TenantID         string `json:"tenantId,omitempty"`
	ProjectID        string `json:"projectId,omitempty"`
	Default          bool   `json:"default,omitempty"`
	ClientEmail      string `json:"clientEmail,omitempty"`
	PrivateKey       string `json:"privateKey,omitempty"`
	PrivateKeyID     string `json:"privateKeyId,omitempty"`
	TokenURI         string `json:"tokenUri,omitempty"`
	ClientID         string `json:"clientId,omitempty"`
	ClientSecret     string `json:"clientSecret,omitempty"`
	Type             string `json:"type,omitempty"`
	BaseURI          string `json:"baseUri,omitempty"`
	UsernameClaim    string `json:"usernameClaim,omitempty"`
	SendWelcomeEmail *bool  `json:"sendWelcomeEmail,omitempty"`
}
