package cmd

import (
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

func GetBreakGlassCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounBreakGlass},
		SynopsisText: "Manage Break-Glass setup",
		HelpText:     "Initiate restoration of admin users",
		NoConfigRead: true,
		NoPreAuth:    true,
	})
}

func GetBreakGlassGetStatusCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounBreakGlass, cst.Status},
		SynopsisText: "Check whether Break-Glass feature is set up for the tenant",
		HelpText: `
Usage:
   • breakglass status
`,
		RunFunc: handleBreakGlassGetStatusCmd,
	})
}

func GetBreakGlassGenerateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounBreakGlass, cst.Generate},
		SynopsisText: "Generate and store admin secret and new admins' shares",
		HelpText: `
Usage:
   • breakglass generate --new-admins 'newAdminUsername1,newAdminUsername2' --min-number-of-shares 2
`,
		FlagsPredictor: []*predictor.Params{
			{Name: cst.NewAdmins, Usage: "New admins list (required)"},
			{Name: cst.MinNumberOfShares, Usage: "Minimum number of shares to apply (required)"},
		},
		RunFunc:    handleBreakGlassGenerateCmd,
		WizardFunc: handleBreakGlassGenerateWizard,
	})
}

func GetBreakGlassApplyCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounBreakGlass, cst.Apply},
		SynopsisText: "Apply shares and break glass",
		HelpText: `
Usage:
   • breakglass apply --shares '{share1},{share2},...,{shareN}'
`,
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Shares, Usage: "List of shares to apply Break Glass action (required)"},
		},
		RunFunc:    handleBreakGlassApplyCmd,
		WizardFunc: handleBreakGlassApplyWizard,
	})
}

// Handlers:

func handleBreakGlassGetStatusCmd(vcli vaultcli.CLI, args []string) int {
	data, err := breakGlassStatus(vcli)
	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func handleBreakGlassGenerateCmd(vcli vaultcli.CLI, args []string) int {
	newAdmins := viper.GetString(cst.NewAdmins)
	numberOfSharesString := viper.GetString(cst.MinNumberOfShares)

	if newAdmins == "" {
		err := errors.NewS("error: must specify " + cst.NewAdmins)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	if numberOfSharesString == "" {
		err := errors.NewS("error: must specify " + cst.MinNumberOfShares)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	numberOfShares, notApiErr := strconv.Atoi(numberOfSharesString)
	if notApiErr != nil {
		err := errors.NewS("error: minimum number of shares must be a valid integer")
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}
	if numberOfShares < 1 {
		err := errors.NewS("error: minimum number of shares must be greater than 1")
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	trimmedNewAdmins := strings.Trim(newAdmins, ",")

	data, err := breakGlassGenerate(vcli, utils.StringToSlice(trimmedNewAdmins), numberOfShares)
	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func handleBreakGlassApplyCmd(vcli vaultcli.CLI, args []string) int {
	shares := viper.GetString(cst.Shares)

	if shares == "" {
		err := errors.NewS("error: must specify " + cst.Shares)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, err := breakGlassApply(vcli, utils.StringToSlice(shares))
	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

// Wizards:

func handleBreakGlassGenerateWizard(vcli vaultcli.CLI) int {
	newAdmins := []string{}
	minNumberOfShares := 0

	for {
		qs := []*survey.Question{
			{
				Name: "newAdmin",
				Prompt: &survey.Input{
					Message: "New admin:",
					Help:    "Choose who the new administrators will be after the Break Glass event.",
				},
				Transform: vaultcli.SurveyTrimSpace,
			},
			{Name: "addMore", Prompt: &survey.Confirm{Message: "Add more?", Default: true}},
		}

		answers := struct {
			NewAdmin string
			AddMore  bool
		}{}
		survErr := survey.Ask(qs, &answers)
		if survErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(survErr))
			return utils.GetExecStatus(survErr)
		}
		if answers.NewAdmin != "" {
			newAdmins = append(newAdmins, answers.NewAdmin)
		}
		if !answers.AddMore {
			break
		}
	}

	if len(newAdmins) == 0 {
		err := errors.NewS("At least one new admin is required.")
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	minNumberOfSharesPrompt := &survey.Input{Message: "Minimum number of shares:"}
	minNumberOfSharesValidation := func(ans interface{}) error {
		answer := ans.(string)
		if len(answer) == 0 {
			return errors.NewS("Minimum number of shares is required.")
		}
		n, err := strconv.Atoi(answer)
		if err != nil {
			return errors.NewS("Minimum number of shares must be a valid integer.")
		}
		if n < 1 {
			return errors.NewS("Minimum number of shares must be greater than 1.")
		}
		if n > len(newAdmins) {
			return errors.NewS("Minimum number of shares cannot be greater than number of admins.")
		}
		return nil
	}
	survErr := survey.AskOne(minNumberOfSharesPrompt, &minNumberOfShares, survey.WithValidator(minNumberOfSharesValidation))
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}

	data, err := breakGlassGenerate(vcli, newAdmins, minNumberOfShares)
	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func handleBreakGlassApplyWizard(vcli vaultcli.CLI) int {
	data, err := breakGlassStatus(vcli)
	if err != nil {
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}
	if !strings.Contains(string(data), "Break Glass feature is set") {
		err = errors.NewS("Break Glass feature is not set.")
	}

	shares := []string{}
	for {
		qs := []*survey.Question{
			{
				Name:      "share",
				Prompt:    &survey.Input{Message: "Share:"},
				Transform: vaultcli.SurveyTrimSpace,
			},
			{Name: "addMore", Prompt: &survey.Confirm{Message: "Add more?", Default: true}},
		}
		answers := struct {
			Share   string
			AddMore bool
		}{}
		survErr := survey.Ask(qs, &answers)
		if survErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(survErr))
			return utils.GetExecStatus(survErr)
		}
		if answers.Share != "" {
			shares = append(shares, answers.Share)
		}
		if !answers.AddMore {
			break
		}
	}

	if len(shares) == 0 {
		err := errors.NewS("At least one share is required.")
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, err = breakGlassApply(vcli, shares)
	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

// API callers:

func breakGlassStatus(vcli vaultcli.CLI) ([]byte, *errors.ApiError) {
	uri := paths.CreateURI("breakglass", nil)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}

func breakGlassGenerate(vcli vaultcli.CLI, newAdmins []string, minNumber int) ([]byte, *errors.ApiError) {
	uri := paths.CreateURI("breakglass/generate", nil)
	body := map[string]interface{}{
		"newAdmins":         newAdmins,
		"minNumberOfShares": minNumber,
	}
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
}

func breakGlassApply(vcli vaultcli.CLI, shares []string) ([]byte, *errors.ApiError) {
	uri := paths.CreateURI("breakglass/apply", nil)
	body := map[string]interface{}{"shares": shares}
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
}
