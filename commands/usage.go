package cmd

import (
	"fmt"
	"net/http"
	"time"

	"thy/errors"
	"thy/internal/predictor"
	"thy/paths"
	"thy/utils"
	"thy/vaultcli"

	"github.com/mitchellh/cli"
	"github.com/spf13/viper"

	cst "thy/constants"
)

func GetUsageCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUsage},
		SynopsisText: "Fetch API usage info",
		HelpText:     fmt.Sprintf("Fetch the number of API calls used daily from %s", cst.ProductName),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.StartDate, Shorthand: "s", Usage: "Start date from which to fetch usage data (required)"},
			{Name: cst.EndDate, Usage: "End date to which to fetch usage data (optional)"},
		},
		MinNumberArgs: 1,
		RunFunc:       handleGetUsageCmd,
	})
}

func handleGetUsageCmd(vcli vaultcli.CLI, args []string) int {
	startDate := viper.GetString(cst.StartDate)
	if startDate == "" {
		err := errors.NewS("error: must specify " + cst.StartDate)
		vcli.Out().WriteResponse(nil, err)
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
	data, err := vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)

	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}
