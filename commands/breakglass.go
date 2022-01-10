package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	cst "thy/constants"
	"thy/errors"
	"thy/format"
	"thy/internal/prompt"
	"thy/paths"
	preds "thy/predictors"
	"thy/requests"
	"thy/utils"

	"github.com/posener/complete"
	"github.com/spf13/viper"
	"github.com/thycotic-rd/cli"
)

type breakGlass struct {
	request   requests.Client
	outClient format.OutClient
}

func newBreakGlass() breakGlass {
	return breakGlass{requests.NewHttpClient(), format.NewDefaultOutClient()}
}

func GetBreakGlassCmd() (cli.Command, error) {
	flagsPredictor := cli.PredictorWrappers{}

	return NewCommand(CommandArgs{
		Path: []string{cst.NounBreakGlass},
		RunFunc: func(args []string) int {
			return cli.RunResultHelp
		},
		SynopsisText:   "breakglass <action>",
		HelpText:       "Initiate restoration of admin users",
		FlagsPredictor: flagsPredictor,
		MinNumberArgs:  0,
	})
}

func GetBreakGlassGetStatusCmd() (cli.Command, error) {
	flagsPredictor := cli.PredictorWrappers{}

	return NewCommand(CommandArgs{
		Path:         []string{cst.NounBreakGlass, cst.Status},
		RunFunc:      newBreakGlass().handleBreakGlassGetStatusCmd,
		SynopsisText: "Check whether Break Glass feature is set up for the tenant",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s
   `, cst.NounBreakGlass, cst.Status),
		FlagsPredictor: flagsPredictor,
		MinNumberArgs:  0,
	})
}

func GetBreakGlassGenerateCmd() (cli.Command, error) {
	flagsPredictor := cli.PredictorWrappers{
		preds.LongFlag(cst.NewAdmins): cli.PredictorWrapper{
			complete.PredictAnything,
			preds.NewFlagValue(preds.Params{Name: cst.NewAdmins, Usage: "New admins list (required)"}),
			false},
		preds.LongFlag(cst.MinNumberOfShares): cli.PredictorWrapper{
			complete.PredictAnything,
			preds.NewFlagValue(preds.Params{Name: cst.MinNumberOfShares, Usage: "Minimum number of shares to apply (required)"}),
			false},
	}

	return NewCommand(CommandArgs{
		Path:         []string{cst.NounBreakGlass, cst.Generate},
		RunFunc:      newBreakGlass().handleBreakGlassGenerateCmd,
		SynopsisText: "Generate and store admin secret and new admins' shares",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s 'newAdminUsername1,newAdminUsername2' --%[4]s 2
   `, cst.NounBreakGlass, cst.Generate, cst.NewAdmins, cst.MinNumberOfShares),
		FlagsPredictor: flagsPredictor,
		MinNumberArgs:  0,
	})
}

func GetBreakGlassApplyCmd() (cli.Command, error) {
	flagsPredictor := cli.PredictorWrappers{
		preds.LongFlag(cst.Shares): cli.PredictorWrapper{
			complete.PredictAnything,
			preds.NewFlagValue(preds.Params{Name: cst.Shares, Usage: "List of shares to apply Break Glass action (required)"}),
			false},
	}

	return NewCommand(CommandArgs{
		Path:         []string{cst.NounBreakGlass, cst.Apply},
		RunFunc:      newBreakGlass().handleBreakGlassApplyCmd,
		SynopsisText: "Apply shares and break glass",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s '{share1},{share2},...,{shareN}'
   `, cst.NounBreakGlass, cst.Apply, cst.Shares),
		FlagsPredictor: flagsPredictor,
		MinNumberArgs:  0,
	})
}

func (b breakGlass) handleBreakGlassGetStatusCmd(args []string) int {
	var err *errors.ApiError
	var data []byte

	uri := paths.CreateURI("breakglass", nil)
	data, err = b.request.DoRequest(http.MethodGet, uri, nil)
	b.outClient.WriteResponse(data, err)

	return utils.GetExecStatus(err)
}

func (b breakGlass) handleBreakGlassGenerateCmd(args []string) int {
	if OnlyGlobalArgs(args) {
		return b.handleBreakGlassGenerateWizard(args)
	}

	var err *errors.ApiError
	var data []byte
	newAdmins := viper.GetString(cst.NewAdmins)
	numberOfSharesString := viper.GetString(cst.MinNumberOfShares)

	if newAdmins == "" {
		err = errors.NewS("error: must specify " + cst.NewAdmins)
		b.outClient.WriteResponse(data, err)
		return utils.GetExecStatus(err)
	}

	if numberOfSharesString == "" {
		err = errors.NewS("error: must specify " + cst.MinNumberOfShares)
		b.outClient.WriteResponse(data, err)
		return utils.GetExecStatus(err)
	}

	numberOfShares, notApiErr := strconv.Atoi(numberOfSharesString)
	if notApiErr != nil {
		return utils.GetExecStatus(notApiErr)
	}

	trimmedNewAdmins := strings.Trim(newAdmins, ",")

	gr := &breakGlassGenerateRequest{
		NewAdmins:         utils.StringToSlice(trimmedNewAdmins),
		MinNumberOfShares: numberOfShares,
	}

	uri := paths.CreateURI("breakglass/generate", nil)
	data, err = b.request.DoRequest(http.MethodPost, uri, gr)
	b.outClient.WriteResponse(data, err)

	return utils.GetExecStatus(err)
}

func (b breakGlass) handleBreakGlassGenerateWizard(args []string) int {
	ui := &cli.BasicUi{
		Writer:      os.Stdout,
		Reader:      os.Stdin,
		ErrorWriter: os.Stderr,
	}
	if b.outClient == nil {
		b.outClient = format.NewDefaultOutClient()
	}

	var numberOfShares int
	var newAdmins string

	if resp, err := prompt.Ask(ui, "Minimum number of shares:"); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		numberOfShares, err = strconv.Atoi(resp)
		if err != nil {
			ui.Error("Invalid input. Please enter a valid integer.")
			return 1
		}
	}

	if resp, err := prompt.Ask(ui, "New admins (comma-separated):"); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		newAdmins = strings.Trim(resp, ",")
	}

	gr := &breakGlassGenerateRequest{
		NewAdmins:         utils.StringToSlice(newAdmins),
		MinNumberOfShares: numberOfShares,
	}

	uri := paths.CreateURI("breakglass/generate", nil)
	data, err := b.request.DoRequest(http.MethodPost, uri, gr)
	b.outClient.WriteResponse(data, err)

	return utils.GetExecStatus(err)
}

func (b breakGlass) handleBreakGlassApplyCmd(args []string) int {
	if OnlyGlobalArgs(args) {
		return b.handleBreakGlassApplyWizard(args)
	}

	var err *errors.ApiError
	var data []byte
	shares := viper.GetString(cst.Shares)

	if shares == "" {
		err = errors.NewS("error: must specify " + cst.Shares)
		b.outClient.WriteResponse(data, err)
		return utils.GetExecStatus(err)
	}

	ar := &breakGlassApplyRequest{Shares: utils.StringToSlice(shares)}

	uri := paths.CreateURI("breakglass/apply", nil)
	data, err = b.request.DoRequest(http.MethodPost, uri, ar)
	b.outClient.WriteResponse(data, err)

	return utils.GetExecStatus(err)
}

func (b breakGlass) handleBreakGlassApplyWizard(args []string) int {
	ui := &cli.BasicUi{
		Writer:      os.Stdout,
		Reader:      os.Stdin,
		ErrorWriter: os.Stderr,
	}
	if b.outClient == nil {
		b.outClient = format.NewDefaultOutClient()
	}

	var shares string

	if resp, err := prompt.Ask(ui, "Shares (comma-separated):"); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		shares = resp
	}

	ar := &breakGlassApplyRequest{Shares: utils.StringToSlice(shares)}

	uri := paths.CreateURI("breakglass/apply", nil)
	data, err := b.request.DoRequest(http.MethodPost, uri, ar)
	b.outClient.WriteResponse(data, err)

	return utils.GetExecStatus(err)
}

type breakGlassGenerateRequest struct {
	NewAdmins         []string `json:"newAdmins"`
	MinNumberOfShares int      `json:"minNumberOfShares"`
}

type breakGlassApplyRequest struct {
	Shares []string `json:"shares"`
}
