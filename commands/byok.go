package cmd

import (
	"net/http"

	"github.com/AlecAivazis/survey/v2"
	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/internal/predictor"
	"github.com/DelineaXPM/dsv-cli/paths"
	"github.com/DelineaXPM/dsv-cli/utils"
	"github.com/DelineaXPM/dsv-cli/vaultcli"
	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

func GetBYOKCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounBYOK},
		SynopsisText: "Manage encryption key",
		HelpText: `Bring your own encryption key for DevOps Secrets Vault

Usage:
	• byok update
`,
		NoConfigRead: true,
		NoPreAuth:    true,
	})
}

func GetBYOKUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounBYOK, cst.Update},
		SynopsisText: "Update encryption key to a new one",
		HelpText: `
Usage:
   • byok update
   • byok update --provider AWS --primary-key arn:aws:kms:us-west-1:012345678999:key/abcdef --secondary-key arn:aws:kms:us-east-1:012345678999:key/abcdef
`,
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataProvider, Usage: "Key provider AWS/GCP (required)"},
			{Name: cst.PrimaryKey, Usage: "Primary key (required)"},
			{Name: cst.SecondaryKey, Usage: "Secondary key (required)"},
		},
		RunFunc:    handleBYOKUpdateCmd,
		WizardFunc: handleBYOKUpdateWizard,
	})
}

func handleBYOKUpdateCmd(vcli vaultcli.CLI, args []string) int {
	provider := viper.GetString(cst.DataProvider)
	primaryKey := viper.GetString(cst.PrimaryKey)
	secondaryKey := viper.GetString(cst.SecondaryKey)
	if provider != "AWS" && provider != "GCP" {
		err := errors.NewS("error: provider must be specified from list: AWS, GCP")
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}
	if primaryKey == "" {
		err := errors.NewS("error: must specify " + cst.DataProvider)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}
	if secondaryKey == "" {
		err := errors.NewS("error: must specify " + cst.DataProvider)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}
	data, err := byokUpdate(vcli, provider, primaryKey, secondaryKey)
	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func handleBYOKUpdateWizard(vcli vaultcli.CLI) int {
	qs := []*survey.Question{
		{
			Name:   "Provider",
			Prompt: &survey.Select{Message: "Provider:", Options: []string{"AWS", "GCP"}},
		},
		{
			Name:      "PrimaryKey",
			Prompt:    &survey.Input{Message: "Primary key:"},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "SecondaryKey",
			Prompt:    &survey.Input{Message: "Secondary key:"},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
	}
	answers := struct {
		Provider     string
		PrimaryKey   string
		SecondaryKey string
	}{}
	survErr := survey.Ask(qs, &answers)
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}
	data, err := byokUpdate(vcli, answers.Provider, answers.PrimaryKey, answers.SecondaryKey)
	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func byokUpdate(vcli vaultcli.CLI, provider, primaryKey, secondaryKey string) ([]byte, *errors.ApiError) {
	uri := paths.CreateURI("config/keys", nil)
	body := map[string]interface{}{
		"keyprovider":  provider,
		"primaryKey":   primaryKey,
		"secondaryKey": secondaryKey,
	}
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
}
