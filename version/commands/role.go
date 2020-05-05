package cmd

import (
	"fmt"
	"strconv"
	"strings"

	cst "thy/constants"
	"thy/errors"
	"thy/format"
	preds "thy/predictors"
	"thy/requests"
	"thy/utils"

	"github.com/posener/complete"
	"github.com/thycotic-rd/cli"
	"github.com/thycotic-rd/viper"
)

type Roles struct {
	request   requests.Client
	outClient format.OutClient
}

func GetNoDataOpRoleWrappers() cli.PredictorWrappers {
	return cli.PredictorWrappers{
		preds.LongFlag(cst.DataName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s ", cst.NounRole)}), false},
		preds.LongFlag(cst.Version):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Version, Shorthand: "", Usage: "List the current and last (n) versions"}), false},
	}
}

func GetRoleCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounRole},
		RunFunc: func(args []string) int {
			name := viper.GetString(cst.DataName)
			if name == "" && len(args) > 0 {
				name = args[0]
			}
			if name == "" {
				return cli.RunResultHelp
			}
			return Roles{requests.NewHttpClient(), nil}.handleRoleReadCmd(args)
		},
		SynopsisText: "role (<name> | --name|-n)",
		HelpText: fmt.Sprintf(`Execute an action on a %[1]s in %[2]s

Usage:
   • %[1]s %[3]s 
   • %[1]s --name %[3]s 
		`, cst.NounRole, cst.ProductName, cst.ExampleRoleName),
		FlagsPredictor: GetNoDataOpRoleWrappers(),
		MinNumberArgs:  1,
	})
}

func GetRoleReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounRole, cst.Read},
		RunFunc:      Roles{requests.NewHttpClient(), nil}.handleRoleReadCmd,
		SynopsisText: fmt.Sprintf("%s %s (<name> | --name|-n)", cst.NounRole, cst.Read),
		HelpText: fmt.Sprintf(`Execute an action on a %[1]s in %[2]s

Usage:
   • %[1]s %[4]s %[3]s 
   • %[1]s %[4]s --name %[3]s 
   • %[1]s %[4]s --name %[3]s  --version 
		`, cst.NounRole, cst.ProductName, cst.ExampleRoleName, cst.Read),
		FlagsPredictor: GetNoDataOpRoleWrappers(),
		MinNumberArgs:  1,
	})
}

func GetRoleSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounRole, cst.Search},
		RunFunc:      Roles{requests.NewHttpClient(), nil}.handleRoleSearchCmd,
		SynopsisText: fmt.Sprintf("%s (<query> | --query)", cst.Search),
		HelpText: fmt.Sprintf(`Search for a %[2]s from %[3]s

		Usage:
		• role %[1]s %[4]s
		• role %[1]s --query %[4]s
				`, cst.Search, cst.NounRole, cst.ProductName, cst.ExampleUserSearch),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Query):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Query, Shorthand: "q", Usage: fmt.Sprintf("%s of %ss to fetch (required)", strings.Title(cst.Query), cst.NounRole)}), false},
			preds.LongFlag(cst.Limit):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Limit, Shorthand: "l", Usage: fmt.Sprint("Maximum number of results per cursor (optional)")}), false},
			preds.LongFlag(cst.Cursor): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Cursor, Usage: fmt.Sprint("Next cursor for additional results (optional)")}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetRoleDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounRole, cst.Delete},
		RunFunc:      Roles{requests.NewHttpClient(), nil}.handleRoleDeleteCmd,
		SynopsisText: fmt.Sprintf("%s %s (<name> | --name|-n)", cst.NounRole, cst.Delete),
		HelpText: fmt.Sprintf(`Execute an action on a %[1]s in %[2]s

Usage:
   • %[1]s %[4]s %[3]s 
   • %[1]s %[4]s --name %[3]s --force
		`, cst.NounRole, cst.ProductName, cst.ExampleRoleName, cst.Delete),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s ", cst.NounRole)}), false},
			preds.LongFlag(cst.Force):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Force, Shorthand: "", Usage: fmt.Sprintf("Immediately delete %s", cst.NounRole), Global: false, ValueType: "bool"}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetRoleRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounRole, cst.Restore},
		RunFunc:      Roles{requests.NewHttpClient(), nil}.handleRoleRestoreCmd,
		SynopsisText: fmt.Sprintf("%s %s (<name> | --name|-n)", cst.NounRole, cst.Restore),
		HelpText: fmt.Sprintf(`Restore a deleted %[1]s in %[2]s

Usage:
   • %[1]s %[4]s %[3]s
		`, cst.NounRole, cst.ProductName, cst.ExampleRoleName, cst.Restore),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s ", cst.NounRole)}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetRoleUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounRole, cst.Update},
		RunFunc:      Roles{requests.NewHttpClient(), nil}.handleRoleUpsertCmd,
		SynopsisText: fmt.Sprintf("%s %s (<name> | --name|-n) --provider --external-id --desc", cst.NounRole, cst.Update),
		HelpText: fmt.Sprintf(`%[4]s a %[1]s in %[2]s

Usage:
   • %[1]s %[4]s %[5]s --external-id msa-1@happy-emu-172.iam.gsa.com --provider ProdGcp --description "msa for prod gcp"
		`, cst.NounRole, cst.ProductName, cst.ExamplePath, cst.Update, cst.ExampleRoleName),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataName):        cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s ", cst.NounRole)}), false},
			preds.LongFlag(cst.DataDescription): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataDescription, Usage: fmt.Sprintf("Description of the %s ", cst.NounRole)}), false},
			preds.LongFlag(cst.DataExternalID):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataExternalID, Usage: fmt.Sprintf("External Id for the %s", cst.NounRole)}), false},
			preds.LongFlag(cst.DataProvider):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataProvider, Usage: fmt.Sprintf("Provider for the %s ", cst.NounRole)}), false},
		},
		MinNumberArgs: 2,
	})
}

func GetRoleCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounRole, cst.Create},
		RunFunc:      Roles{requests.NewHttpClient(), nil}.handleRoleUpsertCmd,
		SynopsisText: fmt.Sprintf("%s %s (<name> | --name|-n) --provider --external-id --desc", cst.NounRole, cst.Create),
		HelpText: fmt.Sprintf(`%[4]s a %[1]s in %[2]s

Usage:
   • %[1]s %[4]s %[5]s --external-id msa-1@happy-emu-172.iam.gsa.com --provider ProdGcp --description "msa for prod gcp"
		`, cst.NounRole, cst.ProductName, cst.ExamplePath, cst.Create, cst.ExampleRoleName),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataName):        cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s ", cst.NounRole)}), false},
			preds.LongFlag(cst.DataDescription): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataDescription, Usage: fmt.Sprintf("Description of the %s ", cst.NounRole)}), false},
			preds.LongFlag(cst.DataExternalID):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataExternalID, Usage: fmt.Sprintf("External Id for the %s", cst.NounRole)}), false},
			preds.LongFlag(cst.DataProvider):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataProvider, Usage: fmt.Sprintf("Provider for the %s ", cst.NounRole)}), false},
		},
		MinNumberArgs: 1,
	})
}

func (r Roles) handleRoleReadCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 {
		name = args[0]
	}
	if name == "" {
		err = errors.NewS("error: must specify " + cst.DataName)
	} else {
		version := viper.GetString(cst.Version)
		if strings.TrimSpace(version) != "" {
			name = fmt.Sprint(name, "/", cst.Version, "/", version)
		}
		uri := utils.CreateResourceURI(cst.NounRole, name, "", true, nil, true)
		data, err = r.request.DoRequest("GET", uri, nil)
	}

	outClient := r.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}

	outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (r Roles) handleRoleSearchCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	query := viper.GetString(cst.Query)
	limit := viper.GetString(cst.Limit)
	cursor := viper.GetString(cst.Cursor)
	if query == "" && len(args) > 0 {
		query = args[0]
	}
	if query == "" {
		err = errors.NewS("error: must specify " + cst.Query)
	} else {
		queryParams := map[string]string{
			cst.SearchKey: query,
			cst.Limit:     limit,
			cst.Cursor:    cursor,
		}
		uri := utils.CreateResourceURI(cst.NounRole, "", "", false, queryParams, true)
		data, err = r.request.DoRequest("GET", uri, nil)
	}
	outClient := r.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (r Roles) handleRoleDeleteCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	force := viper.GetBool(cst.Force)
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 {
		name = args[0]
	}
	if name == "" {
		err = errors.NewS("error: must specify " + cst.DataName)
	} else {
		query := map[string]string{"force": strconv.FormatBool(force)}
		uri := utils.CreateResourceURI(cst.NounRole, name, "", true, query, true)
		data, err = r.request.DoRequest("DELETE", uri, nil)
	}

	if r.outClient == nil {
		r.outClient = format.NewDefaultOutClient()
	}

	r.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (r Roles) handleRoleRestoreCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	if r.outClient == nil {
		r.outClient = format.NewDefaultOutClient()
	}
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 {
		name = args[0]
	}
	if name == "" {
		err = errors.NewS("error: must specify " + cst.DataName)
	} else {
		uri, err := utils.GetResourceURIFromResourcePath(cst.NounRole, name, "", "", true, nil)
		if err == nil {
			uri += "/restore"
			data, err = r.request.DoRequest("PUT", uri, nil)
		}
	}

	r.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (r Roles) handleRoleUpsertCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 {
		name = args[0]
	}
	if name == "" {
		return cli.RunResultHelp
	}
	role := Role{
		ExternalID:  viper.GetString(cst.DataExternalID),
		Provider:    viper.GetString(cst.DataProvider),
		Description: viper.GetString(cst.DataDescription),
		Name:        name,
	}
	reqMethod := strings.ToLower(viper.GetString(cst.LastCommandKey))
	var uri string
	if reqMethod == cst.Create {
		reqMethod = "POST"
		uri = utils.CreateResourceURI(cst.NounRole, "", "", true, nil, true)
	} else {
		reqMethod = "PUT"
		uri = utils.CreateResourceURI(cst.NounRole, name, "", true, nil, true)
	}
	data, err = r.request.DoRequest(reqMethod, uri, &role)

	outClient := r.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

type Role struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ExternalID  string `json:"externalId"`
	Provider    string `json:"provider"`
}
