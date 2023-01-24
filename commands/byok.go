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
		SynopsisText: "Update AWS encryption key to a new",
		HelpText: `
Usage:
   • byok update
   • byok update --primary-key arn:aws:kms:us-west-1:012345678999:key/abcdef --secondary-key arn:aws:kms:us-east-1:012345678999:key/abcdef
`,
		FlagsPredictor: []*predictor.Params{
			{Name: cst.PrimaryKey, Usage: "Primary key (required)"},
			{Name: cst.SecondaryKey, Usage: "Secondary key"},
		},
		RunFunc:    handleBYOKUpdateCmd,
		WizardFunc: handleBYOKUpdateWizard,
	})
}

func handleBYOKUpdateCmd(vcli vaultcli.CLI, args []string) int {
	primaryKey := viper.GetString(cst.PrimaryKey)
	secondaryKey := viper.GetString(cst.SecondaryKey)
	if primaryKey == "" {
		err := errors.NewS("error: must specify " + cst.PrimaryKey)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}
	data, err := byokUpdate(vcli, primaryKey, secondaryKey)
	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func handleBYOKUpdateWizard(vcli vaultcli.CLI) int {
	qs := []*survey.Question{
		{
			Name:      "PrimaryKey",
			Prompt:    &survey.Input{Message: "Primary key:"},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:      "SecondaryKey",
			Prompt:    &survey.Input{Message: "Secondary key:"},
			Transform: vaultcli.SurveyTrimSpace,
		},
	}
	answers := struct {
		PrimaryKey   string
		SecondaryKey string
	}{}
	survErr := survey.Ask(qs, &answers)
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}
	data, err := byokUpdate(vcli, answers.PrimaryKey, answers.SecondaryKey)
	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func byokUpdate(vcli vaultcli.CLI, primaryKey, secondaryKey string) ([]byte, *errors.ApiError) {
	uri := paths.CreateURI("config/keys", nil)
	body := map[string]interface{}{
		"keyprovider":  "AWS",
		"primaryKey":   primaryKey,
		"secondaryKey": secondaryKey,
	}
	return vcli.HTTPClient().DoRequest(http.MethodPut, uri, body)
}
