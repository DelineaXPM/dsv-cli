package cmd

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	cst "thy/constants"
	"thy/errors"
	"thy/internal/predictor"
	"thy/paths"
	"thy/utils"
	"thy/vaultcli"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

func GetClientCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounClient},
		SynopsisText: "client (<client-id> | --client-id)",
		HelpText: fmt.Sprintf(`Execute an action on a %[1]s in %[2]s

Usage:
   • %[1]s %[3]s
   • %[1]s --client-id %[3]s
		`, cst.NounClient, cst.ProductName, cst.ExampleClientID),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.ClientID, Usage: fmt.Sprintf("ID of the %s ", cst.NounClient)},
		},
		MinNumberArgs: 1,
		RunFunc: func(args []string) int {
			name := viper.GetString(cst.DataName)
			if name == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
				name = args[0]
			}
			if name == "" {
				return cli.RunResultHelp
			}
			return handleClientReadCmd(vaultcli.New(), args)
		},
	})
}

func GetClientReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounClient, cst.Read},
		SynopsisText: fmt.Sprintf("%s %s (<client-id> | --client-id)", cst.NounClient, cst.Read),
		HelpText: fmt.Sprintf(`Execute an action on a %[1]s in %[2]s

Usage:
   • %[1]s %[4]s %[3]s
   • %[1]s %[4]s --client-id %[3]s
		`, cst.NounClient, cst.ProductName, cst.ExampleClientID, cst.Read),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.ClientID, Usage: fmt.Sprintf("ID of the %s ", cst.NounClient)},
		},
		MinNumberArgs: 1,
		RunFunc: func(args []string) int {
			return handleClientReadCmd(vaultcli.New(), args)
		},
	})
}

func GetClientDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounClient, cst.Delete},
		SynopsisText: fmt.Sprintf("%s %s (<client-id> | --client-id)", cst.NounClient, cst.Delete),
		HelpText: fmt.Sprintf(`Delete a %[1]s from %[2]s

Usage:
   • %[1]s %[4]s %[3]s
   • %[1]s %[4]s --client-id %[3]s --force
		`, cst.NounClient, cst.ProductName, cst.ExampleClientID, cst.Delete),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.ClientID, Usage: fmt.Sprintf("ID of the %s ", cst.NounClient)},
			{Name: cst.Force, Usage: fmt.Sprintf("Immediately delete %s", cst.NounClient), ValueType: "bool"},
		},
		MinNumberArgs: 1,
		RunFunc: func(args []string) int {
			return handleClientDeleteCmd(vaultcli.New(), args)
		},
	})
}

func GetClientRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounClient, cst.Restore},
		SynopsisText: fmt.Sprintf("%s %s (<client-id> | --client-id)", cst.NounClient, cst.Restore),
		HelpText: fmt.Sprintf(`Restore a deleted %[1]s in %[2]s

Usage:
   • %[1]s %[4]s %[3]s
		`, cst.NounClient, cst.ProductName, cst.ExampleClientID, cst.Restore),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.ClientID, Usage: fmt.Sprintf("ID of the %s ", cst.NounClient)},
		},
		MinNumberArgs: 1,
		RunFunc: func(args []string) int {
			return handleClientRestoreCmd(vaultcli.New(), args)
		},
	})
}

func GetClientCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounClient, cst.Create},
		SynopsisText: fmt.Sprintf("%s %s (<role> | --role) |(<uses> | --uses)|(<desc> | --desc)|(<ttl> | --ttl)| (<url> | --url) | ( <urlTTL> | --urlTTL)", cst.NounClient, cst.Create),
		HelpText: fmt.Sprintf(`%[4]s a %[1]s in %[2]s

Usage:
   • %[1]s %[4]s %[3]s
   • %[1]s %[4]s --role %[3]s
		`, cst.NounClient, cst.ProductName, cst.ExampleRoleName, cst.Create),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.NounRole, Usage: fmt.Sprintf("Name of the %s ", cst.NounRole)},
			{Name: cst.NounBootstrapUrl, Usage: "Whether to generate a one-time use URL instead of secret (optional)", ValueType: "bool"},
			{Name: cst.NounBootstrapUrlTTL, Usage: "TTL for the generated URL (optional)"},
			{Name: cst.NounClientUses, Usage: "The number of times the client credential can be read. If set to 0, it can be used infinitely. Default is 0 (optional)"},
			{Name: cst.NounClientDesc, Usage: "Client credential description (optional)"},
			{Name: cst.NounClientTTL, Usage: "How long until the client credential expires. If set to 0, it can be used indefinitely. Default is 0 (optional)"},
		},
		RunFunc: func(args []string) int {
			if OnlyGlobalArgs(args) {
				return handleClientCreateWizard(vaultcli.New(), args)
			}
			return handleClientCreateCmd(vaultcli.New(), args)
		},
	})
}

func GetClientSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounClient, cst.Search},
		SynopsisText: fmt.Sprintf("%s (<role> | --role)", cst.Search),
		HelpText: fmt.Sprintf(`Search for %[1]ss attached to a given %[5]s in %[2]s

Usage:
   • %[1]s %[4]s %[3]s
   • %[1]s %[4]s --role %[3]s
		`, cst.NounClient, cst.ProductName, cst.ExampleRoleName, cst.Search, cst.NounRole),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.NounRole, Usage: "Role that has attached clients (required)"},
			{Name: cst.Limit, Shorthand: "l", Usage: "Maximum number of results per cursor (optional)"},
			{Name: cst.Cursor, Usage: cst.CursorHelpMessage},
		},
		MinNumberArgs: 1,
		RunFunc: func(args []string) int {
			return handleClientSearchCmd(vaultcli.New(), args)
		},
	})
}

func handleClientReadCmd(vcli vaultcli.CLI, args []string) int {
	clientID := viper.GetString(cst.ClientID)
	if clientID == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		clientID = args[0]
	}
	if clientID == "" {
		err := errors.NewS("error: must specify " + cst.ClientID)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, apiErr := clientRead(vcli, clientID)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleClientDeleteCmd(vcli vaultcli.CLI, args []string) int {
	force := viper.GetBool(cst.Force)
	clientID := viper.GetString(cst.ClientID)
	if clientID == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		clientID = args[0]
	}
	if clientID == "" {
		err := errors.NewS("error: must specify " + cst.ClientID)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, apiErr := clientDelete(vcli, clientID, force)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleClientRestoreCmd(vcli vaultcli.CLI, args []string) int {
	clientID := viper.GetString(cst.ClientID)
	if clientID == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		clientID = args[0]
	}
	if clientID == "" {
		err := errors.NewS("error: must specify " + cst.ClientID)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, apiErr := clientRestore(vcli, clientID)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleClientCreateCmd(vcli vaultcli.CLI, args []string) int {
	roleName := viper.GetString(cst.NounRole)
	if roleName == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		roleName = args[0]
	}
	if roleName == "" {
		return cli.RunResultHelp
	}
	client := &clientCreateRequest{
		Role:         roleName,
		UrlRequested: viper.GetBool(cst.NounBootstrapUrl),
		UrlTTL:       viper.GetInt64(cst.NounBootstrapUrlTTL),
		Uses:         viper.GetInt(cst.NounClientUses),
		Description:  viper.GetString(cst.NounClientDesc),
		TTL:          viper.GetInt64(cst.NounClientTTL),
	}

	data, apiErr := clientCreate(vcli, client)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleClientSearchCmd(vcli vaultcli.CLI, args []string) int {
	role := viper.GetString(cst.NounRole)
	limit := viper.GetString(cst.Limit)
	cursor := viper.GetString(cst.Cursor)
	if role == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		role = args[0]
	}
	if role == "" {
		err := errors.NewS("error: must specify " + cst.NounRole)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, apiErr := clientSearch(vcli, &clientSearchParams{role: role, limit: limit, cursor: cursor})
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

// Wizards:

func handleClientCreateWizard(vcli vaultcli.CLI, args []string) int {
	qs := []*survey.Question{
		{
			Name:   "Role",
			Prompt: &survey.Input{Message: "Role name:"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				_, apiError := roleRead(vcli, answer)
				if apiError != nil &&
					apiError.HttpResponse() != nil &&
					apiError.HttpResponse().StatusCode == http.StatusNotFound {
					return errors.NewS("A role with this name does not exist.")
				}
				return nil
			},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "Description",
			Prompt:    &survey.Input{Message: "Client description:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "TTL",
			Prompt:    &survey.Input{Message: "Client TTL (in seconds):", Default: "0"},
			Validate:  vaultcli.SurveyRequiredInt,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:   "UrlRequested",
			Prompt: &survey.Confirm{Message: "Request Bootstrap URL?", Default: false},
		},
	}

	client := clientCreateRequest{}
	survErr := survey.Ask(qs, &client)
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}

	if client.UrlRequested {
		qs = []*survey.Question{
			{
				Name:      "UrlTTL",
				Prompt:    &survey.Input{Message: "Bootstrap URL TTL (in seconds):"},
				Validate:  vaultcli.SurveyRequiredInt,
				Transform: vaultcli.SurveyTrimSpace,
			},
			{
				Name:      "Uses",
				Prompt:    &survey.Input{Message: "Number of client uses:", Default: "0"},
				Validate:  vaultcli.SurveyRequiredInt,
				Transform: vaultcli.SurveyTrimSpace,
			},
		}

		survErr := survey.Ask(qs, &client)
		if survErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(survErr))
			return utils.GetExecStatus(survErr)
		}
	}

	data, apiErr := clientCreate(vcli, &client)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

// API callers:

type clientCreateRequest struct {
	Role         string `json:"role"`
	UrlRequested bool   `json:"url,omitempty"`
	UrlTTL       int64  `json:"urlTTL,omitempty"`
	TTL          int64  `json:"ttl,omitempty"`
	Uses         int    `json:"usesLimit,omitempty"`
	Description  string `json:"description,omitempty"`
}

func clientCreate(vcli vaultcli.CLI, client *clientCreateRequest) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounClients, "", "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, client)
}

func clientRead(vcli vaultcli.CLI, clientID string) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounClients, clientID, "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}

func clientDelete(vcli vaultcli.CLI, clientID string, force bool) ([]byte, *errors.ApiError) {
	query := map[string]string{"force": strconv.FormatBool(force)}
	uri := paths.CreateResourceURI(cst.NounClients, clientID, "", true, query)
	return vcli.HTTPClient().DoRequest(http.MethodDelete, uri, nil)
}

func clientRestore(vcli vaultcli.CLI, clientID string) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounClients, clientID, "/restore", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPut, uri, nil)
}

type clientSearchParams struct {
	role   string
	limit  string
	cursor string
}

func clientSearch(vcli vaultcli.CLI, p *clientSearchParams) ([]byte, *errors.ApiError) {
	queryParams := map[string]string{}
	if p.role != "" {
		queryParams[cst.NounRole] = p.role
	}
	if p.limit != "" {
		queryParams[cst.Limit] = p.limit
	}
	if p.cursor != "" {
		queryParams[cst.Cursor] = p.cursor
	}
	uri := paths.CreateResourceURI(cst.NounClients, "", "", false, queryParams)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}
