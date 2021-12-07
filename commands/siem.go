package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	cst "thy/constants"
	apperrors "thy/errors"
	"thy/format"
	"thy/paths"
	preds "thy/predictors"
	"thy/requests"
	"thy/utils"

	"github.com/posener/complete"
	"github.com/spf13/viper"
	"github.com/thycotic-rd/cli"
)

type siem struct {
	request   requests.Client
	outClient format.OutClient
}

func GetSiemCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounSiem},
		RunFunc: func(args []string) int {
			path := viper.GetString(cst.Path)
			if path == "" && len(args) > 0 {
				path = args[0]
			}
			if path == "" {
				return cli.RunResultHelp
			}
			return siem{requests.NewHttpClient(), nil}.handleRead(args)
		},
		SynopsisText:  "siem (<action>)",
		HelpText:      "Work with SIEM endpoints",
		MinNumberArgs: 0,
	})
}

func GetSiemCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSiem, cst.Create},
		RunFunc:      siem{requests.NewHttpClient(), nil}.handleCreate,
		SynopsisText: "Create a new SIEM endpoint",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s`, cst.NounSiem, cst.Create),
	})
}

func GetSiemUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSiem, cst.Update},
		RunFunc:      siem{requests.NewHttpClient(), nil}.handleUpdate,
		SynopsisText: "Update an existing SIEM endpoint",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[4]s
   • %[1]s %[2]s --%[3]s %[4]s`, cst.NounSiem, cst.Update, cst.Path, cst.ExampleSIEM),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Path, Usage: "Path to existing SIEM"}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetSiemReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSiem, cst.Read},
		RunFunc:      siem{requests.NewHttpClient(), nil}.handleRead,
		SynopsisText: "Read an existing SIEM endpoint",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[4]s
   • %[1]s %[2]s --%[3]s %[4]s`, cst.NounSiem, cst.Read, cst.Path, cst.ExampleSIEM),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Path, Usage: "Path to existing SIEM"}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetSiemDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSiem, cst.Delete},
		RunFunc:      siem{requests.NewHttpClient(), nil}.handleDelete,
		SynopsisText: "Delete an existing SIEM endpoint",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[4]s
   • %[1]s %[2]s --%[3]s %[4]s`, cst.NounSiem, cst.Delete, cst.Path, cst.ExampleSIEM),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Path, Usage: "Path to existing SIEM"}), false},
		},
		MinNumberArgs: 1,
	})
}

func (s siem) handleCreate([]string) int {
	var apiError *apperrors.ApiError
	var data []byte
	if s.outClient == nil {
		s.outClient = format.NewDefaultOutClient()
	}
	ui := &PasswordUi{
		cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
	}

	params := make(map[string]interface{})

	if resp, err := getStringAndValidateDefault(
		ui, "Type of SIEM endpoint (default:syslog):", "syslog", false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params["siemType"] = resp
	}

	if resp, err := getStringAndValidate(ui, "Name of SIEM endpoint:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params["name"] = resp
	}

	if resp, err := getStringAndValidate(ui, "Host:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params["host"] = resp
	}

	if resp, err := getStringAndValidate(ui, "Port:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		if port, err := strconv.Atoi(resp); err != nil {
			ui.Error("Error: port must be a number.")
			return 1
		} else {
			params["port"] = port
		}
	}

	if resp, err := getStringAndValidate(ui, "Protocol:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params["protocol"] = resp
	}

	if resp, err := getStringAndValidateDefault(ui, "Logging Format (default:rfc5424):", "rfc5424", false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params["loggingFormat"] = resp
	}

	if resp, err := getStringAndValidateDefault(ui, "Authentication Method (default:token):", "token", false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params["authMethod"] = resp
	}
	if resp, err := getStringAndValidate(ui, "Authentication:", false, nil, true, true); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params["auth"] = resp
	}

	if resp, err := getStringAndValidateDefault(ui, "Route Through DSV Engine [y/N]:", "N", false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		resp = strings.ToLower(resp)
		if !utils.EqAny(resp, []string{"y", "yes", "n", "no", ""}) {
			ui.Error("Invalid response, must choose (y)es or (n)o")
			return 1
		}
		if isYes(resp, false) {
			params["sendToEngine"] = true

			if resp, err := getStringAndValidate(ui, "Engine Pool:", false, nil, false, false); err != nil {
				ui.Error(err.Error())
				return 1
			} else {
				params["pool"] = resp
			}
		} else {
			params["sendToEngine"] = false
			params["pool"] = ""
		}
	}

	basePath := strings.Join([]string{cst.Config, cst.NounSiem}, "/")
	uri := paths.CreateURI(basePath, nil)
	data, apiError = s.request.DoRequest(http.MethodPost, uri, params)
	s.outClient.WriteResponse(data, apiError)
	return utils.GetExecStatus(apiError)
}

func (s siem) handleUpdate(args []string) int {
	var apiError *apperrors.ApiError
	var data []byte
	if s.outClient == nil {
		s.outClient = format.NewDefaultOutClient()
	}
	name := viper.GetString(cst.Path)
	if name == "" && len(args) > 0 {
		name = args[0]
	}
	if name == "" {
		s.outClient.FailF("error: must specify %s", cst.Path)
		return 1
	}

	// Check if an endpoint with a given name exists.
	if code := s.handleRead(args); code != 0 {
		return 1
	}

	ui := &PasswordUi{
		cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
	}

	params := make(map[string]interface{})

	if resp, err := getStringAndValidateDefault(
		ui, "Type of SIEM endpoint (default:syslog):", "syslog", false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params["siemType"] = resp
	}

	if resp, err := getStringAndValidate(ui, "Host:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params["host"] = resp
	}

	if resp, err := getStringAndValidate(ui, "Port:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		if port, err := strconv.Atoi(resp); err != nil {
			ui.Error("Error: port must be a number.")
			return 1
		} else {
			params["port"] = port
		}
	}

	if resp, err := getStringAndValidate(ui, "Protocol:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params["protocol"] = resp
	}

	if resp, err := getStringAndValidateDefault(ui, "Logging Format (default:rfc5424):", "rfc5424", false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params["loggingFormat"] = resp
	}

	if resp, err := getStringAndValidateDefault(ui, "Authentication Method (default:token):", "token", false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params["authMethod"] = resp
	}
	if resp, err := getStringAndValidate(ui, "Authentication:", false, nil, true, true); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params["auth"] = resp
	}

	if resp, err := getStringAndValidateDefault(ui, "Route Through DSV Engine [y/N]:", "N", false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		resp = strings.ToLower(resp)
		if !utils.EqAny(resp, []string{"y", "yes", "n", "no", ""}) {
			ui.Error("Invalid response, must choose (y)es or (n)o")
			return 1
		}

		if isYes(resp, false) {
			params["sendToEngine"] = true

			if resp, err := getStringAndValidate(ui, "Engine Pool:", false, nil, false, false); err != nil {
				ui.Error(err.Error())
				return 1
			} else {
				params["pool"] = resp
			}
		} else {
			params["sendToEngine"] = false
			params["pool"] = ""
		}
	}

	basePath := strings.Join([]string{cst.Config, cst.NounSiem}, "/")
	uri := paths.CreateResourceURI(basePath, name, "", true, nil, false)
	data, apiError = s.request.DoRequest(http.MethodPut, uri, params)
	s.outClient.WriteResponse(data, apiError)
	return utils.GetExecStatus(apiError)
}

func (s siem) handleRead(args []string) int {
	var apiError *apperrors.ApiError
	var data []byte
	if s.outClient == nil {
		s.outClient = format.NewDefaultOutClient()
	}
	name := viper.GetString(cst.Path)
	if name == "" && len(args) > 0 {
		name = args[0]
	}
	if name == "" {
		s.outClient.FailF("error: must specify %s", cst.Path)
		return 1
	}
	basePath := strings.Join([]string{cst.Config, cst.NounSiem}, "/")
	uri := paths.CreateResourceURI(basePath, name, "", true, nil, false)
	data, apiError = s.request.DoRequest(http.MethodGet, uri, nil)
	s.outClient.WriteResponse(data, apiError)
	return utils.GetExecStatus(apiError)
}

func (s siem) handleDelete(args []string) int {
	var apiError *apperrors.ApiError
	var data []byte
	if s.outClient == nil {
		s.outClient = format.NewDefaultOutClient()
	}
	name := viper.GetString(cst.Path)
	if name == "" && len(args) > 0 {
		name = args[0]
	}
	if name == "" {
		s.outClient.FailF("error: must specify %s", cst.Path)
		return 1
	}
	basePath := strings.Join([]string{cst.Config, cst.NounSiem}, "/")
	uri := paths.CreateResourceURI(basePath, name, "", true, nil, false)
	data, apiError = s.request.DoRequest(http.MethodDelete, uri, nil)
	s.outClient.WriteResponse(data, apiError)
	return utils.GetExecStatus(apiError)
}
