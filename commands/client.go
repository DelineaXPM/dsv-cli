package cmd

import (
	"fmt"
	"strconv"
	"strings"

	cst "thy/constants"
	"thy/errors"
	"thy/format"
	"thy/paths"
	preds "thy/predictors"
	"thy/requests"
	"thy/utils"

	"github.com/posener/complete"
	"github.com/spf13/viper"
	"github.com/thycotic-rd/cli"
)

type client struct {
	request   requests.Client
	outClient format.OutClient
}

func GetNoDataOpClientWrappers(targetEntity string) cli.PredictorWrappers {
	return cli.PredictorWrappers{
		preds.LongFlag(cst.ClientID): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.ClientID, Usage: fmt.Sprintf("ID of the %s ", targetEntity)}), false},
	}
}

func GetClientCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounClient},
		RunFunc: func(args []string) int {
			name := viper.GetString(cst.DataName)
			if name == "" && len(args) > 0 {
				name = args[0]
			}
			if name == "" {
				return cli.RunResultHelp
			}
			return client{requests.NewHttpClient(), nil}.handleClientReadCmd(args)
		},
		SynopsisText: "client (<client-id> | --client-id)",
		HelpText: fmt.Sprintf(`Execute an action on a %[1]s in %[2]s

Usage:
   • %[1]s %[3]s
   • %[1]s --client-id %[3]s
		`, cst.NounClient, cst.ProductName, cst.ExampleClientID),
		FlagsPredictor: GetNoDataOpClientWrappers(cst.NounClient),
		MinNumberArgs:  1,
	})
}

func GetClientReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounClient, cst.Read},
		RunFunc:      client{requests.NewHttpClient(), nil}.handleClientReadCmd,
		SynopsisText: fmt.Sprintf("%s %s (<client-id> | --client-id)", cst.NounClient, cst.Read),
		HelpText: fmt.Sprintf(`Execute an action on a %[1]s in %[2]s

Usage:
   • %[1]s %[4]s %[3]s
   • %[1]s %[4]s --client-id %[3]s
		`, cst.NounClient, cst.ProductName, cst.ExampleClientID, cst.Read),
		FlagsPredictor: GetNoDataOpClientWrappers(cst.NounClient),
		MinNumberArgs:  1,
	})
}

func GetClientDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounClient, cst.Delete},
		RunFunc:      client{requests.NewHttpClient(), nil}.handleClientDeleteCmd,
		SynopsisText: fmt.Sprintf("%s %s (<client-id> | --client-id)", cst.NounClient, cst.Delete),
		HelpText: fmt.Sprintf(`Delete a %[1]s from %[2]s

Usage:
   • %[1]s %[4]s %[3]s
   • %[1]s %[4]s --client-id %[3]s --force
		`, cst.NounClient, cst.ProductName, cst.ExampleClientID, cst.Delete),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.ClientID): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.ClientID, Usage: fmt.Sprintf("ID of the %s ", cst.NounClient)}), false},
			preds.LongFlag(cst.Force):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Force, Shorthand: "", Usage: fmt.Sprintf("Immediately delete %s", cst.NounClient), Global: false, ValueType: "bool"}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetClientRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounClient, cst.Restore},
		RunFunc:      client{requests.NewHttpClient(), nil}.handleClientRestoreCmd,
		SynopsisText: fmt.Sprintf("%s %s (<client-id> | --client-id)", cst.NounClient, cst.Restore),
		HelpText: fmt.Sprintf(`Restore a deleted %[1]s in %[2]s

Usage:
   • %[1]s %[4]s %[3]s
		`, cst.NounClient, cst.ProductName, cst.ExampleClientID, cst.Restore),
		FlagsPredictor: GetNoDataOpClientWrappers(cst.NounClient),
		MinNumberArgs:  1,
	})
}

func GetClientCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounClient, cst.Create},
		RunFunc:      client{requests.NewHttpClient(), nil}.handleClientUpsertCmd,
		SynopsisText: fmt.Sprintf("%s %s (<role> | --role)", cst.NounClient, cst.Create),
		HelpText: fmt.Sprintf(`%[4]s a %[1]s in %[2]s

Usage:
   • %[1]s %[4]s %[3]s
   • %[1]s %[4]s --role %[3]s
		`, cst.NounClient, cst.ProductName, cst.ExampleRoleName, cst.Create),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.NounRole): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.NounRole, Usage: fmt.Sprintf("Name of the %s ", cst.NounRole)}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetClientSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounClient, cst.Search},
		RunFunc:      client{requests.NewHttpClient(), nil}.handleClientSearchCmd,
		SynopsisText: fmt.Sprintf("%s (<query> | --query)", cst.Search),
		HelpText: fmt.Sprintf(`Search for %[1]ss attached to a given %[5]s in %[2]s

Usage:
   • %[1]s %[4]s %[3]s
   • %[1]s %[4]s --role %[3]s
		`, cst.NounClient, cst.ProductName, cst.ExampleRoleName, cst.Search, cst.NounRole),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Query):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Query, Shorthand: "q", Usage: fmt.Sprintf("%s of %s that has attached clients (required)", strings.Title(cst.Query), cst.NounRole)}), false},
			preds.LongFlag(cst.Limit):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Limit, Shorthand: "l", Usage: fmt.Sprint("Maximum number of results per cursor (optional)")}), false},
			preds.LongFlag(cst.Cursor): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Cursor, Usage: fmt.Sprint("Next cursor for additional results (optional)")}), false},
		},
		MinNumberArgs: 1,
	})
}

func (c client) handleClientReadCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	clientID := viper.GetString(cst.ClientID)
	if clientID == "" && len(args) > 0 {
		clientID = args[0]
	}
	if clientID == "" {
		err = errors.NewS("error: must specify " + cst.ClientID)
	} else {
		uri := paths.CreateResourceURI(cst.NounClient, clientID, "", true, nil, true)
		data, err = c.request.DoRequest("GET", uri, nil)
	}

	outClient := c.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (c client) handleClientDeleteCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	force := viper.GetBool(cst.Force)
	clientID := viper.GetString(cst.ClientID)
	if clientID == "" && len(args) > 0 {
		clientID = args[0]
	}
	if clientID == "" {
		err = errors.NewS("error: must specify " + cst.ClientID)
	} else {
		query := map[string]string{"force": strconv.FormatBool(force)}
		uri := paths.CreateResourceURI(cst.NounClient, clientID, "", true, query, true)
		data, err = c.request.DoRequest("DELETE", uri, nil)
	}
	if c.outClient == nil {
		c.outClient = format.NewDefaultOutClient()
	}

	c.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (c client) handleClientRestoreCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	if c.outClient == nil {
		c.outClient = format.NewDefaultOutClient()
	}
	clientID := viper.GetString(cst.ClientID)
	if clientID == "" && len(args) > 0 {
		clientID = args[0]
	}
	if clientID == "" {
		err = errors.NewS("error: must specify " + cst.ClientID)
	} else {
		uri := paths.CreateResourceURI(cst.NounClient, clientID, "/restore", true, nil, true)
		data, err = c.request.DoRequest("PUT", uri, nil)
	}
	c.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (c client) handleClientUpsertCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	roleName := viper.GetString(cst.NounRole)
	if roleName == "" && len(args) > 0 {
		roleName = args[0]
	}
	if roleName == "" {
		return cli.RunResultHelp
	}
	client := Client{
		Role: roleName,
	}
	reqMethod := strings.ToLower(viper.GetString(cst.LastCommandKey))
	var uri string
	if reqMethod == cst.Create {
		reqMethod = "POST"
		uri = paths.CreateResourceURI(cst.NounClient, "", "", true, nil, true)
	} else {
		reqMethod = "PUT"
		uri = paths.CreateResourceURI(cst.NounClient, viper.GetString(cst.ClientID), "", true, nil, true)
	}
	data, err = c.request.DoRequest(reqMethod, uri, &client)
	outClient := c.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (c client) handleClientSearchCmd(args []string) int {
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
			cst.NounRole: query,
			cst.Limit:    limit,
			cst.Cursor:   cursor,
		}
		uri := paths.CreateResourceURI(cst.NounClient, "", "", false, queryParams, true)
		data, err = c.request.DoRequest("GET", uri, nil)
	}

	outClient := c.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

type Client struct {
	Role string `json:"role"`
}
