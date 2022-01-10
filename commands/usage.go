package cmd

import (
	"fmt"
	"net/http"
	"time"

	"thy/errors"
	"thy/format"
	"thy/paths"
	preds "thy/predictors"
	"thy/requests"
	"thy/utils"

	"github.com/posener/complete"
	"github.com/spf13/viper"
	"github.com/thycotic-rd/cli"

	cst "thy/constants"
)

type usage struct {
	request   requests.Client
	outClient format.OutClient
}

func GetUsageCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUsage},
		RunFunc:      usage{requests.NewHttpClient(), nil}.handleGetUsageCmd,
		SynopsisText: "usage",
		HelpText:     fmt.Sprintf("Fetch the number of API calls used daily from %s", cst.ProductName),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.StartDate): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.StartDate, Shorthand: "s", Usage: fmt.Sprintf("Start date from which to fetch usage data (required)")}), false},
			preds.LongFlag(cst.EndDate):   cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.EndDate, Usage: fmt.Sprintf("End date to which to fetch usage data (optional)")}), false},
		},
		MinNumberArgs: 1,
	})
}

func (u usage) handleGetUsageCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	if u.outClient == nil {
		u.outClient = format.NewDefaultOutClient()
	}
	startDate := viper.GetString(cst.StartDate)
	if startDate == "" {
		err = errors.NewS("error: must specify " + cst.StartDate)
		u.outClient.WriteResponse(data, err)
		return utils.GetExecStatus(err)
	}

	endDate := viper.GetString(cst.EndDate)
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02") // end date is today
	}

	usageRequest := map[string]string{
		"startDate": startDate,
		"endDate":   endDate,
	}

	uri := paths.CreateURI(cst.NounUsage, usageRequest)
	data, err = u.request.DoRequest(http.MethodGet, uri, nil)

	u.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}
