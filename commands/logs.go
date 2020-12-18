package cmd

import (
	"fmt"
	"time"

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

type logs struct {
	request   requests.Client
	outClient format.OutClient
}

func GetLogsSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounLogs},
		RunFunc:      logs{requests.NewHttpClient(), nil}.handleLogsSearch,
		SynopsisText: "system logs search",
		HelpText: fmt.Sprintf(`Search system logs

Usage:
	• %[1]s --%[2]s 2020-01-01 --%[3]s 2020-01-04 --%[4]s 100
	• %[1]s --%[2]s 2020-01-01
	`, cst.NounLogs, cst.StartDate, cst.EndDate, cst.Limit),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.StartDate):     cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.StartDate, Shorthand: "s", Usage: "Start date from which to fetch system log data (required)"}), false},
			preds.LongFlag(cst.EndDate):       cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.EndDate, Usage: "End date to which to fetch system log data (optional)"}), false},
			preds.LongFlag(cst.Limit):         cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Limit, Shorthand: "l", Usage: "Maximum number of results per cursor (optional)"}), false},
			preds.LongFlag(cst.Cursor):        cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Cursor, Usage: "Next cursor for additional results (optional)"}), false},
			preds.LongFlag(cst.Path):          cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Path, Usage: "Path (optional)"}), false},
			preds.LongFlag(cst.NounPrincipal): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.NounPrincipal, Usage: "Principal name (optional)"}), false},
			preds.LongFlag(cst.DataAction):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataAction, Usage: "Action performed (optional)"}), false},
		},
		MinNumberArgs: 1,
	})
}

func (l logs) handleLogsSearch(args []string) int {
	var err *errors.ApiError
	var data []byte
	if l.outClient == nil {
		l.outClient = format.NewDefaultOutClient()
	}
	startDate := viper.GetString(cst.StartDate)
	endDate := viper.GetString(cst.EndDate)
	if startDate == "" {
		err = errors.NewS("error: must specify " + cst.StartDate)
		l.outClient.WriteResponse(data, err)
		return utils.GetExecStatus(err)
	} else if endDate == "" {
		endDate = time.Now().Format("2006-01-02") // end date is today
	}

	queryParams := map[string]string{
		"startDate":       startDate,
		"endDate":         endDate,
		cst.NounPrincipal: viper.GetString(cst.NounPrincipal),
		cst.Path:          viper.GetString(cst.Path),
		cst.DataAction:    viper.GetString(cst.DataAction),
		cst.Limit:         viper.GetString(cst.Limit),
		cst.Cursor:        viper.GetString(cst.Cursor),
	}
	uri := paths.CreateURI("system/log", queryParams)
	data, err = l.request.DoRequest("GET", uri, nil)

	l.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}
