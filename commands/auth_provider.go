package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

func GetAuthProviderCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounConfig, cst.NounAuthProvider},
		SynopsisText: "manage 3rd party authentication providers",
		HelpText:     fmt.Sprintf("Execute an action on an %s from %s", cst.NounAuthProvider, cst.ProductName),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataName, Shorthand: "n", Usage: fmt.Sprintf("Target %s to an %s", cst.Path, cst.NounAuthProvider)},
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
			return handleAuthProviderReadCmd(vcli, args)
		},
	})
}

func GetAuthProviderReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.Config, cst.NounAuthProvider, cst.Read},
		SynopsisText: fmt.Sprintf("%s %s %s (<name> | --name|-n)", cst.NounConfig, cst.NounAuthProvider, cst.Read),
		HelpText: `Read an authentication provider

Usage:
   • config auth-provider read aws-dev
   • config auth-provider read --name aws-dev
`,
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataName, Shorthand: "n", Usage: fmt.Sprintf("Target %s to an %s", cst.Path, cst.NounAuthProvider)},
			{Name: cst.Version, Usage: "List the current and last (n) versions"},
		},
		MinNumberArgs: 1,
		RunFunc:       handleAuthProviderReadCmd,
	})
}

func GetAuthProviderDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.Config, cst.NounAuthProvider, cst.Delete},
		SynopsisText: fmt.Sprintf("%s %s %s (<name> | --name|-n)", cst.Config, cst.NounAuthProvider, cst.Delete),
		HelpText: `Delete an authentication provider

Usage:
   • config auth-provider delete aws-dev
   • config auth-provider delete --name aws-dev --force
`,
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataName, Shorthand: "n", Usage: fmt.Sprintf("Target %s to an %s", cst.Path, cst.NounAuthProvider)},
			{Name: cst.Force, Usage: fmt.Sprintf("Immediately delete %s", cst.NounAuthProvider), ValueType: "bool"},
		},
		MinNumberArgs: 1,
		RunFunc:       handleAuthProviderDeleteCmd,
	})
}

func GetAuthProviderRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.Config, cst.NounAuthProvider, cst.Restore},
		SynopsisText: fmt.Sprintf("%s %s %s (<name> | --name|-n)", cst.Config, cst.NounAuthProvider, cst.Restore),
		HelpText: `Restore an authentication provider

Usage:
   • config auth-provider restore aws-dev
   • config auth-provider restore --name aws-dev
`,
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataName, Shorthand: "n", Usage: fmt.Sprintf("Target %s to an %s", cst.Path, cst.NounAuthProvider)},
		},
		MinNumberArgs: 1,
		RunFunc:       handleAuthProviderRestoreCmd,
	})
}

func GetAuthProviderCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.Config, cst.NounAuthProvider, cst.Create},
		SynopsisText: fmt.Sprintf("%s %s %s (<name> | --name|-n) (--type) ((--data|-d) | --aws-account-id | --azure-tenant-id | --gcp-project-id)", cst.NounConfig, cst.NounAuthProvider, cst.Create),
		HelpText: `Add an authentication provider

Usage:
   • config auth-provider create aws-dev --aws-account-id 11652944433808  --type aws
   • config auth-provider create --name azure-prod --azure-tenant-id 164543 --type azure
   • config auth-provider create --name GCP-prod --gcp-project-id test-proj --type gcp
   • config auth-provider create --data @/tmp/data.json

GCP GCE metadata auth provider can be created in the command line, but a GCP Service Account must be done using a file.
See the Authentication:GCP portion of the documentation.
`,
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("%s to be stored in an auth provider. Prefix with '@' to denote filepath", strings.Title(cst.Data)), Predictor: predictor.NewPrefixFilePredictor("*")},
			{Name: cst.DataType, Usage: "Auth provider type (azure,aws,gcp,thycoticone)", Predictor: predictor.AuthProviderTypePredictor{}},
			{Name: cst.DataName, Shorthand: "n", Usage: "Auth provider friendly name"},
			{Name: cst.DataTenantID, Usage: "Azure Tenant ID"},
			{Name: cst.DataAccountID, Usage: "AWS Account ID"},
			{Name: cst.DataProjectID, Usage: "GCP Project ID"},
			{Name: cst.ThyOneAuthClientBaseUri, Usage: "Thycotic One base URI"},
			{Name: cst.ThyOneAuthClientID, Usage: "Thycotic One client ID"},
			{Name: cst.ThyOneAuthClientSecret, Usage: "Thycotic One client secret"},
			{Name: cst.SendWelcomeEmail, Usage: "Whether to send welcome email for thycotic-one users linked to the auth provider (true or false)"},
		},
		RunFunc:    handleAuthProviderCreate,
		WizardFunc: handleAuthProviderCreateWizard,
	})
}

func GetAuthProviderUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.Config, cst.NounAuthProvider, cst.Update},
		SynopsisText: fmt.Sprintf("%s %s %s (<name> | --name|-n) (--type) ((--data|-d) | --aws-account-id | --azure-tenant-id | --gcp-project-id)", cst.NounConfig, cst.NounAuthProvider, cst.Update),
		HelpText: `Update an authentication provider

Usage:
   • config auth-provider update aws-dev --aws-account-id 11652944433808  --type aws
   • config auth-provider update --name azure-prod --azure-tenant-id 164543 --type azure
   • config auth-provider update --name GCP-prod --gcp-project-id test-proj --type gcp
   • config auth-provider update --data @/tmp/data.json
`,
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("%s to be stored in an auth provider. Prefix with '@' to denote filepath", strings.Title(cst.Data)), Predictor: predictor.NewPrefixFilePredictor("*")},
			{Name: cst.DataType, Usage: "Auth provider type (azure,aws,gcp,thycoticone)", Predictor: predictor.AuthProviderTypePredictor{}},
			{Name: cst.DataName, Shorthand: "n", Usage: "Auth provider friendly name"},
			{Name: cst.DataTenantID, Usage: "Azure Tenant ID"},
			{Name: cst.DataAccountID, Usage: "AWS Account ID"},
			{Name: cst.DataProjectID, Usage: "GCP Project ID"},
			{Name: cst.ThyOneAuthClientBaseUri, Usage: "Thycotic One base URI"},
			{Name: cst.ThyOneAuthClientID, Usage: "Thycotic One client ID"},
			{Name: cst.ThyOneAuthClientSecret, Usage: "Thycotic One client secret"},
			{Name: cst.SendWelcomeEmail, Usage: "Whether to send welcome email for thycotic-one users linked to the auth provider (true or false)"},
		},
		RunFunc:    handleAuthProviderUpdate,
		WizardFunc: handleAuthProviderUpdateWizard,
	})
}

func GetAuthProviderEditCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.Config, cst.NounAuthProvider, cst.Edit},
		SynopsisText: fmt.Sprintf("%s %s %s (<name> | --name|-n)", cst.NounConfig, cst.NounAuthProvider, cst.Edit),
		HelpText: `Edit an authentication provider

Usage:
   • config auth-provider edit aws-dev
   • config auth-provider edit --name aws-dev
`,
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataName, Shorthand: "n", Usage: fmt.Sprintf("Target %s to an %s", cst.Path, cst.NounAuthProvider)},
		},
		MinNumberArgs: 1,
		RunFunc:       handleAuthProviderEdit,
	})
}

func GetAuthProviderRollbackCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.Config, cst.NounAuthProvider, cst.Rollback},
		SynopsisText: fmt.Sprintf("%s %s %s (<name> | --name|-n)", cst.NounConfig, cst.NounAuthProvider, cst.Rollback),
		HelpText: `Rollback an authentication provider

Usage:
   • config auth-provider rollback aws-dev
   • config auth-provider rollback --name aws-dev --version 1
`,
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataName, Shorthand: "n", Usage: fmt.Sprintf("Target %s to an %s", cst.Path, cst.NounAuthProvider)},
			{Name: cst.Version, Usage: "The version to which to rollback"},
		},
		MinNumberArgs: 1,
		RunFunc:       handleAuthProviderRollbackCmd,
	})
}

func GetAuthProviderSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.Config, cst.NounAuthProvider, cst.Search},
		SynopsisText: fmt.Sprintf("%s %s %s (<query> | --query)", cst.NounConfig, cst.NounAuthProvider, cst.Search),
		HelpText: `Search for an authentication provider

Usage:
   • config auth-provider search aws-dev
   • config auth-provider search --query aws-dev
`,
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Query, Shorthand: "q", Usage: fmt.Sprintf("Filter %s of items to fetch (required)", strings.Title(cst.Query))},
			{Name: cst.Limit, Shorthand: "l", Usage: cst.LimitHelpMessage},
			{Name: cst.Cursor, Usage: cst.CursorHelpMessage},
		},
		RunFunc: handleAuthProviderSearchCmd,
	})
}

func handleAuthProviderReadCmd(vcli vaultcli.CLI, args []string) int {
	name := getAuthProviderName(args)
	if name == "" {
		return cli.RunResultHelp
	}
	name = paths.ProcessResource(name)
	ver := viper.GetString(cst.Version)
	if strings.TrimSpace(ver) != "" {
		name = fmt.Sprint(name, "/", cst.Version, "/", ver)
	}

	data, apiErr := authProviderRead(vcli, name)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleAuthProviderDeleteCmd(vcli vaultcli.CLI, args []string) int {
	name := getAuthProviderName(args)
	if name == "" {
		return cli.RunResultHelp
	}
	data, apiErr := authProviderDelete(vcli, name, viper.GetBool(cst.Force))
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleAuthProviderRestoreCmd(vcli vaultcli.CLI, args []string) int {
	name := getAuthProviderName(args)
	if name == "" {
		return cli.RunResultHelp
	}
	data, apiErr := authProviderRestore(vcli, name)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleAuthProviderCreate(vcli vaultcli.CLI, args []string) int {
	var model authProviderCreateRequest
	data := viper.GetString(cst.Data)
	if data != "" {
		if err := json.Unmarshal([]byte(data), &model); err != nil {
			vcli.Out().Fail(err)
			return utils.GetExecStatus(err)
		}
	} else {
		model = authProviderCreateRequest{
			Name: getAuthProviderName(args),
			Type: viper.GetString(cst.DataType),
			Properties: AuthProviderProperties{
				AccountID:    viper.GetString(cst.DataAccountID),
				TenantID:     viper.GetString(cst.DataTenantID),
				ProjectID:    viper.GetString(cst.DataProjectID),
				BaseURI:      viper.GetString(cst.ThyOneAuthClientBaseUri),
				ClientID:     viper.GetString(cst.ThyOneAuthClientID),
				ClientSecret: viper.GetString(cst.ThyOneAuthClientSecret),
			},
		}
	}
	if model.Name == "" {
		return cli.RunResultHelp
	}

	if model.Type == cst.ThyOne {
		sendWelcomeEmail, err := strconv.ParseBool(viper.GetString(cst.SendWelcomeEmail))
		if err == nil {
			model.Properties.SendWelcomeEmail = &sendWelcomeEmail
		}
	}

	resp, apiErr := authProviderCreate(vcli, &model)
	vcli.Out().WriteResponse(resp, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleAuthProviderUpdate(vcli vaultcli.CLI, args []string) int {
	name := getAuthProviderName(args)
	if name == "" {
		return cli.RunResultHelp
	}

	var model authProviderUpdateRequest

	data := viper.GetString(cst.Data)
	if data != "" {
		if err := json.Unmarshal([]byte(data), &model); err != nil {
			vcli.Out().Fail(err)
			return utils.GetExecStatus(err)
		}
	} else {
		model = authProviderUpdateRequest{
			Type: viper.GetString(cst.DataType),
			Properties: AuthProviderProperties{
				AccountID:    viper.GetString(cst.DataAccountID),
				TenantID:     viper.GetString(cst.DataTenantID),
				ProjectID:    viper.GetString(cst.DataProjectID),
				BaseURI:      viper.GetString(cst.ThyOneAuthClientBaseUri),
				ClientID:     viper.GetString(cst.ThyOneAuthClientID),
				ClientSecret: viper.GetString(cst.ThyOneAuthClientSecret),
			},
		}
	}

	if model.Type == cst.ThyOne {
		sendWelcomeEmail, parseErr := strconv.ParseBool(viper.GetString(cst.SendWelcomeEmail))
		if parseErr == nil {
			model.Properties.SendWelcomeEmail = &sendWelcomeEmail
		}
	}

	resp, apiErr := authProviderUpdate(vcli, name, &model)
	vcli.Out().WriteResponse(resp, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleAuthProviderRollbackCmd(vcli vaultcli.CLI, args []string) int {
	name := getAuthProviderName(args)
	if name == "" {
		return cli.RunResultHelp
	}
	version := viper.GetString(cst.Version)

	// If version is not provided, get the current auth-provider item and parse the version from it.
	// Submit a request for a version that's previous relative to the one found.
	if version == "" {
		data, apiErr := authProviderRead(vcli, name)
		if apiErr != nil {
			vcli.Out().WriteResponse(data, apiErr)
			return utils.GetExecStatus(apiErr)
		}

		v, err := utils.GetPreviousVersion(data)
		if err != nil {
			vcli.Out().Fail(err)
			return 1
		}
		version = v
	}

	data, apiErr := authProviderRollback(vcli, name, version)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleAuthProviderSearchCmd(vcli vaultcli.CLI, args []string) int {
	query := viper.GetString(cst.Query)
	limit := viper.GetString(cst.Limit)
	cursor := viper.GetString(cst.Cursor)
	if query == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		query = args[0]
	}

	data, apiErr := authProviderSearch(vcli,
		&authProviderSearchParams{query: query, limit: limit, cursor: cursor})
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleAuthProviderEdit(vcli vaultcli.CLI, args []string) int {
	name := getAuthProviderName(args)
	if name == "" {
		return cli.RunResultHelp
	}

	data, apiErr := authProviderRead(vcli, paths.ProcessResource(name))
	if apiErr != nil {
		vcli.Out().WriteResponse(data, apiErr)
		return utils.GetExecStatus(apiErr)
	}

	saveFunc := func(data []byte) (resp []byte, err *errors.ApiError) {
		var model authProviderUpdateRequest
		if mErr := json.Unmarshal(data, &model); mErr != nil {
			return nil, errors.New(mErr).Grow("invalid format for auth provider")
		}
		_, apiErr := authProviderUpdate(vcli, paths.ProcessResource(name), &model)
		return nil, apiErr
	}

	data, apiErr = vcli.Edit(data, saveFunc)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

// Wizards:

func handleAuthProviderCreateWizard(vcli vaultcli.CLI) int {
	qs := []*survey.Question{
		{
			Name:   "Name",
			Prompt: &survey.Input{Message: "Auth provider name:"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if len(answer) == 0 {
					return errors.NewS("Value is required")
				}
				_, apiError := authProviderRead(vcli, answer)
				if apiError == nil {
					return errors.NewS("An authentication provider with this name already exists.")
				}
				return nil
			},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name: "Type",
			Prompt: &survey.Select{
				Message: "Auth provider type:",
				Options: []string{"AWS", "Azure", "GCP", "Thycotic One"},
			},
		},
	}
	provider := &authProviderCreateRequest{}
	survErr := survey.Ask(qs, provider)
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}

	var props *AuthProviderProperties
	var err error
	switch provider.Type {
	case "AWS":
		props, err = authProviderAWSWizard()
	case "Azure":
		props, err = authProviderAzureWizard()
	case "GCP":
		props, err = authProviderGCPWizard()
	case "Thycotic One":
		props, err = authProviderThycoticOneWizard()
	default:
		return 1 // Unsupported auth provider type. Should be unreachable.
	}

	if err != nil {
		vcli.Out().WriteResponse(nil, errors.New(err))
		return utils.GetExecStatus(err)
	}

	if provider.Type == "Thycotic One" {
		provider.Type = cst.ThyOne
	}
	provider.Properties = *props

	resp, apiErr := authProviderCreate(vcli, provider)
	vcli.Out().WriteResponse(resp, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleAuthProviderUpdateWizard(vcli vaultcli.CLI) int {
	var name string
	namePrompt := &survey.Input{Message: "Auth provider name:"}
	survErr := survey.AskOne(namePrompt, &name, survey.WithValidator(vaultcli.SurveyRequired))
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}
	name = strings.TrimSpace(name)

	resp, apiErr := authProviderRead(vcli, name)
	if apiErr != nil {
		vcli.Out().WriteResponse(resp, apiErr)
		return utils.GetExecStatus(apiErr)
	}
	respData := struct {
		Type string `json:"type"`
	}{}
	jsonErr := json.Unmarshal(resp, &respData)
	if apiErr != nil {
		err := errors.New(jsonErr).Grow("Failed to determine type of the provider")
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	} else if respData.Type == "" {
		err := errors.NewS("Failed to determine type of the provider")
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	var props *AuthProviderProperties
	var err error
	switch respData.Type {
	case "aws":
		props, err = authProviderAWSWizard()
	case "azure":
		props, err = authProviderAzureWizard()
	case "gcp":
		props, err = authProviderGCPWizard()
	case cst.ThyOne:
		props, err = authProviderThycoticOneWizard()
	default:
		err = fmt.Errorf("Unsupported auth provider type: %s", respData.Type)
	}

	if err != nil {
		vcli.Out().WriteResponse(nil, errors.New(err))
		return utils.GetExecStatus(err)
	}

	provider := &authProviderUpdateRequest{
		Type:       respData.Type,
		Properties: *props,
	}

	resp, apiErr = authProviderUpdate(vcli, name, provider)
	vcli.Out().WriteResponse(resp, apiErr)
	return utils.GetExecStatus(apiErr)
}

func authProviderAWSWizard() (*AuthProviderProperties, error) {
	var accountID string
	accIDPrompt := &survey.Input{Message: "AWS account ID:"}
	survErr := survey.AskOne(accIDPrompt, &accountID, survey.WithValidator(vaultcli.SurveyRequired))
	if survErr != nil {
		return nil, survErr
	}
	return &AuthProviderProperties{AccountID: strings.TrimSpace(accountID)}, nil
}

func authProviderAzureWizard() (*AuthProviderProperties, error) {
	var tenantID string
	tenantIDPrompt := &survey.Input{Message: "Azure tenant ID:"}
	survErr := survey.AskOne(tenantIDPrompt, &tenantID, survey.WithValidator(vaultcli.SurveyRequired))
	if survErr != nil {
		return nil, survErr
	}
	return &AuthProviderProperties{TenantID: strings.TrimSpace(tenantID)}, nil
}

func authProviderGCPWizard() (*AuthProviderProperties, error) {
	var dataFilePath string
	pathPrompt := &survey.Input{Message: "Path to data file with provider properties:"}
	survErr := survey.AskOne(pathPrompt, &dataFilePath, survey.WithValidator(vaultcli.SurveyRequired))
	if survErr != nil {
		return nil, survErr
	}
	dataFilePath = strings.TrimSpace(dataFilePath)
	dataBytes, err := os.ReadFile(dataFilePath)
	if err != nil {
		return nil, err
	}
	dataMap := struct {
		Properties *AuthProviderProperties `json:"properties"`
	}{}
	err = json.Unmarshal([]byte(dataBytes), &dataMap)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse properties file: %v", err)
	}
	if dataMap.Properties == nil {
		return nil, fmt.Errorf("Missing properties field in the %s file", dataFilePath)
	}

	return dataMap.Properties, nil
}

func authProviderThycoticOneWizard() (*AuthProviderProperties, error) {
	qs := []*survey.Question{
		{
			Name:      "BaseURL",
			Prompt:    &survey.Input{Message: "Base URL:"},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "ClientID",
			Prompt:    &survey.Input{Message: "Client ID:"},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "ClientSecret",
			Prompt:    &survey.Input{Message: "Client secret:"},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:   "WelcomeEmail",
			Prompt: &survey.Confirm{Message: "Send welcome email:", Default: false},
		},
	}
	answers := struct {
		BaseURL      string
		ClientID     string
		ClientSecret string
		WelcomeEmail bool
	}{}
	survErr := survey.Ask(qs, &answers)
	if survErr != nil {
		return nil, survErr
	}

	return &AuthProviderProperties{
		BaseURI:          answers.BaseURL,
		ClientID:         answers.ClientID,
		ClientSecret:     answers.ClientSecret,
		SendWelcomeEmail: &answers.WelcomeEmail,
	}, nil
}

// Helpers:

func getAuthProviderName(args []string) string {
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name = args[0]
	}
	return name
}

type authProvider struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func listAuthProviders(vcli vaultcli.CLI) ([]*authProvider, *errors.ApiError) {
	data, err := authProviderSearch(vcli, &authProviderSearchParams{limit: "500"})
	if err != nil {
		return nil, err
	}
	resp := struct {
		Data   []*authProvider `json:"data"`
		Cursor string          `json:"cursor"`
	}{}
	jsonErr := json.Unmarshal(data, &resp)
	if jsonErr != nil {
		return nil, errors.New(jsonErr)
	}
	if resp.Cursor != "" {
		log.Printf("[warning] Not all authentication providers were retrieved.")
	}
	return resp.Data, nil
}

// API callers:

type AuthProviderProperties struct {
	AccountID        string `json:"accountId,omitempty"`
	TenantID         string `json:"tenantId,omitempty"`
	ProjectID        string `json:"projectId,omitempty"`
	Default          bool   `json:"default,omitempty"`
	ClientEmail      string `json:"clientEmail,omitempty"`
	PrivateKey       string `json:"privateKey,omitempty"`
	PrivateKeyID     string `json:"privateKeyId,omitempty"`
	TokenURI         string `json:"tokenUri,omitempty"`
	ClientID         string `json:"clientId,omitempty"`
	ClientSecret     string `json:"clientSecret,omitempty"`
	Type             string `json:"type,omitempty"`
	BaseURI          string `json:"baseUri,omitempty"`
	UsernameClaim    string `json:"usernameClaim,omitempty"`
	SendWelcomeEmail *bool  `json:"sendWelcomeEmail,omitempty"`
}

type authProviderCreateRequest struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Properties AuthProviderProperties `json:"properties"`
}

func authProviderCreate(vcli vaultcli.CLI, body *authProviderCreateRequest) ([]byte, *errors.ApiError) {
	baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")
	uri := paths.CreateResourceURI(baseType, "", "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
}

type authProviderUpdateRequest struct {
	Type       string                 `json:"type"`
	Properties AuthProviderProperties `json:"properties"`
}

func authProviderUpdate(vcli vaultcli.CLI, name string, body *authProviderUpdateRequest) ([]byte, *errors.ApiError) {
	baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")
	uri := paths.CreateResourceURI(baseType, name, "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPut, uri, body)
}

func authProviderRead(vcli vaultcli.CLI, name string) ([]byte, *errors.ApiError) {
	baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")
	uri := paths.CreateResourceURI(baseType, name, "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}

func authProviderRollback(vcli vaultcli.CLI, name string, version string) ([]byte, *errors.ApiError) {
	baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")
	path := fmt.Sprintf("%s/rollback/%s", name, version)
	uri := paths.CreateResourceURI(baseType, path, "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPut, uri, nil)
}

func authProviderDelete(vcli vaultcli.CLI, name string, force bool) ([]byte, *errors.ApiError) {
	query := map[string]string{"force": strconv.FormatBool(force)}
	baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")
	uri := paths.CreateResourceURI(baseType, name, "", true, query)
	return vcli.HTTPClient().DoRequest(http.MethodDelete, uri, nil)
}

func authProviderRestore(vcli vaultcli.CLI, name string) ([]byte, *errors.ApiError) {
	baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")
	uri := paths.CreateResourceURI(baseType, name, "/restore", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPut, uri, nil)
}

type authProviderSearchParams struct {
	query  string
	limit  string
	cursor string
}

func authProviderSearch(vcli vaultcli.CLI, p *authProviderSearchParams) ([]byte, *errors.ApiError) {
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
	baseType := strings.Join([]string{cst.Config, cst.NounAuth}, "/")
	uri := paths.CreateResourceURI(baseType, "", "", false, queryParams)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}
