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

type poolHandler struct {
	request   requests.Client
	outClient format.OutClient
}

func GetPoolCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounPool},
		RunFunc: func(args []string) int {
			path := viper.GetString(cst.Path)
			if path == "" && len(args) > 0 {
				path = args[0]
			}
			if path == "" {
				return cli.RunResultHelp
			}
			return poolHandler{requests.NewHttpClient(), nil}.handleRead(args)
		},
		SynopsisText:  "pool (<action>)",
		HelpText:      "Work with engine pools",
		MinNumberArgs: 0,
	})
}

func GetPoolCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPool, cst.Create},
		RunFunc:      poolHandler{requests.NewHttpClient(), nil}.handleCreate,
		SynopsisText: "Create a new empty pool of engines",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s mypool`, cst.NounPool, cst.Create, cst.DataName),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s (required)", cst.NounPool)}), false},
		},
		MinNumberArgs: 0,
	})
}

func GetPoolReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPool, cst.Read},
		RunFunc:      poolHandler{requests.NewHttpClient(), nil}.handleRead,
		SynopsisText: "Get information on an existing pool of engines",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s mypool`, cst.NounPool, cst.Read, cst.DataName),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s (required)", cst.NounPool)}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetPoolListCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPool, cst.List},
		RunFunc:      poolHandler{requests.NewHttpClient(), nil}.handleList,
		SynopsisText: "List the names of all existing pools",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s`, cst.NounPool, cst.List),
	})
}

func GetPoolDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPool, cst.Delete},
		RunFunc:      poolHandler{requests.NewHttpClient(), nil}.handleDelete,
		SynopsisText: "Delete an existing pool of engines",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s mypool`, cst.NounPool, cst.Delete, cst.DataName),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Shorthand: "n", Name: cst.DataName, Usage: fmt.Sprintf("Name of the %s (required)", cst.NounPool)}), false},
		},
		MinNumberArgs: 1,
	})
}

func (p poolHandler) handleRead(args []string) int {
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
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
		uri := paths.CreateResourceURI(cst.NounPool, paths.ProcessResource(name), "", true, nil, true)
		data, err = p.request.DoRequest("GET", uri, nil)
	}

	p.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (p poolHandler) handleCreate(args []string) int {
	if OnlyGlobalArgs(args) {
		return p.handleCreateWizard(args)
	}

	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}
	name := viper.GetString(cst.DataName)
	if name == "" {
		err := errors.NewS("error: must specify " + cst.DataName)
		p.outClient.WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}
	pool := Pool{
		Name: name,
	}

	data, err := p.submitPool(pool)
	p.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (p poolHandler) handleList(args []string) int {
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}
	var err *errors.ApiError
	var data []byte
	uri := paths.CreateResourceURI(cst.NounPool, "", "", false, nil, true)
	data, err = p.request.DoRequest("GET", uri, nil)

	p.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (p poolHandler) handleDelete(args []string) int {
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 {
		name = args[0]
	}
	if name == "" {
		err := errors.NewS("error: must specify " + cst.DataName)
		p.outClient.WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	query := map[string]string{"force": strconv.FormatBool(true)}
	uri := paths.CreateResourceURI(cst.NounPool, paths.ProcessResource(name), "", true, query, true)

	data, err := p.request.DoRequest(http.MethodDelete, uri, nil)
	p.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (p poolHandler) handleCreateWizard(args []string) int {
	ui := &cli.BasicUi{
		Writer:      os.Stdout,
		Reader:      os.Stdin,
		ErrorWriter: os.Stderr,
	}
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}

	var pool Pool

	if resp, err := getStringAndValidate(ui, "Pool name:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		pool.Name = resp
	}

	data, err := p.submitPool(pool)
	p.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (p poolHandler) submitPool(pool Pool) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounPool, "", "", true, nil, true)
	return p.request.DoRequest(http.MethodPost, uri, &pool)
}

type Pool struct {
	Name string
}
