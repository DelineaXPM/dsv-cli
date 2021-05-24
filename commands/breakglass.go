package cmd

import (
	"fmt"
	"strconv"
	"strings"
	cst "thy/constants"
	"thy/errors"
	"thy/format"
	"thy/paths"
	preds "thy/predictors"
	"thy/requests"
	"thy/utils"

	"github.com/posener/complete"
	"github.com/spf13/viper"

	"github.com/thycotic-rd/cli"
)

type breakglass struct {
	request   requests.Client
	outClient format.OutClient
}

func newBreakglass() breakglass {
	return breakglass{requests.NewHttpClient(), format.NewDefaultOutClient()}
}

func GetBreakglassCmd() (cli.Command, error) {
	flagsPredictor := cli.PredictorWrappers{}

	return NewCommand(CommandArgs{
		Path: []string{cst.NounBreakglass},
		RunFunc: func(args []string) int {
			return cli.RunResultHelp
		},
		SynopsisText:   "breakglass <action>",
		HelpText:       "Initiate restoration of admin users",
		FlagsPredictor: flagsPredictor,
		MinNumberArgs:  0,
	})
}

func GetBreakglassGetStatusCmd() (cli.Command, error) {
	flagsPredictor := cli.PredictorWrappers{}

	return NewCommand(CommandArgs{
		Path:         []string{cst.NounBreakglass, cst.Status},
		RunFunc:      newBreakglass().handleBreakglassGetStatusCmd,
		SynopsisText: "Check whether Break Glass feature is set up for the tenant",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s
   `, cst.NounBreakglass, cst.Status),
		FlagsPredictor: flagsPredictor,
		MinNumberArgs:  0,
	})
}

func GetBreakglassGenerateCmd() (cli.Command, error) {
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
		Path:         []string{cst.NounBreakglass, cst.Generate},
		RunFunc:      newBreakglass().handleBreakglassGenerateCmd,
		SynopsisText: "Generate and store admin secret and new admins' shares",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s 'newAdminUsername1,newAdminUsername2' --%[4]s 2
   `, cst.NounBreakglass, cst.Generate, cst.NewAdmins, cst.MinNumberOfShares),
		FlagsPredictor: flagsPredictor,
		MinNumberArgs:  2,
	})
}

func GetBreakglassApplyCmd() (cli.Command, error) {
	flagsPredictor := cli.PredictorWrappers{
		preds.LongFlag(cst.Shares): cli.PredictorWrapper{
			complete.PredictAnything,
			preds.NewFlagValue(preds.Params{Name: cst.Shares, Usage: "List of shares to apply Break Glass action (required)"}),
			false},
	}

	return NewCommand(CommandArgs{
		Path:         []string{cst.NounBreakglass, cst.Apply},
		RunFunc:      newBreakglass().handleBreakglassApplyCmd,
		SynopsisText: "Apply shares and break glass",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s '{share1},{share2},...,{shareN}'
   `, cst.NounBreakglass, cst.Apply, cst.Shares),
		FlagsPredictor: flagsPredictor,
		MinNumberArgs:  1,
	})
}

func (self breakglass) handleBreakglassGetStatusCmd(args []string) int {
	var err *errors.ApiError
	var data []byte

	uri := paths.CreateURI("breakglass", nil)
	data, err = self.request.DoRequest("GET", uri, nil)
	self.outClient.WriteResponse(data, err)

	return utils.GetExecStatus(err)
}

func (self breakglass) handleBreakglassGenerateCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	newAdmins := viper.GetString(cst.NewAdmins)
	numberOfSharesString := viper.GetString(cst.MinNumberOfShares)

	if newAdmins == "" {
		err = errors.NewS("error: must specify " + cst.NewAdmins)
		self.outClient.WriteResponse(data, err)
		return utils.GetExecStatus(err)
	}

	if numberOfSharesString == "" {
		err = errors.NewS("error: must specify " + cst.MinNumberOfShares)
		self.outClient.WriteResponse(data, err)
		return utils.GetExecStatus(err)
	}

	numberOfShares, notApiErr := strconv.Atoi(numberOfSharesString)
	if notApiErr != nil {
		return utils.GetExecStatus(notApiErr)
	}

	trimmedNewAdmins := strings.Trim(newAdmins, ",")

	gr := &breakglassGenerateRequest{
		NewAdmins:         utils.StringToSlice(trimmedNewAdmins),
		MinNumberOfShares: numberOfShares,
	}

	uri := paths.CreateURI("breakglass/generate", nil)
	data, err = self.request.DoRequest("POST", uri, gr)
	self.outClient.WriteResponse(data, err)

	return utils.GetExecStatus(err)
}

func (self breakglass) handleBreakglassApplyCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	shares := viper.GetString(cst.Shares)

	if shares == "" {
		err = errors.NewS("error: must specify " + cst.Shares)
		self.outClient.WriteResponse(data, err)
		return utils.GetExecStatus(err)
	}

	ar := &breakglassApplyRequest{Shares: utils.StringToSlice(shares)}

	uri := paths.CreateURI("breakglass/apply", nil)
	data, err = self.request.DoRequest("POST", uri, ar)
	self.outClient.WriteResponse(data, err)

	return utils.GetExecStatus(err)
}

type breakglassGenerateRequest struct {
	NewAdmins         []string `json:"newAdmins"`
	MinNumberOfShares int      `json:"minNumberOfShares"`
}

type breakglassApplyRequest struct {
	Shares []string `json:"shares"`
}
