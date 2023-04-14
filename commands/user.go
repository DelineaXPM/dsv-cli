package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/internal/predictor"
	"github.com/DelineaXPM/dsv-cli/paths"
	"github.com/DelineaXPM/dsv-cli/utils"
	"github.com/DelineaXPM/dsv-cli/vaultcli"

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
			name := viper.GetString(cst.DataUsername)
			if name == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
				name = args[0]
			}
			if name == "" {
				return cli.RunResultHelp
			}
			return wrapError(handleUserReadCmd)(vcli, args)
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
		RunFuncE:      handleUserReadCmd,
	})
}

func GetUserSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUser, cst.Search},
		SynopsisText: "user search (<query> | --query)",
		HelpText: `Search for users from DevOps Secrets Vault

Usage:
   • user search adm
   • user search --query adm --limit 10
   • user search --sort asc --sorted-by created
`,
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Query, Shorthand: "q", Usage: fmt.Sprintf("%s of %ss to fetch (optional)", strings.Title(cst.Query), cst.NounUser)},
			{Name: cst.Limit, Shorthand: "l", Usage: cst.LimitHelpMessage},
			{Name: cst.Cursor, Usage: cst.CursorHelpMessage},
			{Name: cst.Sort, Usage: cst.SortHelpMessage},
			{Name: cst.SortedBy, Usage: "Sort by name, created or lastModified field (optional)", Default: "lastModified"},
		},
		RunFuncE: handleUserSearchCmd,
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
		RunFuncE:      handleUserDeleteCmd,
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
		RunFuncE:      handleUserRestoreCmd,
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
			{Name: cst.DataUsername, Usage: "Used as id (required) (must conform to /[a-zA-Z0-9_-@+.]{3,100}/)."},
			{Name: cst.DataDisplayname, Usage: "Name to display in UI."},
			{Name: cst.DataPassword, Usage: "Must be 8-100 chars, with an uppercase and special char from this list: ~!@#$%^&*()."},
			{Name: cst.DataExternalID, Usage: "Identifier attached to federated login e.g. AWS or ARN."},
			{Name: cst.DataProvider, Usage: "Used for linking user with federated/external auth, must match name of Auth Provider in administration section."},
		},
		RunFuncE:   handleUserCreateCmd,
		WizardFunc: handleUserCreateWizard,
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
			{Name: cst.DataPassword, Usage: "Uses interactive prompt if not sent as flag.  Must be 8-100 chars, with an uppercase and special char from this list: ~!@#$%^&*()."},
			{Name: cst.DataUsername, Usage: "Existing username to update"},
			{Name: cst.DataDisplayname, Usage: "New display name to show for username."},
		},
		RunFuncE:   handleUserUpdateCmd,
		WizardFunc: handleUserUpdateWizard,
	})
}

func handleUserReadCmd(vcli vaultcli.CLI, args []string) error {
	userName := viper.GetString(cst.DataUsername)
	if userName == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		userName = args[0]
	}
	if userName == "" {
		return fmt.Errorf("error: must specify --username")
	}

	userName = paths.ProcessResource(userName)
	version := viper.GetString(cst.Version)
	if strings.TrimSpace(version) != "" {
		userName = fmt.Sprint(userName, "/", cst.Version, "/", version)
	}

	data, err := userRead(vcli, userName)
	if err != nil {
		return err
	}
	vcli.Out().WriteResponse(data, nil)
	return nil
}

func handleUserSearchCmd(vcli vaultcli.CLI, args []string) error {
	query := viper.GetString(cst.Query)
	limit := viper.GetString(cst.Limit)
	cursor := viper.GetString(cst.Cursor)
	sort := viper.GetString(cst.Sort)
	sortedBy := viper.GetString(cst.SortedBy)

	if query == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		query = args[0]
	}

	data, apiErr := userSearch(vcli, &userSearchParams{
		query:    query,
		limit:    limit,
		cursor:   cursor,
		sort:     sort,
		sortedBy: sortedBy,
	})
	if apiErr != nil {
		return apiErr
	}
	vcli.Out().WriteResponse(data, nil)
	return nil
}

func handleUserDeleteCmd(vcli vaultcli.CLI, args []string) error {
	userName := viper.GetString(cst.DataUsername)
	force := viper.GetBool(cst.Force)
	if userName == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		userName = args[0]
	}
	if userName == "" {
		return fmt.Errorf("error: must specify --username")
	}

	data, apiErr := userDelete(vcli, paths.ProcessResource(userName), force)
	if apiErr != nil {
		return apiErr
	}
	vcli.Out().WriteResponse(data, nil)
	return nil
}

func handleUserRestoreCmd(vcli vaultcli.CLI, args []string) error {
	userName := viper.GetString(cst.DataUsername)
	if userName == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		userName = args[0]
	}
	if userName == "" {
		return fmt.Errorf("error: must specify --username")
	}

	data, apiErr := userRestore(vcli, paths.ProcessResource(userName))
	if apiErr != nil {
		return apiErr
	}
	vcli.Out().WriteResponse(data, nil)
	return nil
}

func handleUserCreateCmd(vcli vaultcli.CLI, args []string) error {
	userName := viper.GetString(cst.DataUsername)
	password := viper.GetString(cst.DataPassword)
	provider := viper.GetString(cst.DataProvider)
	externalID := viper.GetString(cst.DataExternalID)

	if err := vaultcli.ValidateUsername(userName); err != nil {
		return fmt.Errorf("error: username %q is invalid: %w", userName, err)
	}

	isUserLocal := provider == "" && externalID == ""
	if password == "" && isUserLocal {
		return fmt.Errorf("error: must specify password for local users")
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
	if apiError != nil {
		return apiError
	}
	vcli.Out().WriteResponse(resp, nil)
	return nil
}

func handleUserUpdateCmd(vcli vaultcli.CLI, args []string) error {
	username := viper.GetString(cst.DataUsername)
	if username == "" {
		return fmt.Errorf("error: must specify --username")
	}

	displayNameExists := hasFlag(args, "--"+cst.DataDisplayname)
	passData := viper.GetString(cst.DataPassword)
	displayName := viper.GetString(cst.DataDisplayname)
	if passData == "" && !displayNameExists {
		return errMustSpecifyPasswordOrDisplayname
	}

	displayNameLen := len(displayName)
	if displayNameExists && (displayNameLen < 3 || displayNameLen > 100) {
		return errWrongDisplayName
	}

	body := &userUpdateRequest{Password: passData, DisplayName: displayName}
	resp, apiError := userUpdate(vcli, username, body)
	if apiError != nil {
		return apiError
	}
	vcli.Out().WriteResponse(resp, nil)
	return nil
}

// Wizards:

func handleUserCreateWizard(vcli vaultcli.CLI) int {
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

func handleUserUpdateWizard(vcli vaultcli.CLI) int {
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
	query    string
	limit    string
	cursor   string
	sort     string
	sortedBy string
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
	if p.sort != "" {
		queryParams[cst.Sort] = p.sort
	}
	if p.sortedBy != "" {
		queryParams["sortedBy"] = p.sortedBy
	}
	uri := paths.CreateURI(cst.NounUsers, queryParams)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}
