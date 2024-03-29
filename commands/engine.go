package cmd

import (
	"fmt"
	"net/http"
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

func GetEngineCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEngine},
		SynopsisText: "Manage engines",
		HelpText:     "Work with engines",
		RunFunc: func(vcli vaultcli.CLI, args []string) int {
			path := viper.GetString(cst.Path)
			if path == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
				path = args[0]
			}
			if path == "" {
				return cli.RunResultHelp
			}
			return handleEngineReadCmd(vcli, args)
		},
	})
}

func GetEngineReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEngine, cst.Read},
		SynopsisText: "Get information on an existing engine",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myengine`, cst.NounEngine, cst.Read, cst.DataName),
		FlagsPredictor: []*predictor.Params{
			{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s (required)", cst.NounEngine)},
		},
		MinNumberArgs: 1,
		RunFunc:       handleEngineReadCmd,
	})
}

func GetEngineListCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEngine, cst.List},
		SynopsisText: "List the names of all existing engines and their appropriate pool names",
		HelpText: `
Usage:
   • engine list
   • engine list --sort asc --sorted-by name --pool-name my_pool
   • engine list --sort desc --sorted-by created
   • engine list --query my_engine`,
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Query, Shorthand: "q", Usage: "Partial search by engine name (optional)"},
			{Name: cst.DataPoolName, Usage: "Pool name (optional)"},
			{Name: cst.Sort, Usage: cst.SortHelpMessage, Default: "desc"},
			{Name: cst.SortedBy, Usage: "Sort by name or created field (optional)", Default: "created"},
		},
		RunFunc: handleEngineListCmd,
	})
}

func GetEngineDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEngine, cst.Delete},
		SynopsisText: "Delete an existing engine",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myengine`, cst.NounEngine, cst.Delete, cst.DataName),
		FlagsPredictor: []*predictor.Params{
			{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s (required)", cst.NounEngine)},
		},
		MinNumberArgs: 1,
		RunFunc:       handleEngineDeleteCmd,
	})
}

func GetEngineCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEngine, cst.Create},
		SynopsisText: "Create a new engine",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myengine --pool-name mypool`, cst.NounEngine, cst.Create, cst.DataName),
		FlagsPredictor: []*predictor.Params{
			{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s (required)", cst.NounEngine)},
			{Name: cst.DataPoolName, Usage: fmt.Sprintf("Name of the %s (required)", cst.NounPool)},
		},
		RunFunc:    handleEngineCreateCmd,
		WizardFunc: handleEngineCreateWizard,
	})
}

func GetEnginePingCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEngine, cst.Ping},
		SynopsisText: "Ping a running engine",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myengine`, cst.NounEngine, cst.Ping, cst.DataName),
		FlagsPredictor: []*predictor.Params{
			{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s (required)", cst.NounEngine)},
		},
		MinNumberArgs: 1,
		RunFunc:       handleEnginePingCmd,
	})
}

func handleEngineReadCmd(vcli vaultcli.CLI, args []string) int {
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name = args[0]
	}
	if name == "" {
		err := errors.NewS("error: must specify " + cst.DataName)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, apiErr := engineRead(vcli, paths.ProcessResource(name))
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleEngineListCmd(vcli vaultcli.CLI, args []string) int {
	searchTerm := viper.GetString(cst.Query)
	poolName := viper.GetString(cst.DataPoolName)
	sort := viper.GetString(cst.Sort)
	sortedBy := viper.GetString(cst.SortedBy)

	data, apiErr := engineList(vcli, &engineListParams{
		searchTerm: searchTerm,
		poolName:   poolName,
		sort:       sort,
		sortedBy:   sortedBy,
	})
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleEngineDeleteCmd(vcli vaultcli.CLI, args []string) int {
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name = args[0]
	}
	if name == "" {
		err := errors.NewS("error: must specify " + cst.DataName)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, apiErr := engineDelete(vcli, paths.ProcessResource(name))
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleEnginePingCmd(vcli vaultcli.CLI, args []string) int {
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name = args[0]
	}
	if name == "" {
		err := errors.NewS("error: must specify " + cst.DataName)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, apiErr := enginePing(vcli, paths.ProcessResource(name))
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleEngineCreateCmd(vcli vaultcli.CLI, args []string) int {
	engineName := viper.GetString(cst.DataName)
	poolName := viper.GetString(cst.DataPoolName)
	if engineName == "" || poolName == "" {
		err := errors.NewS("error: must specify engine name and pool name")
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}
	if err := vaultcli.ValidateName(engineName); err != nil {
		vcli.Out().FailF("error: engine name %q is invalid: %v", engineName, err)
		return utils.GetExecStatus(err)
	}

	data, apiErr := engineCreate(vcli, engineName, poolName)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

// Wizards:

func handleEngineCreateWizard(vcli vaultcli.CLI) int {
	qs := []*survey.Question{
		{
			Name:   "EngineName",
			Prompt: &survey.Input{Message: "Engine name:"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if err := vaultcli.ValidateName(answer); err != nil {
					return err
				}
				_, apiError := engineRead(vcli, answer)
				if apiError == nil {
					return errors.NewS("An engine with this name already exists.")
				}
				return nil
			},
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:   "PoolName",
			Prompt: &survey.Input{Message: "Pool name:"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				_, apiError := poolRead(vcli, answer)
				if apiError != nil &&
					apiError.HttpResponse() != nil &&
					apiError.HttpResponse().StatusCode == http.StatusNotFound {
					return errors.NewS("A pool with this name does not exist.")
				}
				return nil
			},
			Transform: vaultcli.SurveyTrimSpace,
		},
	}

	answers := struct {
		EngineName string
		PoolName   string
	}{}
	survErr := survey.Ask(qs, &answers)
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}

	data, apiErr := engineCreate(vcli, answers.EngineName, answers.PoolName)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

// API callers:

func engineCreate(vcli vaultcli.CLI, engineName string, poolName string) ([]byte, *errors.ApiError) {
	body := map[string]string{"name": engineName, "poolName": poolName}
	uri := paths.CreateResourceURI(cst.NounEngines, "", "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, &body)
}

func engineRead(vcli vaultcli.CLI, engineName string) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounEngines, engineName, "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}

func enginePing(vcli vaultcli.CLI, engineName string) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounEngines, engineName, "/ping", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, nil)
}

func engineDelete(vcli vaultcli.CLI, engineName string) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounEngines, engineName, "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodDelete, uri, nil)
}

type engineListParams struct {
	searchTerm string
	poolName   string
	sort       string
	sortedBy   string
}

func engineList(vcli vaultcli.CLI, params *engineListParams) ([]byte, *errors.ApiError) {
	queryParams := map[string]string{}
	if params.searchTerm != "" {
		queryParams[cst.SearchTerm] = params.searchTerm
	}
	if params.poolName != "" {
		queryParams["poolName"] = params.poolName
	}
	if params.sort != "" {
		queryParams[cst.Sort] = params.sort
	}
	if params.sortedBy != "" {
		queryParams["sortedBy"] = params.sortedBy
	}
	uri := paths.CreateResourceURI(cst.NounEngines, "", "", false, queryParams)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}
