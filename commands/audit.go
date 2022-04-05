package cmd

import (
	"net/http"
	"time"

	cst "thy/constants"
	"thy/errors"
	"thy/internal/predictor"
	"thy/paths"
	"thy/utils"
	"thy/vaultcli"

	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

func GetAuditSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounAudit},
		SynopsisText: "audit search",
		HelpText: `Search audit records

Usage:
   • audit --startdate 2020-01-21
   • audit --startdate 2020-01-21 --enddate 2020-01-22 --limit 10
   • audit --startdate 2020-01-21 --actions POST --path secrets
`,
		FlagsPredictor: []*predictor.Params{
			{Name: cst.StartDate, Shorthand: "s", Usage: "Start date from which to fetch audit data (required)"},
			{Name: cst.EndDate, Usage: "End date to which to fetch audit data (optional)"},
			{Name: cst.Limit, Shorthand: "l", Usage: "Maximum number of results per cursor (optional)"},
			{Name: cst.Cursor, Usage: cst.CursorHelpMessage},
			{Name: cst.Path, Usage: "Path (optional)"},
			{Name: cst.NounPrincipal, Usage: "Principal name (optional)"},
			{Name: cst.DataAction, Usage: "Action performed (POST, GET, PUT, PATCH or DELETE) (optional)"},
		},
		MinNumberArgs: 1,
		RunFunc: func(args []string) int {
			return handleAuditSearch(vaultcli.New(), args)
		},
	})
}

func handleAuditSearch(vcli vaultcli.CLI, args []string) int {
	s := viper.GetString(cst.StartDate)
	e := viper.GetString(cst.EndDate)
	if s == "" {
		err := errors.NewS("error: must specify " + cst.StartDate)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	const layout = "2006-01-02"

	startDate, parsingErr := time.Parse(layout, s)
	if parsingErr != nil {
		err := errors.NewS("error: must correctly specify " + cst.StartDate)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	var endDate time.Time
	if e == "" {
		endDate = time.Now() // end date is today
	} else if s == e {
		endDate = startDate
	} else {
		endDate, parsingErr = time.Parse(layout, e)
		if parsingErr != nil {
			err := errors.NewS("error: must correctly specify " + cst.EndDate)
			vcli.Out().WriteResponse(nil, err)
			return utils.GetExecStatus(err)
		}
	}

	if time.Now().Before(startDate) {
		err := errors.NewS("error: start date cannot be in the future")
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	if endDate.Before(startDate) {
		err := errors.NewS("error: start date cannot be after end date")
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	// Always add one day to the end date to include data for that day.
	endDate = endDate.AddDate(0, 0, 1)
	queryParams := map[string]string{
		"startDate":        startDate.Format(layout),
		"endDate":          endDate.Format(layout),
		cst.NounPrincipal:  viper.GetString(cst.NounPrincipal),
		cst.Path:           viper.GetString(cst.Path),
		cst.DataAction[:6]: viper.GetString(cst.DataAction),
		cst.Limit:          viper.GetString(cst.Limit),
		cst.Cursor:         viper.GetString(cst.Cursor),
	}
	uri := paths.CreateURI("audit", queryParams)
	data, err := vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)

	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}
