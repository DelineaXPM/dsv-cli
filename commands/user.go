package cmd

import (
	"encoding/json"
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

var (
	errMustSpecifyPasswordOrDisplayname = errors.NewF("error: must specify %s or %s", cst.DataPassword, cst.DataDisplayname)
	errWrongDisplayName                 = errors.NewS("error: displayname field must be between 3 and 100 characters")
)

func GetUserCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUser},
		SynopsisText: "Manage users",
		HelpText: fmt.Sprintf(`Execute an action on a %s from %s

Usage:
   • user %[3]s
   • user --username %[3]s
`, cst.NounUser, cst.ProductName, cst.ExampleUser),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataUsername, Usage: "Username of user to fetch (required)"},
			{Name: cst.Version, Usage: "List the current and last (n) versions"},
		},
		MinNumberArgs: 1,
		RunFunc: func(vcli vaultcli.CLI, args []string) int {
			userData := viper.GetString(cst.DataUsername)
			if userData == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
				userData = args[0]
			}
			if userData == "" {
				return cli.RunResultHelp
			}
			return handleUserReadCmd(vcli, args)
		},
	})
}

func GetUserReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUser, cst.Read},
		SynopsisText: "user read (<username> | --username)",
		HelpText: fmt.Sprintf(`Read a %[1]s from %[2]s

Usage:
   • user read %[3]s
   • user read --username %[3]s
`, cst.NounUser, cst.ProductName, cst.ExampleUser),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataUsername, Usage: "Username of user to fetch (required)"},
			{Name: cst.Version, Usage: "List the current and last (n) versions"},
		},
		MinNumberArgs: 1,
		RunFunc: func(vcli vaultcli.CLI, args []string) int {
			return handleUserReadCmd(vcli, args)
		},
	})
}

func GetUserSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUser, cst.Search},
		SynopsisText: "user search (<query> | --query)",
		HelpText: fmt.Sprintf(`Search for a %[1]s from %[2]s

Usage:
   • user search %[3]s
   • user search --query %[3]s
`, cst.NounUser, cst.ProductName, cst.ExampleUserSearch),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Query, Shorthand: "q", Usage: fmt.Sprintf("%s of %ss to fetch (optional)", strings.Title(cst.Query), cst.NounUser)},
			{Name: cst.Limit, Shorthand: "l", Usage: "Maximum number of results per cursor (optional)"},
			{Name: cst.Cursor, Usage: cst.CursorHelpMessage},
		},
		RunFunc: func(vcli vaultcli.CLI, args []string) int {
			return handleUserSearchCmd(vcli, args)
		},
	})
}

func GetUserDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUser, cst.Delete},
		SynopsisText: "user delete (<username> | --username)",
		HelpText: fmt.Sprintf(`Delete a %[1]s from %[2]s

Usage:
   • user delete %[3]s
   • user delete --username %[3]s --force
`, cst.NounUser, cst.ProductName, cst.ExampleUser),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataUsername, Usage: fmt.Sprintf("%s of %s to fetch (required)", strings.Title(cst.DataUsername), cst.NounUser)},
			{Name: cst.Force, Usage: fmt.Sprintf("Immediately delete %s", cst.NounUser), ValueType: "bool"},
		},
		MinNumberArgs: 1,
		RunFunc: func(vcli vaultcli.CLI, args []string) int {
			return handleUserDeleteCmd(vcli, args)
		},
	})
}

func GetUserRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUser, cst.Restore},
		SynopsisText: "user restore (<username> | --username)",
		HelpText: fmt.Sprintf(`Restore a deleted %[1]s in %[2]s
Usage:
   • user restore %[3]s
   • user restore --username %[3]s
`, cst.NounUser, cst.ProductName, cst.ExampleUser),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataUsername, Usage: fmt.Sprintf("%s of %s to fetch (required)", strings.Title(cst.DataUsername), cst.NounUser)},
		},
		MinNumberArgs: 1,
		RunFunc: func(vcli vaultcli.CLI, args []string) int {
			return handleUserRestoreCmd(vcli, args)
		},
	})
}

func GetUserCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUser, cst.Create},
		SynopsisText: "user create (<username> <password> | --username --password)",
		HelpText: fmt.Sprintf(`Create a %[1]s in %[2]s

Usage:
   • user create --username %[3]s --password %[4]s
   • user create --username %[3]s --external-id svc1@project1.iam.gserviceaccount.com --provider project1.gcloud --password %[4]s
`, cst.NounUser, cst.ProductName, cst.ExampleUser, cst.ExamplePassword),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataUsername, Usage: fmt.Sprintf("%s of %s to be updated (required)", strings.Title(cst.DataUsername), cst.NounUser)},
			{Name: cst.DataDisplayname, Usage: fmt.Sprintf("%s of %s to be updated", strings.Title(cst.DataDisplayname), cst.NounUser)},
			{Name: cst.DataPassword, Usage: fmt.Sprintf("%s of %s to be updated (required)", strings.Title(cst.Password), cst.NounUser)},
			{Name: cst.DataExternalID, Usage: fmt.Sprintf("%s of %s to be updated", strings.Title(strings.Replace(cst.DataExternalID, ".", " ", -1)), cst.NounUser)},
			{Name: cst.DataProvider, Usage: fmt.Sprintf("External %s of %s to be updated", strings.Title(cst.DataProvider), cst.NounUser)},
		},
		RunFunc: func(vcli vaultcli.CLI, args []string) int {
			if OnlyGlobalArgs(args) {
				return handleUserCreateWorkflow(vcli, args)
			}
			return handleUserCreateCmd(vcli, args)
		},
	})
}

func GetUserUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUser, cst.Update},
		SynopsisText: "user update (<username> <password> | (--username) --password)",
		HelpText: fmt.Sprintf(`Update a %[1]s's password in %[2]s

Usage:
   • user update --username %[3]s --password %[4]s
`, cst.NounUser, cst.ProductName, cst.ExampleUser, cst.ExamplePassword),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataPassword, Usage: fmt.Sprintf("%s of %s to be updated (required)", strings.Title(cst.Password), cst.NounUser)},
			{Name: cst.DataUsername, Usage: fmt.Sprintf("%s of %s to be updated (required)", strings.Title(cst.DataUsername), cst.NounUser)},
			{Name: cst.DataDisplayname, Usage: fmt.Sprintf("%s of %s to be updated", strings.Title(cst.DataDisplayname), cst.NounUser)},
		},
		RunFunc: func(vcli vaultcli.CLI, args []string) int {
			if OnlyGlobalArgs(args) {
				return handleUserUpdateWorkflow(vcli, args)
			}
			return handleUserUpdateCmd(vcli, args)
		},
	})
}

func handleUserReadCmd(vcli vaultcli.CLI, args []string) int {
	userName := viper.GetString(cst.DataUsername)
	if userName == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		userName = args[0]
	}
	if userName == "" {
		err := errors.NewS("error: must specify " + cst.DataUsername)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	userName = paths.ProcessResource(userName)
	version := viper.GetString(cst.Version)
	if strings.TrimSpace(version) != "" {
		userName = fmt.Sprint(userName, "/", cst.Version, "/", version)
	}

	data, err := userRead(vcli, userName)
	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func handleUserSearchCmd(vcli vaultcli.CLI, args []string) int {
	query := viper.GetString(cst.Query)
	limit := viper.GetString(cst.Limit)
	cursor := viper.GetString(cst.Cursor)

	if query == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		query = args[0]
	}

	data, apiErr := userSearch(vcli, &userSearchParams{query: query, limit: limit, cursor: cursor})
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleUserDeleteCmd(vcli vaultcli.CLI, args []string) int {
	userName := viper.GetString(cst.DataUsername)
	force := viper.GetBool(cst.Force)
	if userName == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		userName = args[0]
	}
	if userName == "" {
		err := errors.NewS("error: must specify " + cst.DataUsername)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, apiErr := userDelete(vcli, paths.ProcessResource(userName), force)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleUserRestoreCmd(vcli vaultcli.CLI, args []string) int {
	userName := viper.GetString(cst.DataUsername)
	if userName == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		userName = args[0]
	}
	if userName == "" {
		err := errors.NewS("error: must specify " + cst.DataUsername)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, apiErr := userRestore(vcli, paths.ProcessResource(userName))
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleUserCreateCmd(vcli vaultcli.CLI, args []string) int {
	userName := viper.GetString(cst.DataUsername)
	password := viper.GetString(cst.DataPassword)
	provider := viper.GetString(cst.DataProvider)
	externalID := viper.GetString(cst.DataExternalID)

	if err := vaultcli.ValidateUsername(userName); err != nil {
		vcli.Out().FailF("error: %s %q is invalid: %v", cst.DataUsername, userName, err)
		return utils.GetExecStatus(err)
	}

	isUserLocal := provider == "" && externalID == ""
	if password == "" && isUserLocal {
		err := errors.NewS("error: must specify password for local users")
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	if !isUserLocal {
		password = ""
	}

	displayName := viper.GetString(cst.DataDisplayname)

	body := &userCreateRequest{
		Username:    userName,
		Password:    password,
		DisplayName: displayName,
		Provider:    provider,
		ExternalID:  externalID,
	}
	resp, apiError := userCreate(vcli, body)
	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func handleUserUpdateCmd(vcli vaultcli.CLI, args []string) int {
	username := viper.GetString(cst.DataUsername)
	if username == "" {
		err := errors.NewS("error: must specify " + cst.DataUsername)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	displayNameExists := hasFlag(args, "--"+cst.DataDisplayname)
	passData := viper.GetString(cst.DataPassword)
	displayName := viper.GetString(cst.DataDisplayname)
	if passData == "" && !displayNameExists {
		err := errMustSpecifyPasswordOrDisplayname
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	displayNameLen := len(displayName)
	if displayNameExists && (displayNameLen < 3 || displayNameLen > 100) {
		err := errWrongDisplayName
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	body := &userUpdateRequest{Password: passData, DisplayName: displayName}
	resp, apiError := userUpdate(vcli, username, body)
	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

// Wizards:

func handleUserCreateWorkflow(vcli vaultcli.CLI, args []string) int {
	providers, apiError := listAuthProviders(vcli)
	if apiError != nil {
		httpResp := apiError.HttpResponse()
		// If API returns "403 Forbidden" still allow to try to create a local user.
		if httpResp == nil || httpResp.StatusCode != http.StatusForbidden {
			vcli.Out().FailS("Failed to get available auth providers.")
			return utils.GetExecStatus(apiError)
		}
	}

	qs := []*survey.Question{
		{
			Name:   "Username",
			Prompt: &survey.Input{Message: "Username:"},
			Validate: func(ans interface{}) error {
				answer := ans.(string)
				answer = strings.TrimSpace(answer)
				if len(answer) == 0 {
					return errors.NewS("Value is required")
				}
				if err := vaultcli.ValidateUsername(answer); err != nil {
					return err
				}
				_, apiError := userRead(vcli, answer)
				if apiError == nil {
					return errors.NewS("A user with this username already exists.")
				}
				return nil
			},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "DisplayName",
			Prompt:    &survey.Input{Message: "Display name:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
	}

	if len(providers) != 0 {
		providersOpts := []string{"local"}
		for _, p := range providers {
			providersOpts = append(providersOpts, fmt.Sprintf("%s [type: %s]", p.Name, p.Type))
		}
		qs = append(qs, &survey.Question{
			Name: "provider",
			Prompt: &survey.Select{
				Message: "Select provider:",
				Options: providersOpts,
			},
		})
	}

	answers := struct {
		Username    string
		DisplayName string
		Provider    string
	}{}
	survErr := survey.Ask(qs, &answers)
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}

	var password, provider, externalID string

	if answers.Provider == "" || answers.Provider == "local" {
		passwordPrompt := &survey.Password{Message: "New password for the user:"}
		survErr := survey.AskOne(passwordPrompt, &password, survey.WithValidator(survey.Required))
		if survErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(survErr))
			return utils.GetExecStatus(survErr)
		}
	} else {
		provider = strings.Split(answers.Provider, " [type: ")[0]

		if !strings.HasSuffix(answers.Provider, fmt.Sprintf("[type: %s]", cst.ThyOne)) {
			externalIDPrompt := &survey.Password{Message: "External ID:"}
			survErr := survey.AskOne(externalIDPrompt, &externalID, survey.WithValidator(vaultcli.SurveyRequired))
			if survErr != nil {
				vcli.Out().WriteResponse(nil, errors.New(survErr))
				return utils.GetExecStatus(survErr)
			}
		}
	}

	body := &userCreateRequest{
		Username:    answers.Username,
		Password:    password,
		DisplayName: answers.DisplayName,
		Provider:    provider,
		ExternalID:  strings.TrimSpace(externalID),
	}
	resp, apiError := userCreate(vcli, body)
	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func handleUserUpdateWorkflow(vcli vaultcli.CLI, args []string) int {
	var username string
	usernamePrompt := &survey.Input{Message: "Username:"}
	survErr := survey.AskOne(usernamePrompt, &username, survey.WithValidator(vaultcli.SurveyRequired))
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}

	var provider string
	providerFetched := false
	data, apiError := userRead(vcli, username)
	if apiError != nil {
		httpResp := apiError.HttpResponse()
		if httpResp == nil || httpResp.StatusCode != http.StatusForbidden {
			vcli.Out().Fail(apiError)
			return utils.GetExecStatus(apiError)
		}

		var confirm bool
		confirmPrompt := &survey.Confirm{
			Message: "You are not allowed to read user with that username. Do you want to continue?",
			Default: true,
		}
		survErr := survey.AskOne(confirmPrompt, &confirm)
		if survErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(survErr))
			return utils.GetExecStatus(survErr)
		}
		if !confirm {
			return 0
		}
	} else {
		existingUser := struct {
			Provider string `json:"provider"`
		}{}
		err := json.Unmarshal(data, &existingUser)
		if err != nil {
			vcli.Out().Fail(err)
			return utils.GetExecStatus(err)
		}
		provider = existingUser.Provider
		providerFetched = true
	}

	var password, displayName string

	if !providerFetched || (providerFetched && provider == "") {
		var confirm bool
		confirmPrompt := &survey.Confirm{
			Message: "Would you like to update the password?",
			Default: false,
		}
		survErr := survey.AskOne(confirmPrompt, &confirm)
		if survErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(survErr))
			return utils.GetExecStatus(survErr)
		}

		if confirm {
			passwordPrompt := &survey.Password{Message: "New password for the user:"}
			survErr := survey.AskOne(passwordPrompt, &password, survey.WithValidator(survey.Required))
			if survErr != nil {
				vcli.Out().WriteResponse(nil, errors.New(survErr))
				return utils.GetExecStatus(survErr)
			}
		}
	}

	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: "Would you like to update the display name?",
		Default: false,
	}
	survErr = survey.AskOne(confirmPrompt, &confirm)
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}

	if confirm {
		displayPrompt := &survey.Input{Message: "Display name:"}
		survErr := survey.AskOne(displayPrompt, &displayName, survey.WithValidator(vaultcli.SurveyRequired))
		if survErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(survErr))
			return utils.GetExecStatus(survErr)
		}
		displayName = strings.TrimSpace(displayName)
	}

	if password == "" && displayName == "" {
		return 0
	}

	body := &userUpdateRequest{Password: password, DisplayName: displayName}
	resp, apiError := userUpdate(vcli, username, body)
	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

// API callers:

type userCreateRequest struct {
	Username    string `json:"userName"`
	Password    string `json:"password,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Provider    string `json:"provider,omitempty"`
	ExternalID  string `json:"externalId,omitempty"`
}

func userCreate(vcli vaultcli.CLI, body *userCreateRequest) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounUsers, "", "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, &body)
}

func userRead(vcli vaultcli.CLI, username string) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounUsers, username, "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}

type userUpdateRequest struct {
	Password    string `json:"password,omitempty"`
	DisplayName string `json:"displayname,omitempty"`
}

func userUpdate(vcli vaultcli.CLI, username string, body *userUpdateRequest) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounUsers, username, "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPut, uri, body)
}

func userDelete(vcli vaultcli.CLI, username string, force bool) ([]byte, *errors.ApiError) {
	query := map[string]string{"force": strconv.FormatBool(force)}
	uri := paths.CreateResourceURI(cst.NounUsers, username, "", true, query)
	return vcli.HTTPClient().DoRequest(http.MethodDelete, uri, nil)
}

func userRestore(vcli vaultcli.CLI, username string) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounUsers, username, "/restore", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPut, uri, nil)
}

type userSearchParams struct {
	query  string
	limit  string
	cursor string
}

func userSearch(vcli vaultcli.CLI, p *userSearchParams) ([]byte, *errors.ApiError) {
	queryParams := map[string]string{}
	if p.query != "" {
		queryParams[cst.SearchKey] = p.query
	}
	if p.limit != "" {
		queryParams[cst.Limit] = p.limit
	}
	if p.cursor != "" {
		queryParams[cst.Cursor] = p.cursor
	}
	uri := paths.CreateURI(cst.NounUsers, queryParams)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}
