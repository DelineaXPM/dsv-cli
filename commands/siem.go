package cmd

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	cst "thy/constants"
	"thy/errors"
	"thy/internal/predictor"
	"thy/paths"
	"thy/utils"
	"thy/vaultcli"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

func GetSiemCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSiem},
		SynopsisText: "siem (<action>)",
		HelpText:     "Work with SIEM endpoints",
		RunFunc: func(args []string) int {
			path := viper.GetString(cst.Path)
			if path == "" && len(args) > 0 {
				path = args[0]
			}
			if path == "" {
				return cli.RunResultHelp
			}
			return handleSiemRead(vaultcli.New(), args)
		},
	})
}

func GetSiemCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSiem, cst.Create},
		SynopsisText: "Create a new SIEM endpoint",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s`, cst.NounSiem, cst.Create),
		RunFunc: func(args []string) int {
			return handleSiemCreate(vaultcli.New(), args)
		},
	})
}

func GetSiemUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSiem, cst.Update},
		SynopsisText: "Update an existing SIEM endpoint",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[4]s
   • %[1]s %[2]s --%[3]s %[4]s`, cst.NounSiem, cst.Update, cst.Path, cst.ExampleSIEM),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Usage: "Path to existing SIEM"},
		},
		MinNumberArgs: 1,
		RunFunc: func(args []string) int {
			return handleSiemUpdate(vaultcli.New(), args)
		},
	})
}

func GetSiemReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSiem, cst.Read},
		SynopsisText: "Read an existing SIEM endpoint",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[4]s
   • %[1]s %[2]s --%[3]s %[4]s`, cst.NounSiem, cst.Read, cst.Path, cst.ExampleSIEM),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Usage: "Path to existing SIEM"},
		},
		MinNumberArgs: 1,
		RunFunc: func(args []string) int {
			return handleSiemRead(vaultcli.New(), args)
		},
	})
}

func GetSiemDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSiem, cst.Delete},
		SynopsisText: "Delete an existing SIEM endpoint",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[4]s
   • %[1]s %[2]s --%[3]s %[4]s`, cst.NounSiem, cst.Delete, cst.Path, cst.ExampleSIEM),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Usage: "Path to existing SIEM"},
		},
		MinNumberArgs: 1,
		RunFunc: func(args []string) int {
			return handleSiemDelete(vaultcli.New(), args)
		},
	})
}

func handleSiemCreate(vcli vaultcli.CLI, args []string) int {
	qs := []*survey.Question{
		{
			Name:   "SIEMType",
			Prompt: &survey.Input{Message: "Type of SIEM endpoint:", Default: "syslog"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if len(answer) == 0 {
					return errors.NewS("Value is required.")
				}
				return nil
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				return strings.TrimSpace(ans.(string))
			},
		},
		{
			Name:   "Name",
			Prompt: &survey.Input{Message: "Name of SIEM endpoint:"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if len(answer) == 0 {
					return errors.NewS("Value is required.")
				}
				return nil
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				return strings.TrimSpace(ans.(string))
			},
		},
		{
			Name:   "Host",
			Prompt: &survey.Input{Message: "Host:"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if len(answer) == 0 {
					return errors.NewS("Value is required.")
				}
				return nil
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				return strings.TrimSpace(ans.(string))
			},
		},
		{
			Name:   "Port",
			Prompt: &survey.Input{Message: "Port:"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if len(answer) == 0 {
					return errors.NewS("Value is required.")
				}
				_, err := strconv.Atoi(answer)
				if err != nil {
					return errors.NewS("Value must be a number.")
				}
				return nil
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				answer := strings.TrimSpace(ans.(string))
				_, val := strconv.Atoi(answer)
				return val
			},
		},
		{
			Name:   "Protocol",
			Prompt: &survey.Input{Message: "Protocol:"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if len(answer) == 0 {
					return errors.NewS("Value is required.")
				}
				return nil
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				return strings.TrimSpace(ans.(string))
			},
		},
		{
			Name:   "LoggingFormat",
			Prompt: &survey.Input{Message: "Logging Format:", Default: "rfc5424"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if len(answer) == 0 {
					return errors.NewS("Value is required.")
				}
				return nil
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				return strings.TrimSpace(ans.(string))
			},
		},
		{
			Name:   "AuthMethod",
			Prompt: &survey.Input{Message: "Authentication Method:", Default: "token"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if len(answer) == 0 {
					return errors.NewS("Value is required.")
				}
				return nil
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				return strings.TrimSpace(ans.(string))
			},
		},
		{
			Name:   "Auth",
			Prompt: &survey.Password{Message: "Authentication:"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if len(answer) == 0 {
					return errors.NewS("Value is required.")
				}
				return nil
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				return strings.TrimSpace(ans.(string))
			},
		},
		{
			Name:   "SendToEngine",
			Prompt: &survey.Confirm{Message: "Route Through DSV Engine:", Default: false},
		},
	}

	answers := siemCreateRequest{}
	survErr := survey.Ask(qs, &answers)
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}

	if answers.SendToEngine {
		poolPrompt := &survey.Input{Message: "Engine Pool:"}
		poolValidation := func(ans interface{}) error {
			answer := strings.TrimSpace(ans.(string))
			if len(answer) == 0 {
				return errors.NewS("Value is required.")
			}
			return nil
		}
		survErr := survey.AskOne(poolPrompt, &answers.Pool, survey.WithValidator(poolValidation))
		if survErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(survErr))
			return utils.GetExecStatus(survErr)
		}
	}

	data, apiError := siemCreate(vcli, &answers)
	vcli.Out().WriteResponse(data, apiError)
	return utils.GetExecStatus(apiError)
}

func handleSiemUpdate(vcli vaultcli.CLI, args []string) int {
	name := viper.GetString(cst.Path)
	if name == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name = args[0]
	}
	if name == "" {
		vcli.Out().FailF("error: must specify %s", cst.Path)
		return 1
	}

	_, apiErr := siemRead(vcli, name)
	if apiErr != nil {
		vcli.Out().WriteResponse(nil, apiErr)
		return utils.GetExecStatus(apiErr)
	}

	qs := []*survey.Question{
		{
			Name:   "SIEMType",
			Prompt: &survey.Input{Message: "Type of SIEM endpoint:", Default: "syslog"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if len(answer) == 0 {
					return errors.NewS("Value is required.")
				}
				return nil
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				return strings.TrimSpace(ans.(string))
			},
		},
		{
			Name:   "Host",
			Prompt: &survey.Input{Message: "Host:"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if len(answer) == 0 {
					return errors.NewS("Value is required.")
				}
				return nil
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				return strings.TrimSpace(ans.(string))
			},
		},
		{
			Name:   "Port",
			Prompt: &survey.Input{Message: "Port:"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if len(answer) == 0 {
					return errors.NewS("Value is required.")
				}
				_, err := strconv.Atoi(answer)
				if err != nil {
					return errors.NewS("Value must be a number.")
				}
				return nil
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				answer := strings.TrimSpace(ans.(string))
				_, val := strconv.Atoi(answer)
				return val
			},
		},
		{
			Name:   "Protocol",
			Prompt: &survey.Input{Message: "Protocol:"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if len(answer) == 0 {
					return errors.NewS("Value is required.")
				}
				return nil
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				return strings.TrimSpace(ans.(string))
			},
		},
		{
			Name:   "LoggingFormat",
			Prompt: &survey.Input{Message: "Logging Format:", Default: "rfc5424"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if len(answer) == 0 {
					return errors.NewS("Value is required.")
				}
				return nil
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				return strings.TrimSpace(ans.(string))
			},
		},
		{
			Name:   "AuthMethod",
			Prompt: &survey.Input{Message: "Authentication Method:", Default: "token"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if len(answer) == 0 {
					return errors.NewS("Value is required.")
				}
				return nil
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				return strings.TrimSpace(ans.(string))
			},
		},
		{
			Name:   "Auth",
			Prompt: &survey.Password{Message: "Authentication:"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if len(answer) == 0 {
					return errors.NewS("Value is required.")
				}
				return nil
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				return strings.TrimSpace(ans.(string))
			},
		},
		{
			Name:   "SendToEngine",
			Prompt: &survey.Confirm{Message: "Route Through DSV Engine:", Default: false},
		},
	}

	answers := siemUpdateRequest{}
	survErr := survey.Ask(qs, &answers)
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}

	if answers.SendToEngine {
		poolPrompt := &survey.Input{Message: "Engine Pool:"}
		poolValidation := func(ans interface{}) error {
			answer := strings.TrimSpace(ans.(string))
			if len(answer) == 0 {
				return errors.NewS("Value is required.")
			}
			return nil
		}
		survErr := survey.AskOne(poolPrompt, &answers.Pool, survey.WithValidator(poolValidation))
		if survErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(survErr))
			return utils.GetExecStatus(survErr)
		}
	}

	data, apiErr := siemUpdate(vcli, name, &answers)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleSiemRead(vcli vaultcli.CLI, args []string) int {
	name := viper.GetString(cst.Path)
	if name == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name = args[0]
	}
	if name == "" {
		vcli.Out().FailF("error: must specify %s", cst.Path)
		return 1
	}
	data, apiErr := siemRead(vcli, name)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleSiemDelete(vcli vaultcli.CLI, args []string) int {
	name := viper.GetString(cst.Path)
	if name == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name = args[0]
	}
	if name == "" {
		vcli.Out().FailF("error: must specify %s", cst.Path)
		return 1
	}
	data, apiErr := siemDelete(vcli, name)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

// API callers:

type siemCreateRequest struct {
	SIEMType      string `json:"siemType"`
	Name          string `json:"name"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Protocol      string `json:"protocol"`
	LoggingFormat string `json:"loggingFormat"`
	AuthMethod    string `json:"authMethod"`
	Auth          string `json:"auth"`
	SendToEngine  bool   `json:"sendToEngine"`
	Pool          string `json:"pool"`
}

func siemCreate(vcli vaultcli.CLI, body *siemCreateRequest) ([]byte, *errors.ApiError) {
	basePath := strings.Join([]string{cst.Config, cst.NounSiem}, "/")
	uri := paths.CreateURI(basePath, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
}

type siemUpdateRequest struct {
	SIEMType      string `json:"siemType"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Protocol      string `json:"protocol"`
	LoggingFormat string `json:"loggingFormat"`
	AuthMethod    string `json:"authMethod"`
	Auth          string `json:"auth"`
	SendToEngine  bool   `json:"sendToEngine"`
	Pool          string `json:"pool"`
}

func siemUpdate(vcli vaultcli.CLI, name string, body *siemUpdateRequest) ([]byte, *errors.ApiError) {
	basePath := strings.Join([]string{cst.Config, cst.NounSiem}, "/")
	uri := paths.CreateResourceURI(basePath, name, "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPut, uri, body)
}

func siemRead(vcli vaultcli.CLI, name string) ([]byte, *errors.ApiError) {
	basePath := strings.Join([]string{cst.Config, cst.NounSiem}, "/")
	uri := paths.CreateResourceURI(basePath, name, "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}

func siemDelete(vcli vaultcli.CLI, name string) ([]byte, *errors.ApiError) {
	basePath := strings.Join([]string{cst.Config, cst.NounSiem}, "/")
	uri := paths.CreateResourceURI(basePath, name, "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodDelete, uri, nil)
}
