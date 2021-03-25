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
	"thy/utils"

	"github.com/mitchellh/mapstructure"
	"github.com/posener/complete"

	"github.com/spf13/viper"
	"github.com/thycotic-rd/cli"
)

var (
	errMustSpecifyPassowrdOrDisplayname = errors.NewF("error: must specify %s or %s", cst.DataPassword, cst.DataDisplayname)
)

type User struct {
	request   requests.Client
	outClient format.OutClient
}

func GetDataOpUserWrappers(targetEntity string) cli.PredictorWrappers {
	return cli.PredictorWrappers{
		preds.LongFlag(cst.DataUsername):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataUsername, Usage: fmt.Sprintf("%s of %s to be updated (required)", strings.Title(cst.DataUsername), targetEntity)}), false},
		preds.LongFlag(cst.DataDisplayname): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataDisplayname, Usage: fmt.Sprintf("%s of %s to be updated", strings.Title(cst.DataDisplayname), targetEntity)}), false},
		preds.LongFlag(cst.DataPassword):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataPassword, Usage: fmt.Sprintf("%s of %s to be updated (required)", strings.Title(cst.Password), targetEntity)}), false},
		preds.LongFlag(cst.DataExternalID):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataExternalID, Usage: fmt.Sprintf("%s of %s to be updated", strings.Title(strings.Replace(cst.DataExternalID, ".", " ", -1)), targetEntity)}), false},
		preds.LongFlag(cst.DataProvider):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataProvider, Usage: fmt.Sprintf("External %s of %s to be updated", strings.Title(cst.DataProvider), targetEntity)}), false},
	}
}
func GetNoDataOpUserWrappers(targetEntity string) cli.PredictorWrappers {
	return cli.PredictorWrappers{
		preds.LongFlag(cst.DataUsername): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataUsername, Usage: fmt.Sprintf("%s of %s to fetch (required)", strings.Title(cst.DataUsername), targetEntity)}), false},
		preds.LongFlag(cst.Version):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Version, Shorthand: "", Usage: "List the current and last (n) versions"}), false},
	}
}

func GetUserCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounUser},
		RunFunc: func(args []string) int {
			userData := viper.GetString(cst.DataUsername)
			if userData == "" && len(args) > 0 {
				userData = args[0]
			}
			if userData == "" {
				return cli.RunResultHelp
			}
			return User{requests.NewHttpClient(), nil}.handleUserReadCmd(args)
		},
		SynopsisText: "user (<username> | --username)",
		HelpText: fmt.Sprintf(`Execute an action on a %s from %s

Usage:
   • user %[3]s
   • user --username %[3]s
		`, cst.NounUser, cst.ProductName, cst.ExampleUser),
		FlagsPredictor: GetNoDataOpUserWrappers(cst.NounUser),
		MinNumberArgs:  1,
	})
}

func GetUserReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUser, cst.Read},
		RunFunc:      User{requests.NewHttpClient(), nil}.handleUserReadCmd,
		SynopsisText: fmt.Sprintf("%s (<username> | --username)", cst.Read),
		HelpText: fmt.Sprintf(`Read a %[2]s from %[3]s

Usage:
   • user %[1]s %[4]s
   • user %[1]s --username %[4]s
		`, cst.Read, cst.NounUser, cst.ProductName, cst.ExampleUser),
		FlagsPredictor: GetNoDataOpUserWrappers(cst.NounUser),
		MinNumberArgs:  1,
	})
}

func GetUserSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUser, cst.Search},
		RunFunc:      User{requests.NewHttpClient(), nil}.handleUserSearchCmd,
		SynopsisText: fmt.Sprintf("%s (<query> | --query)", cst.Search),
		HelpText: fmt.Sprintf(`Search for a %[2]s from %[3]s

		Usage:
		• user %[1]s %[4]s
		• user %[1]s --query %[4]s
				`, cst.Search, cst.NounUser, cst.ProductName, cst.ExampleUserSearch),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Query):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Query, Shorthand: "q", Usage: fmt.Sprintf("%s of %ss to fetch (required)", strings.Title(cst.Query), cst.NounUser)}), false},
			preds.LongFlag(cst.Limit):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Limit, Shorthand: "l", Usage: fmt.Sprint("Maximum number of results per cursor (optional)")}), false},
			preds.LongFlag(cst.Cursor): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Cursor, Usage: constants.CursorHelpMessage}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetUserDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUser, cst.Delete},
		RunFunc:      User{requests.NewHttpClient(), nil}.handleUserDeleteCmd,
		SynopsisText: fmt.Sprintf("%s (<username> | --username)", cst.Delete),
		HelpText: fmt.Sprintf(`Delete a %[2]s from %[3]s

Usage:
   • user %[1]s %[4]s
   • user %[1]s --username %[4]s --force
		`, cst.Delete, cst.NounUser, cst.ProductName, cst.ExampleUser),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataUsername): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataUsername, Usage: fmt.Sprintf("%s of %s to fetch (required)", strings.Title(cst.DataUsername), cst.NounUser)}), false},
			preds.LongFlag(cst.Force):        cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Force, Shorthand: "", Usage: fmt.Sprintf("Immediately delete %s", cst.NounUser), Global: false, ValueType: "bool"}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetUserRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUser, cst.Read},
		RunFunc:      User{requests.NewHttpClient(), nil}.handleUserRestoreCmd,
		SynopsisText: fmt.Sprintf("%s %s (<username> | --username)", cst.NounUser, cst.Restore),
		HelpText: fmt.Sprintf(`Restore a deleted %[2]s in %[3]s
Usage:
	• user %[1]s %[4]s
	• user %[1]s --username %[4]s
				`, cst.Restore, cst.NounUser, cst.ProductName, cst.ExamplePath),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataUsername): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataUsername, Usage: fmt.Sprintf("%s of %s to fetch (required)", strings.Title(cst.DataUsername), cst.NounUser)}), false},
		},
		ArgsPredictorFunc: preds.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
	})
}

func GetUserCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUser, cst.Create},
		RunFunc:      User{requests.NewHttpClient(), nil}.handleUserCreateCmd,
		SynopsisText: fmt.Sprintf("%s (<username> <password> | --username --password)", cst.Create),
		HelpText: fmt.Sprintf(`Create a %[2]s in %[3]s

Usage:
   • user %[1]s --username %[4]s --password %[5]s
   • user %[1]s --username %[4]s --external-id svc1@project1.iam.gserviceaccount.com --provider project1.gcloud --password %[5]s
		`, cst.Create, cst.NounUser, cst.ProductName, cst.ExampleUser, cst.ExamplePassword),
		FlagsPredictor: GetDataOpUserWrappers(cst.NounUser),
		MinNumberArgs:  0,
	})
}

func GetUserUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUser, cst.Update},
		RunFunc:      User{requests.NewHttpClient(), nil}.handleUserUpdateCmd,
		SynopsisText: fmt.Sprintf("%s (<username> <password> | (--username) --password)", cst.Update),
		HelpText: fmt.Sprintf(`Update a %[2]s's password in %[3]s

Usage:
   • user %[1]s --username %[4]s --password %[5]s
		`, cst.Update, cst.NounUser, cst.ProductName, cst.ExampleUser, cst.ExamplePassword),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataPassword):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataPassword, Usage: fmt.Sprintf("%s of %s to be updated (required)", strings.Title(cst.Password), cst.NounUser)}), false},
			preds.LongFlag(cst.DataUsername):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataUsername, Usage: fmt.Sprintf("%s of %s to be updated (required)", strings.Title(cst.DataUsername), cst.NounUser)}), false},
			preds.LongFlag(cst.DataDisplayname): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataDisplayname, Usage: fmt.Sprintf("%s of %s to be updated", strings.Title(cst.DataDisplayname), cst.NounUser)}), false},
		},
		MinNumberArgs: 0,
	})
}

func (u User) handleUserReadCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	userName := viper.GetString(cst.DataUsername)
	if userName == "" && len(args) > 0 {
		userName = args[0]
	}
	if userName == "" {
		err = errors.NewS("error: must specify " + cst.DataUsername)
	} else {
		userName = paths.ProcessResource(userName)
		version := viper.GetString(cst.Version)
		if strings.TrimSpace(version) != "" {
			userName = fmt.Sprint(userName, "/", cst.Version, "/", version)
		}
		uri := paths.CreateResourceURI(cst.NounUser, userName, "", true, nil, true)
		data, err = u.request.DoRequest("GET", uri, nil)
	}

	outClient := u.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}

	outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (u User) handleUserSearchCmd(args []string) int {
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
		uri := paths.CreateResourceURI(cst.NounUser, "", "", false, queryParams, true)
		data, err = u.request.DoRequest("GET", uri, nil)
	}
	outClient := u.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (u User) handleUserDeleteCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	userName := viper.GetString(cst.DataUsername)
	force := viper.GetBool(cst.Force)
	if userName == "" && len(args) > 0 {
		userName = args[0]
	}
	if userName == "" {
		err = errors.NewS("error: must specify " + cst.DataUsername)
	} else {
		query := map[string]string{"force": strconv.FormatBool(force)}
		uri := paths.CreateResourceURI(cst.NounUser, paths.ProcessResource(userName), "", true, query, true)
		data, err = u.request.DoRequest("DELETE", uri, nil)
	}
	if u.outClient == nil {
		u.outClient = format.NewDefaultOutClient()
	}

	u.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (u User) handleUserRestoreCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	if u.outClient == nil {
		u.outClient = format.NewDefaultOutClient()
	}
	userName := viper.GetString(cst.DataUsername)
	if userName == "" && len(args) > 0 {
		userName = args[0]
	}
	if userName == "" {
		err = errors.NewS("error: must specify " + cst.DataUsername)
	} else {
		uri := paths.CreateResourceURI(cst.NounUser, paths.ProcessResource(userName), "/restore", true, nil, true)
		data, err = u.request.DoRequest("PUT", uri, nil)
	}

	u.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (u User) handleUserCreateCmd(args []string) int {
	if OnlyGlobalArgs(args) {
		return u.handleUserWorkflow(args)
	}
	if u.outClient == nil {
		u.outClient = format.NewDefaultOutClient()
	}
	userName := viper.GetString(cst.DataUsername)
	if userName == "" {
		err := errors.NewS("error: must specify " + cst.DataUsername)
		u.outClient.WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}
	provider := viper.GetString(cst.DataProvider)
	externalID := viper.GetString(cst.DataExternalID)

	isUserLocal := provider == "" && externalID == ""
	password := viper.GetString(cst.DataPassword)
	if password == "" && isUserLocal {
		err := errors.NewS("error: must specify password for local users")
		u.outClient.WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	displayName := viper.GetString(cst.DataDisplayname)
	data := map[string]string{
		"name":        userName,
		"username":    userName,
		"displayName": displayName,
	}

	if password != "" && isUserLocal {
		data["password"] = password
	} else {
		data["provider"] = provider
		data["externalId"] = externalID
	}

	resp, apiError := u.submitUser("", data, false)
	u.outClient.WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func (u User) handleUserUpdateCmd(args []string) int {
	if OnlyGlobalArgs(args) {
		return u.handleUserWorkflow(args)
	}
	if u.outClient == nil {
		u.outClient = format.NewDefaultOutClient()
	}
	userName := viper.GetString(cst.DataUsername)
	if userName == "" {
		err := errors.NewS("error: must specify " + cst.DataUsername)
		u.outClient.WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}
	passData := viper.GetString(cst.DataPassword)
	displayName := viper.GetString(cst.DataDisplayname)
	if passData == "" && displayName == "" {
		err := errMustSpecifyPassowrdOrDisplayname
		u.outClient.WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data := map[string]string{
		"password":    passData,
		"displayName": displayName,
	}

	resp, apiError := u.submitUser(userName, data, true)
	u.outClient.WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func (u User) readUser(name string) (*userModel, error) {
	uri := paths.CreateResourceURI(cst.NounUser, name, "", true, nil, true)
	data, apiError := u.request.DoRequest("GET", uri, nil)
	if apiError != nil {
		return nil, apiError
	}
	var um userModel
	err := json.Unmarshal(data, &um)
	if err != nil {
		return nil, err
	}
	return &um, nil
}

func (u User) handleUserWorkflow(args []string) int {
	ui := &PasswordUi{
		cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
	}
	if u.outClient == nil {
		u.outClient = format.NewDefaultOutClient()
	}
	params := make(map[string]string)
	isUpdate := viper.GetString(cst.LastCommandKey) == cst.Update
	if resp, err := getStringAndValidate(ui, "Username:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params["username"] = resp
	}

	// Update only allows to change the password or display name, relevant only for local users (does not have a provider).
	if isUpdate {
		// Read the user, make sure it exists and is a local user (does not have a provider).
		existingUser, err := u.readUser(params["username"])
		if err != nil {
			u.outClient.Fail(err)
			return utils.GetExecStatus(err)
		}

		var passwordResp, displannameResp bool

		// Password
		if resp, err := getStringAndValidateDefault(
			ui, "Would you like to update the password [y/n] (default: n):", "n", true, nil, false, false); err != nil {
			ui.Error(err.Error())
			return utils.GetExecStatus(err)
		} else {
			passwordResp = isYes(resp, false)
			if passwordResp {
				if existingUser.Provider != "" {
					u.outClient.FailS("User has a third-party auth provider, so there is nothing to update.")
					return 1
				}

				if resp, err := getStringAndValidate(ui, "New password for the user:", false, nil, true, true); err != nil {
					ui.Error(err.Error())
					return 1
				} else {
					params["password"] = resp
				}
			}
		}

		// Display name
		if resp, err := getStringAndValidateDefault(
			ui, "Would you like to update the display name [y/n] (default: n):", "n", true, nil, false, false); err != nil {
			ui.Error(err.Error())
			return utils.GetExecStatus(err)
		} else {
			displannameResp = isYes(resp, false)
			if displannameResp {
				if resp, err := getStringAndValidate(ui, "Display name:", true, nil, false, false); err != nil {
					ui.Error(err.Error())
					return 1
				} else {
					params[cst.DataDisplayname] = resp
				}
			}
		}

		if !passwordResp && !displannameResp {
			return 0
		}

		resp, apiError := u.submitUser(params["username"], params, true)
		u.outClient.WriteResponse(resp, apiError)
		return utils.GetExecStatus(apiError)
	}

	// Create user workflow
	if resp, err := getStringAndValidate(ui, "Display name:", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params[cst.DataDisplayname] = resp
	}

	baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")
	data, err := handleSearch(nil, baseType, u.request)
	if err != nil {
		u.outClient.Fail(err)
		return utils.GetExecStatus(err)
	}
	providers, parseErr := parseAuthProviders(data)
	if parseErr != nil {
		u.outClient.FailS("Failed to parse out available auth providers.")
		return utils.GetExecStatus(parseErr)
	}

	if len(providers) == 0 {
		if resp, err := getStringAndValidate(ui, "New password for the user:", false, nil, true, true); err != nil {
			ui.Error(err.Error())
			return 1
		} else {
			params["password"] = resp
		}
	} else {
		var providerName string
		options := []option{{"local", "local"}}
		for _, p := range providers {
			v := fmt.Sprintf("%s:%s", p.Name, p.Type)
			options = append(options, option{v, strings.Replace(v, ":", " - ", 1)})
		}
		if resp, err := getStringAndValidate(ui, "Provider:", true, options, false, false); err != nil {
			ui.Error(err.Error())
			return 1
		} else {
			providerName = resp
		}

		if p := strings.Split(providerName, ":"); p[0] == "local" {
			if resp, err := getStringAndValidate(ui, "New password for the user:", false, nil, true, true); err != nil {
				ui.Error(err.Error())
				return 1
			} else {
				params["password"] = resp
			}
		} else if p[1] == cst.ThyOne {
			params["provider"] = strings.Split(providerName, ":")[0]
		} else {
			if resp, err := getStringAndValidate(ui, "External ID:", false, nil, false, false); err != nil {
				ui.Error(err.Error())
				return 1
			} else {
				params["provider"] = strings.Split(providerName, ":")[0]
				params["externalId"] = resp
			}
		}
	}

	resp, apiError := u.submitUser(params["username"], params, false)
	u.outClient.WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func (u User) submitUser(path string, data map[string]string, update bool) ([]byte, *errors.ApiError) {
	if update {
		uri := paths.CreateResourceURI(cst.NounUser, path, "", true, nil, true)
		return u.request.DoRequest("PUT", uri, &data)
	}
	uri := paths.CreateResourceURI(cst.NounUser, "", "", true, nil, true)
	return u.request.DoRequest("POST", uri, &data)
}

func parseAuthProviders(data []byte) ([]authProvider, error) {
	var providers []authProvider
	var resp map[string]interface{}
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	d, ok := resp["data"].([]interface{})
	if !ok {
		return nil, nil
	}
	err = mapstructure.Decode(d, &providers)
	if err != nil {
		return nil, err
	}
	return providers, nil
}

type userModel struct {
	UserName   string `json:"userName" `
	ExternalID string `json:"externalId" `
	Provider   string `json:"provider" `
}
