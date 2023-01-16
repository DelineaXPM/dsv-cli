package cmd

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/cli"
	"github.com/spf13/viper"

	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/internal/predictor"
	"github.com/DelineaXPM/dsv-cli/paths"
	"github.com/DelineaXPM/dsv-cli/utils"
	"github.com/DelineaXPM/dsv-cli/vaultcli"
)

func GetRoleCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounRole},
		SynopsisText: "Manage roles",
		HelpText: fmt.Sprintf(`Execute an action on a role in %[1]s

Usage:
   • role %[2]s
   • role --name %[2]s
`, cst.ProductName, cst.ExampleRoleName),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataName, Shorthand: "n", Usage: "Name of the role"},
			{Name: cst.Version, Usage: "List the current and last (n) versions"},
		},
		MinNumberArgs: 1,
		RunFunc: func(vcli vaultcli.CLI, args []string) int {
			name := viper.GetString(cst.DataName)
			if name == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
				name = args[0]
			}
			if name == "" {
				return cli.RunResultHelp
			}
			return handleRoleReadCmd(vcli, args)
		},
	})
}

func GetRoleReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounRole, cst.Read},
		SynopsisText: "read (<name> | --name|-n)",
		HelpText: fmt.Sprintf(`Read a role in %[1]s

Usage:
   • role read %[2]s
   • role read --name %[2]s
   • role read --name %[2]s  --version
`, cst.ProductName, cst.ExampleRoleName),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataName, Shorthand: "n", Usage: "Name of the role"},
			{Name: cst.Version, Usage: "List the current and last (n) versions"},
		},
		MinNumberArgs: 1,
		RunFunc:       handleRoleReadCmd,
	})
}

func GetRoleSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounRole, cst.Search},
		SynopsisText: "search (<query> | --query)",
		HelpText: fmt.Sprintf(`Search for a role from %[1]s

Usage:
   • role search %[2]s
   • role search --query %[2]s
`, cst.ProductName, cst.ExampleUserSearch),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Query, Shorthand: "q", Usage: "Query of roles to fetch (optional)"},
			{Name: cst.Limit, Shorthand: "l", Usage: cst.LimitHelpMessage},
			{Name: cst.Cursor, Usage: cst.CursorHelpMessage},
			{Name: cst.Sort, Usage: cst.SortHelpMessage, Default: "desc"},
			{Name: cst.SortedBy, Usage: "Sort by name, created or lastModified field (optional)", Default: "lastModified"},
		},
		RunFunc: handleRoleSearchCmd,
	})
}

func GetRoleDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounRole, cst.Delete},
		SynopsisText: "delete (<name> | --name|-n)",
		HelpText: fmt.Sprintf(`Delete a role in %[1]s

Usage:
   • role delete %[2]s
   • role delete --name %[2]s --force
`, cst.ProductName, cst.ExampleRoleName),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataName, Shorthand: "n", Usage: "Name of the role"},
			{Name: cst.Force, Usage: "Immediately delete the role", ValueType: "bool"},
		},
		MinNumberArgs: 1,
		RunFunc:       handleRoleDeleteCmd,
	})
}

func GetRoleRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounRole, cst.Restore},
		SynopsisText: "restore (<name> | --name|-n)",
		HelpText: fmt.Sprintf(`Restore a deleted role in %[1]s

Usage:
   • role restore %[2]s
`, cst.ProductName, cst.ExampleRoleName),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataName, Shorthand: "n", Usage: "Name of the role"},
		},
		MinNumberArgs: 1,
		RunFunc:       handleRoleRestoreCmd,
	})
}

func GetRoleUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounRole, cst.Update},
		SynopsisText: "update (<name> | --name|-n) --desc",
		HelpText: fmt.Sprintf(`Update a role in %[1]s

Usage:
   • role update --name %[2]s --desc "msa for prod gcp"
`, cst.ProductName, cst.ExampleRoleName),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataName, Shorthand: "n", Usage: "Name of the role (required)"},
			{Name: cst.DataDescription, Usage: "Description of the role"},
		},
		RunFunc:    handleRoleUpdateCmd,
		WizardFunc: handleRoleUpdateWizard,
	})
}

func GetRoleCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounRole, cst.Create},
		SynopsisText: "create (<name> | --name|-n) --provider --external-id --desc",
		HelpText: fmt.Sprintf(`Create a role in %[1]s

Usage:
   • role create --name %[2]s --external-id msa-1@happy-emu-172.iam.gsa.com --provider ProdGcp --desc "msa for prod gcp"
`, cst.ProductName, cst.ExampleRoleName),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataName, Shorthand: "n", Usage: "Name of the role (required)"},
			{Name: cst.DataDescription, Usage: "Description of the role"},
			{Name: cst.DataExternalID, Usage: "External Id for the role"},
			{Name: cst.DataProvider, Usage: "Provider for the role"},
		},
		RunFunc:    handleRoleCreateCmd,
		WizardFunc: handleRoleCreateWizard,
	})
}

func handleRoleReadCmd(vcli vaultcli.CLI, args []string) int {
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name = args[0]
	}
	if name == "" {
		err := errors.NewS("error: must specify " + cst.DataName)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	name = paths.ProcessResource(name)
	version := viper.GetString(cst.Version)
	if strings.TrimSpace(version) != "" {
		name = fmt.Sprint(name, "/", cst.Version, "/", version)
	}

	data, apiErr := roleRead(vcli, name)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleRoleSearchCmd(vcli vaultcli.CLI, args []string) int {
	query := viper.GetString(cst.Query)
	limit := viper.GetString(cst.Limit)
	cursor := viper.GetString(cst.Cursor)
	sort := viper.GetString(cst.Sort)
	sortedBy := viper.GetString(cst.SortedBy)

	if query == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		query = args[0]
	}

	data, apiErr := roleSearch(vcli, &roleSearchParams{
		query:    query,
		limit:    limit,
		cursor:   cursor,
		sort:     sort,
		sortedBy: sortedBy,
	})
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleRoleDeleteCmd(vcli vaultcli.CLI, args []string) int {
	force := viper.GetBool(cst.Force)
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name = args[0]
	}
	if name == "" {
		err := errors.NewS("error: must specify " + cst.DataName)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, apiErr := roleDelete(vcli, paths.ProcessResource(name), force)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleRoleRestoreCmd(vcli vaultcli.CLI, args []string) int {
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name = args[0]
	}
	if name == "" {
		err := errors.NewS("error: must specify " + cst.DataName)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, apiErr := roleRestore(vcli, paths.ProcessResource(name))
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleRoleCreateCmd(vcli vaultcli.CLI, args []string) int {
	name := viper.GetString(cst.DataName)
	if name == "" {
		err := errors.NewS("error: must specify " + cst.DataName)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}
	if err := vaultcli.ValidateName(name); err != nil {
		vcli.Out().FailF("error: role name %q is invalid: %v", name, err)
		return utils.GetExecStatus(err)
	}

	role := &roleCreateRequest{
		Name:        name,
		Description: viper.GetString(cst.DataDescription),
		ExternalID:  viper.GetString(cst.DataExternalID),
		Provider:    viper.GetString(cst.DataProvider),
	}
	if (role.Provider != "" && role.ExternalID == "") || (role.Provider == "" && role.ExternalID != "") {
		err := errors.NewS("error: must specify both provider and external ID for third-party roles")
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, err := roleCreate(vcli, role)
	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func handleRoleUpdateCmd(vcli vaultcli.CLI, args []string) int {
	name := viper.GetString(cst.DataName)
	if name == "" {
		err := errors.NewS("error: must specify " + cst.DataName)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, apiErr := roleUpdate(vcli, name, viper.GetString(cst.DataDescription))
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

// Wizards:

func handleRoleCreateWizard(vcli vaultcli.CLI) int {
	providers, apiError := listAuthProviders(vcli)
	if apiError != nil {
		httpResp := apiError.HttpResponse()
		// If API returns "403 Forbidden" still allow to try to create a role with "local" as provider.
		if httpResp == nil || httpResp.StatusCode != http.StatusForbidden {
			vcli.Out().FailS("Failed to get available auth providers.")
			return utils.GetExecStatus(apiError)
		}
	}

	qs := []*survey.Question{
		{
			Name:   "Name",
			Prompt: &survey.Input{Message: "Role name:"},
			Validate: func(ans interface{}) error {
				answer := ans.(string)
				answer = strings.TrimSpace(answer)
				if len(answer) == 0 {
					return errors.NewS("Value is required")
				}
				if err := vaultcli.ValidateName(answer); err != nil {
					return err
				}
				_, apiError := roleRead(vcli, answer)
				if apiError == nil {
					return errors.NewS("A role with this name already exists.")
				}
				return nil
			},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "Description",
			Prompt:    &survey.Input{Message: "Description of the role:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
	}

	if len(providers) != 0 {
		providersOpts := []string{"local"}
		for _, p := range providers {
			if p.Type == cst.ThyOne {
				// Skip thycoticone - roles cannot have it as a provider.
				continue
			}
			providersOpts = append(providersOpts, fmt.Sprintf("%s [type: %s]", p.Name, p.Type))
		}
		if len(providersOpts) != 1 {
			qs = append(qs, &survey.Question{
				Name: "Provider",
				Prompt: &survey.Select{
					Message: "Select provider:",
					Options: providersOpts,
				},
			})
		}
	}

	answers := struct {
		Name        string
		Description string
		Provider    string
	}{}
	survErr := survey.Ask(qs, &answers)
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}

	var provider, externalID string

	if answers.Provider != "" && answers.Provider != "local" {
		provider = strings.Split(answers.Provider, " [type: ")[0]

		externalIDPrompt := &survey.Input{Message: "External ID:"}
		survErr := survey.AskOne(externalIDPrompt, &externalID, survey.WithValidator(vaultcli.SurveyRequired))
		if survErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(survErr))
			return utils.GetExecStatus(survErr)
		}
	}

	role := &roleCreateRequest{
		Name:        answers.Name,
		Description: answers.Description,
		Provider:    provider,
		ExternalID:  strings.TrimSpace(externalID),
	}
	data, apiErr := roleCreate(vcli, role)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleRoleUpdateWizard(vcli vaultcli.CLI) int {
	var roleName string
	namePrompt := &survey.Input{Message: "Role name:"}
	survErr := survey.AskOne(namePrompt, &roleName, survey.WithValidator(vaultcli.SurveyRequired))
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}

	data, apiError := roleRead(vcli, roleName)
	if apiError != nil {
		httpResp := apiError.HttpResponse()
		if httpResp == nil || httpResp.StatusCode != http.StatusForbidden {
			vcli.Out().Fail(apiError)
			return utils.GetExecStatus(apiError)
		}

		var confirm bool
		confirmPrompt := &survey.Confirm{
			Message: "You are not allowed to read role with that name. Do you want to continue?",
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
	}

	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: "Would you like to update the description?",
		Default: false,
	}
	survErr = survey.AskOne(confirmPrompt, &confirm)
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}
	if !confirm {
		return 0
	}

	var description string
	if confirm {
		descriptionPrompt := &survey.Input{Message: "Description of the role:"}
		survErr := survey.AskOne(descriptionPrompt, &description)
		if survErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(survErr))
			return utils.GetExecStatus(survErr)
		}
	}

	data, apiErr := roleUpdate(vcli, roleName, description)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

// API callers:

type roleCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Provider    string `json:"provider"`
	ExternalID  string `json:"externalId"`
}

func roleCreate(vcli vaultcli.CLI, body *roleCreateRequest) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounRoles, "", "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, &body)
}

func roleRead(vcli vaultcli.CLI, name string) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounRoles, name, "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}

func roleUpdate(vcli vaultcli.CLI, name string, desc string) ([]byte, *errors.ApiError) {
	body := map[string]string{"description": desc}
	uri := paths.CreateResourceURI(cst.NounRoles, name, "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPut, uri, &body)
}

func roleDelete(vcli vaultcli.CLI, name string, force bool) ([]byte, *errors.ApiError) {
	query := map[string]string{"force": strconv.FormatBool(force)}
	uri := paths.CreateResourceURI(cst.NounRoles, name, "", true, query)
	return vcli.HTTPClient().DoRequest(http.MethodDelete, uri, nil)
}

func roleRestore(vcli vaultcli.CLI, name string) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounRoles, name, "/restore", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPut, uri, nil)
}

type roleSearchParams struct {
	query    string
	limit    string
	cursor   string
	sort     string
	sortedBy string
}

func roleSearch(vcli vaultcli.CLI, p *roleSearchParams) ([]byte, *errors.ApiError) {
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
	uri := paths.CreateURI(cst.NounRoles, queryParams)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}
