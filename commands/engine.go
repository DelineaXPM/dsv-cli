package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

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

type engineHandler struct {
	request   requests.Client
	outClient format.OutClient
}

func GetEngineCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounEngine},
		RunFunc: func(args []string) int {
			path := viper.GetString(cst.Path)
			if path == "" && len(args) > 0 {
				path = args[0]
			}
			if path == "" {
				return cli.RunResultHelp
			}
			return engineHandler{requests.NewHttpClient(), nil}.handleRead(args)
		},
		SynopsisText:  "engine (<action>)",
		HelpText:      "Work with engines",
		MinNumberArgs: 0,
	})
}

func GetEngineReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEngine, cst.Read},
		RunFunc:      engineHandler{requests.NewHttpClient(), nil}.handleRead,
		SynopsisText: "Get information on an existing engine",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myengine`, cst.NounEngine, cst.Read, cst.DataName),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s (required)", cst.NounEngine)}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetEngineListCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEngine, cst.List},
		RunFunc:      engineHandler{requests.NewHttpClient(), nil}.handleList,
		SynopsisText: "List the names of all existing engines and their appropriate pool names",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s`, cst.NounEngine, cst.List),
	})
}

func GetEngineDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEngine, cst.Delete},
		RunFunc:      engineHandler{requests.NewHttpClient(), nil}.handleDelete,
		SynopsisText: "Delete an existing engine",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myengine`, cst.NounEngine, cst.Delete, cst.DataName),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s (required)", cst.NounEngine)}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetEngineCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEngine, cst.Create},
		RunFunc:      engineHandler{requests.NewHttpClient(), nil}.handleCreate,
		SynopsisText: "Create a new engine",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myengine --pool-name mypool`, cst.NounEngine, cst.Create, cst.DataName),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataName):     cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s (required)", cst.NounEngine)}), false},
			preds.LongFlag(cst.DataPoolName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataPoolName, Usage: fmt.Sprintf("Name of the %s (required)", cst.NounPool)}), false},
		},
		MinNumberArgs: 0,
	})
}

func GetEnginePingCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEngine, cst.Ping},
		RunFunc:      engineHandler{requests.NewHttpClient(), nil}.handlePing,
		SynopsisText: "Ping a running engine",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s myengine`, cst.NounEngine, cst.Ping, cst.DataName),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s (required)", cst.NounEngine)}), false},
		},
		MinNumberArgs: 1,
	})
}

func (e engineHandler) handleRead(args []string) int {
	if e.outClient == nil {
		e.outClient = format.NewDefaultOutClient()
	}
	var err *errors.ApiError
	var data []byte
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 {
		name = args[0]
	}
	if name == "" {
		err = errors.NewS("error: must specify " + cst.DataName)
	} else {
		uri := paths.CreateResourceURI(cst.NounEngine, paths.ProcessResource(name), "", true, nil, true)
		data, err = e.request.DoRequest("GET", uri, nil)
	}

	e.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (e engineHandler) handleList(args []string) int {
	if e.outClient == nil {
		e.outClient = format.NewDefaultOutClient()
	}
	var err *errors.ApiError
	var data []byte
	uri := paths.CreateResourceURI(cst.NounEngine, "", "", false, nil, true)
	data, err = e.request.DoRequest("GET", uri, nil)

	e.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (e engineHandler) handleDelete(args []string) int {
	if e.outClient == nil {
		e.outClient = format.NewDefaultOutClient()
	}
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 {
		name = args[0]
	}
	if name == "" {
		err := errors.NewS("error: must specify " + cst.DataName)
		e.outClient.WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	query := map[string]string{"force": strconv.FormatBool(true)}
	uri := paths.CreateResourceURI(cst.NounEngine, paths.ProcessResource(name), "", true, query, true)

	data, err := e.request.DoRequest(http.MethodDelete, uri, nil)
	e.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (e engineHandler) handlePing(args []string) int {
	if e.outClient == nil {
		e.outClient = format.NewDefaultOutClient()
	}
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 {
		name = args[0]
	}
	if name == "" {
		err := errors.NewS("error: must specify " + cst.DataName)
		e.outClient.WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	uri := paths.CreateResourceURI(cst.NounEngine, paths.ProcessResource(name), "/ping", true, nil, true)
	data, err := e.request.DoRequest(http.MethodPost, uri, nil)
	e.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (e engineHandler) handleCreate(args []string) int {
	if OnlyGlobalArgs(args) {
		return e.handleCreateWizard(args)
	}
	if e.outClient == nil {
		e.outClient = format.NewDefaultOutClient()
	}
	engineName := viper.GetString(cst.DataName)
	poolName := viper.GetString(cst.DataPoolName)
	if engineName == "" || poolName == "" {
		err := errors.NewS("error: must specify engine name and pool name")
		e.outClient.WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}
	engine := engineCreate{
		Name:     engineName,
		PoolName: poolName,
	}

	data, err := e.submitEngine(engine)
	e.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (e engineHandler) handleCreateWizard(args []string) int {
	ui := &cli.BasicUi{
		Writer:      os.Stdout,
		Reader:      os.Stdin,
		ErrorWriter: os.Stderr,
	}
	if e.outClient == nil {
		e.outClient = format.NewDefaultOutClient()
	}

	var engine engineCreate

	if resp, err := getStringAndValidate(ui, "Engine name:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		engine.Name = resp
	}

	if resp, err := getStringAndValidate(ui, "Pool name:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		engine.PoolName = resp
	}

	data, err := e.submitEngine(engine)
	e.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (e engineHandler) submitEngine(engine engineCreate) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounEngine, "", "", true, nil, true)
	return e.request.DoRequest(http.MethodPost, uri, &engine)
}

type engineCreate struct {
	Name     string
	PoolName string
}
