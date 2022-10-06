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
		SynopsisText: "Show audit records",
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
			{Name: cst.Sort, Usage: "Change result sorting order (asc|desc) [default: desc] when search field is specified (optional)"},
		},
		MinNumberArgs: 1,
		RunFunc: func(vcli vaultcli.CLI, args []string) int {
			return handleAuditSearch(vcli, args)
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
		"startDate": startDate.Format(layout),
		"endDate":   endDate.Format(layout),
	}
	if nounPrincipal := viper.GetString(cst.NounPrincipal); nounPrincipal != "" {
		queryParams[cst.NounPrincipal] = nounPrincipal
	}
	if path := viper.GetString(cst.Path); path != "" {
		queryParams[cst.Path] = path
	}
	if dataAction := viper.GetString(cst.DataAction); dataAction != "" {
		queryParams[cst.DataAction[:6]] = dataAction
	}
	if limit := viper.GetString(cst.Limit); limit != "" {
		queryParams[cst.Limit] = limit
	}
	if cursor := viper.GetString(cst.Cursor); cursor != "" {
		queryParams[cst.Cursor] = cursor
	}
	if sort := viper.GetString(cst.Sort); sort != "" {
		queryParams[cst.Sort] = sort
	}
	uri := paths.CreateURI("audit", queryParams)
	data, err := vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)

	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}
