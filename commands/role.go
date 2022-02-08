package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	cst "thy/constants"
	"thy/errors"
	"thy/format"
	"thy/internal/prompt"
	"thy/paths"
	preds "thy/predictors"
	"thy/requests"
	"thy/utils"

	"github.com/mitchellh/mapstructure"
	"github.com/posener/complete"
	"github.com/spf13/viper"
	"github.com/thycotic-rd/cli"
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
			preds.LongFlag(cst.Cursor): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Cursor, Usage: cst.CursorHelpMessage}), false},
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
		SynopsisText: fmt.Sprintf("%s %s (<name> | --name|-n) --desc", cst.NounRole, cst.Update),
		HelpText: fmt.Sprintf(`%[4]s a %[1]s in %[2]s

Usage:
   • %[1]s %[4]s --%[5]s %[6]s --desc "msa for prod gcp"
		`, cst.NounRole, cst.ProductName, cst.ExamplePath, cst.Update, cst.DataName, cst.ExampleRoleName),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataName):        cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s (required)", cst.NounRole)}), false},
			preds.LongFlag(cst.DataDescription): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataDescription, Usage: fmt.Sprintf("Description of the %s ", cst.NounRole)}), false},
		},
		MinNumberArgs: 0,
	})
}

func GetRoleCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounRole, cst.Create},
		RunFunc:      Roles{requests.NewHttpClient(), nil}.handleRoleUpsertCmd,
		SynopsisText: fmt.Sprintf("%s %s (<name> | --name|-n) --provider --external-id --desc", cst.NounRole, cst.Create),
		HelpText: fmt.Sprintf(`%[4]s a %[1]s in %[2]s

Usage:
   • %[1]s %[4]s --%[5]s %[6]s --external-id msa-1@happy-emu-172.iam.gsa.com --provider ProdGcp --desc "msa for prod gcp"
		`, cst.NounRole, cst.ProductName, cst.ExamplePath, cst.Create, cst.DataName, cst.ExampleRoleName),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataName):        cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s (required)", cst.NounRole)}), false},
			preds.LongFlag(cst.DataDescription): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataDescription, Usage: fmt.Sprintf("Description of the %s ", cst.NounRole)}), false},
			preds.LongFlag(cst.DataExternalID):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataExternalID, Usage: fmt.Sprintf("External Id for the %s", cst.NounRole)}), false},
			preds.LongFlag(cst.DataProvider):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataProvider, Usage: fmt.Sprintf("Provider for the %s ", cst.NounRole)}), false},
		},
		MinNumberArgs: 0,
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
		name = paths.ProcessResource(name)
		version := viper.GetString(cst.Version)
		if strings.TrimSpace(version) != "" {
			name = fmt.Sprint(name, "/", cst.Version, "/", version)
		}
		uri := paths.CreateResourceURI(cst.NounRole, name, "", true, nil, true)
		data, err = r.request.DoRequest(http.MethodGet, uri, nil)
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
		uri := paths.CreateResourceURI(cst.NounRole, "", "", false, queryParams, true)
		data, err = r.request.DoRequest(http.MethodGet, uri, nil)
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
		uri := paths.CreateResourceURI(cst.NounRole, paths.ProcessResource(name), "", true, query, true)
		data, err = r.request.DoRequest(http.MethodDelete, uri, nil)
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
		uri := paths.CreateResourceURI(cst.NounRole, paths.ProcessResource(name), "/restore", true, nil, true)
		data, err = r.request.DoRequest(http.MethodPut, uri, nil)
	}

	r.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (r Roles) handleRoleUpsertCmd(args []string) int {
	if OnlyGlobalArgs(args) {
		return r.handleRoleWorkflow(args)
	}
	if r.outClient == nil {
		r.outClient = format.NewDefaultOutClient()
	}
	name := viper.GetString(cst.DataName)
	if name == "" {
		err := errors.NewS("error: must specify " + cst.DataName)
		r.outClient.WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}
	isUpdate := viper.GetString(cst.LastCommandKey) == cst.Update
	role := Role{
		Description: viper.GetString(cst.DataDescription),
		Name:        name,
	}
	if !isUpdate {
		role.ExternalID = viper.GetString(cst.DataExternalID)
		role.Provider = viper.GetString(cst.DataProvider)
		if (role.Provider != "" && role.ExternalID == "") || (role.Provider == "" && role.ExternalID != "") {
			err := errors.NewS("error: must specify both provider and external ID for third-party roles")
			r.outClient.WriteResponse(nil, err)
			return utils.GetExecStatus(err)
		}
	}

	data, err := r.submitRole(name, role, isUpdate)
	r.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (r Roles) handleRoleWorkflow(args []string) int {
	ui := &PasswordUi{
		cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
	}
	if r.outClient == nil {
		r.outClient = format.NewDefaultOutClient()
	}
	params := make(map[string]string)
	isUpdate := viper.GetString(cst.LastCommandKey) == cst.Update
	if resp, err := prompt.Ask(ui, "Role name:"); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params["name"] = resp
	}

	if resp, err := prompt.AskDefault(ui, "Description of the role:", ""); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params["description"] = resp
	}

	if !isUpdate {
		baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")

		// If we were able to obtain a list of auth providers, proceed with selection,
		// otherwise role does not have optional provider or external id
		if data, err := handleSearch(nil, baseType, r.request); err == nil {
			providers, parseErr := parseAuthProviders(data)
			if parseErr != nil {
				r.outClient.FailS("Failed to parse out available auth providers.")
				return utils.GetExecStatus(parseErr)
			}

			options := []prompt.Option{}
			for _, p := range providers {
				// Skip thycoticone - roles cannot have it as a provider.
				if p.Type == cst.ThyOne {
					continue
				}
				v := fmt.Sprintf("%s:%s", p.Name, p.Type)
				options = append(options, prompt.Option{v, strings.Replace(v, ":", " - ", 1)})
			}
			if len(options) > 0 {
				var providerName string
				if resp, err := prompt.Choose(ui, "Provider:", prompt.Option{"local", "local"}, options...); err != nil {
					ui.Error(err.Error())
					return 1
				} else {
					providerName = resp
				}
				if p := strings.Split(providerName, ":"); p[0] != "local" {
					if resp, err := prompt.Ask(ui, "External ID:"); err != nil {
						ui.Error(err.Error())
						return 1
					} else {
						params["provider"] = strings.Split(providerName, ":")[0]
						params["externalId"] = resp
					}
				}
			}
		} else if err.HttpResponse().StatusCode != 403 {
			r.outClient.FailS(err.Error())
			return utils.GetExecStatus(err)
		} else {
			if resp, err := prompt.AskDefault(ui, "Provider:", ""); err != nil {
				ui.Error(err.Error())
				return 1
			} else {
				params["provider"] = resp
			}

			if resp, err := prompt.AskDefault(ui, "External ID:", ""); err != nil {
				ui.Error(err.Error())
				return 1
			} else {
				params["externalId"] = resp
			}
		}
	}

	var role Role
	mapstructure.Decode(params, &role)
	resp, apiError := r.submitRole(params["name"], role, isUpdate)
	r.outClient.WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func (r Roles) submitRole(path string, role Role, update bool) ([]byte, *errors.ApiError) {
	if update {
		uri := paths.CreateResourceURI(cst.NounRole, path, "", true, nil, true)
		return r.request.DoRequest(http.MethodPut, uri, &role)

	} else {
		uri := paths.CreateResourceURI(cst.NounRole, "", "", true, nil, true)
		return r.request.DoRequest(http.MethodPost, uri, &role)
	}
}

type Role struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ExternalID  string `json:"externalId"`
	Provider    string `json:"provider"`
}
