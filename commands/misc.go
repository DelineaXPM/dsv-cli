package cmd

import (
	"fmt"
	"strings"
	cst "thy/constants"
	"thy/errors"
	"thy/format"
	"thy/requests"
	"thy/utils"

	"github.com/thycotic-rd/cli"
	"github.com/thycotic-rd/viper"
)

type Misc struct {
	outClient format.OutClient
}

func GetWhoAmICmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounWhoAmI},
		RunFunc:      Misc{}.handleWhoAmICmd,
		SynopsisText: cst.NounWhoAmI,
		HelpText:     fmt.Sprintf("%s returns the current user identity, accounting for config, env, and flags", cst.NounWhoAmI),
		NoPreAuth:    true,
	})
}

func GetEvaluateFlagCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.EvaluateFlag},
		RunFunc:      Misc{}.handleEvaluateFlag,
		SynopsisText: cst.EvaluateFlag,
		HelpText:     fmt.Sprintf("%s returns the value of the variable, accounting for config, env, and flags", cst.EvaluateFlag),
		NoPreAuth:    true,
	})
}

func (m Misc) handleWhoAmICmd(args []string) int {
	var user string
	if viper.GetString(cst.AuthType) == "azure" {
		outClient := m.outClient
		if outClient == nil {
			outClient = format.NewDefaultOutClient()
		}
		outClient.WriteResponse([]byte("whoami does not yet support displaying an ID for a user authenticated with Azure"), nil)
		return 0
	}
	if u := viper.GetString(cst.Username); u != "" {
		user = u
	} else if u := viper.GetString(cst.AwsProfile); u != "" {
		user = u
	} else if u := viper.GetString(cst.AuthClientID); u != "" {
		user = u
	}
	outClient := m.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse([]byte(user), nil)
	return 0
}

func (m Misc) handleEvaluateFlag(args []string) int {
	if len(args) < 1 {
		return cli.RunResultHelp
	}
	arg := args[0]
	if strings.HasPrefix(arg, "--") {
		arg = arg[2:]
	}
	arg = strings.Replace(arg, "-", ".", -1)
	arg = strings.Replace(arg, "_", ".", -1)

	data := []byte(viper.GetString(arg))
	outClient := m.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse(data, nil)
	return 0
}

func handleSearch(args []string, resourceType string, request requests.Client) (data []byte, err *errors.ApiError) {
	query := viper.GetString(cst.Query)
	limit := viper.GetString(cst.Limit)
	cursor := viper.GetString(cst.Cursor)
	if query == "" && len(args) == 1 {
		query = args[0]
	}

	queryParams := map[string]string{
		cst.SearchKey: query,
		cst.Limit:     limit,
		cst.Cursor:    cursor,
	}
	uri := utils.CreateResourceURI(resourceType, "", "", false, queryParams, false)
	return request.DoRequest("GET", uri, nil)
}
