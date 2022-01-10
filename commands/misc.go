package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"thy/auth"

	cst "thy/constants"
	"thy/errors"
	"thy/format"
	"thy/paths"
	"thy/requests"

	"github.com/spf13/viper"
	"github.com/thycotic-rd/cli"
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
	if m.outClient == nil {
		m.outClient = format.NewDefaultOutClient()
	}

	subject, err := auth.GetCurrentIdentity()
	if err == nil {
		m.outClient.WriteResponse([]byte(subject), nil)
	} else {
		m.outClient.FailS("Failed to parse the subject from the auth token, try re-authenticating")
	}
	return 0
}

func (m Misc) handleEvaluateFlag(args []string) int {
	if len(args) < 1 {
		return cli.RunResultHelp
	}
	arg := args[0]
	arg = strings.TrimPrefix(arg, "--")
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
	uri := paths.CreateResourceURI(resourceType, "", "", false, queryParams, false)
	return request.DoRequest(http.MethodGet, uri, nil)
}

func hasFlag(args []string, flagName string) bool {
	for _, fn := range args {
		if strings.HasPrefix(fn, flagName) {
			return true
		}
	}
	return false
}
