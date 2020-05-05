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

type User struct {
	request   requests.Client
	outClient format.OutClient
}

func GetDataOpUserWrappers(targetEntity string) cli.PredictorWrappers {
	return cli.PredictorWrappers{
		preds.LongFlag(cst.DataUsername):   cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataUsername, Usage: fmt.Sprintf("%s of %s to be updated (required)", strings.Title(cst.DataUsername), targetEntity)}), false},
		preds.LongFlag(cst.DataPassword):   cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataPassword, Usage: fmt.Sprintf("%s of %s to be updated (required)", strings.Title(cst.Password), targetEntity)}), false},
		preds.LongFlag(cst.DataExternalID): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataExternalID, Usage: fmt.Sprintf("%s of %s to be updated", strings.Title(strings.Replace(cst.DataExternalID, ".", " ", -1)), targetEntity)}), false},
		preds.LongFlag(cst.DataProvider):   cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataProvider, Usage: fmt.Sprintf("External %s of %s to be updated", strings.Title(cst.DataProvider), targetEntity)}), false},
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
			preds.LongFlag(cst.Cursor): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Cursor, Usage: fmt.Sprint("Next cursor for additional results (optional)")}), false},
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
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounUser, cst.Restore),
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
		RunFunc:      User{requests.NewHttpClient(), nil}.handleUserPostCmd,
		SynopsisText: fmt.Sprintf("%s (<username> <password> | --username --password)", cst.Create),
		HelpText: fmt.Sprintf(`Create a %[2]s in %[3]s

Usage:
   • user %[1]s %[4]s %[5]s
   • user %[1]s --username %[4]s --password %[5]s
   • user %[1]s --username %[4]s --external-id svc1@project1.iam.gserviceaccount.com --provider project1.gcloud
		`, cst.Create, cst.NounUser, cst.ProductName, cst.ExampleUser, cst.ExamplePassword),
		FlagsPredictor: GetDataOpUserWrappers(cst.NounUser),
		MinNumberArgs:  2,
	})
}

func (u User) handleUserReadCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	userData := viper.GetString(cst.DataUsername)
	if userData == "" && len(args) > 0 {
		userData = args[0]
	}
	if userData == "" {
		err = errors.NewS("error: must specify " + cst.DataUsername)
	} else {
		version := viper.GetString(cst.Version)
		if strings.TrimSpace(version) != "" {
			userData = fmt.Sprint(userData, "/", cst.Version, "/", version)
		}
		uri := utils.CreateResourceURI(cst.NounUser, userData, "", true, nil, true)
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
		uri := utils.CreateResourceURI(cst.NounUser, "", "", false, queryParams, true)
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
		uri := utils.CreateResourceURI(cst.NounUser, userName, "", true, query, true)
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
		uri, err := utils.GetResourceURIFromResourcePath(cst.NounUser, userName, "", "", true, nil)
		if err == nil {
			uri += "/restore"
			data, err = u.request.DoRequest("PUT", uri, nil)
		}
	}

	u.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (u User) handleUserPostCmd(args []string) int {
	userData := viper.GetString(cst.DataUsername)
	if userData == "" && len(args) > 0 {
		userData = args[0]
	}
	if userData == "" {
		err := errors.NewS("error: must specify " + cst.DataUsername)
		outClient := u.outClient
		if outClient == nil {
			outClient = format.NewDefaultOutClient()
		}
		outClient.WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}
	passData := viper.GetString(cst.DataPassword)
	if passData == "" && len(args) > 1 {
		passData = args[1]
	}
	if passData == "" {
		err := errors.NewS("error: must specify " + cst.DataPassword)
		outClient := u.outClient
		if outClient == nil {
			outClient = format.NewDefaultOutClient()
		}
		outClient.WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}
	providerData := viper.GetString(cst.DataProvider)
	externalIdData := viper.GetString(cst.DataExternalID)

	data := map[string]string{
		"name":       userData,
		"username":   userData,
		"password":   passData,
		"provider":   providerData,
		"externalId": externalIdData,
	}

	uri := utils.CreateResourceURI(cst.NounUser, "", "", true, nil, true)
	resp, err := u.request.DoRequest("POST", uri, &data)
	outClient := u.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}
