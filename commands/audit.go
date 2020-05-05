package cmd

import (
	"fmt"
	cst "thy/constants"
	"thy/errors"
	"thy/format"
	preds "thy/predictors"
	"thy/requests"
	"thy/utils"
	"time"

	"github.com/posener/complete"
	"github.com/thycotic-rd/cli"
	"github.com/thycotic-rd/viper"
)

type audit struct {
	request   requests.Client
	outClient format.OutClient
}

func GetAuditSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounAudit},
		RunFunc:      audit{requests.NewHttpClient(), nil}.handleAuditSearch,
		SynopsisText: "audit search",
		HelpText: fmt.Sprintf(`Search audit records

Usage:
   • %[1]s --%[2]s 2020-01-01 --%[3]s 2020-01-04 --%[4]s 100
   • %[1]s --%[2]s 2020-01-01
   `, cst.NounAudit, cst.StartDate, cst.EndDate, cst.Limit),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.StartDate):     cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.StartDate, Usage: "Start date from which to fetch audit data (required)"}), false},
			preds.LongFlag(cst.EndDate):       cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.EndDate, Usage: "End date to which to fetch audit data (optional)"}), false},
			preds.LongFlag(cst.Limit):         cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Limit, Shorthand: "l", Usage: "Maximum number of results per cursor (optional)"}), false},
			preds.LongFlag(cst.Cursor):        cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Cursor, Usage: "Next cursor for additional results (optional)"}), false},
			preds.LongFlag(cst.Path):          cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Path, Usage: "Path (optional)"}), false},
			preds.LongFlag(cst.NounPrincipal): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.NounPrincipal, Usage: "Principal name (optional)"}), false},
			preds.LongFlag(cst.DataAction):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataAction, Usage: "Action performed (optional)"}), false},
		},
		MinNumberArgs: 1,
	})
}

func (a audit) handleAuditSearch(args []string) int {
	var err *errors.ApiError
	var data []byte
	if a.outClient == nil {
		a.outClient = format.NewDefaultOutClient()
	}
	s := viper.GetString(cst.StartDate)
	e := viper.GetString(cst.EndDate)
	if s == "" {
		err = errors.NewS("error: must specify " + cst.StartDate)
		a.outClient.WriteResponse(data, err)
		return utils.GetExecStatus(err)
	}

	const layout = "2006-01-02"

	startDate, parsingErr := time.Parse(layout, s)
	if parsingErr != nil {
		err = errors.NewS("error: must correctly specify " + cst.StartDate)
		a.outClient.WriteResponse(data, err)
		return utils.GetExecStatus(err)
	}

	var endDate time.Time
	if e == "" || s == e {
		endDate = startDate
	} else {
		endDate, parsingErr = time.Parse(layout, e)
		if parsingErr != nil {
			err = errors.NewS("error: must correctly specify " + cst.EndDate)
			a.outClient.WriteResponse(data, err)
			return utils.GetExecStatus(err)
		}

	}

	if time.Now().Before(startDate) {
		err = errors.NewS("error: start date cannot be in the future")
		a.outClient.WriteResponse(data, err)
		return utils.GetExecStatus(err)
	}

	if endDate.Before(startDate) {
		err = errors.NewS("error: start date cannot be after end date")
		a.outClient.WriteResponse(data, err)
		return utils.GetExecStatus(err)
	}

	// Always add one day to the end date to include data for that day.
	endDate = endDate.AddDate(0, 0, 1)
	queryParams := map[string]string{
		"startDate":       startDate.Format(layout),
		"endDate":         endDate.Format(layout),
		cst.NounPrincipal: viper.GetString(cst.NounPrincipal),
		cst.Path:          viper.GetString(cst.Path),
		cst.DataAction:    viper.GetString(cst.DataAction),
		cst.Limit:         viper.GetString(cst.Limit),
		cst.Cursor:        viper.GetString(cst.Cursor),
	}
	uri := utils.CreateURI("audit", queryParams)
	data, err = a.request.DoRequest("GET", uri, nil)

	a.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}
