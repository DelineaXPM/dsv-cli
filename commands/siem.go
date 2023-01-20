package cmd

import (
	"fmt"
	"net/http"
	"strconv"
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

func GetSiemCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSiem},
		SynopsisText: "Manage SIEM endpoints",
		HelpText:     "Work with SIEM endpoints",
		RunFunc: func(vcli vaultcli.CLI, args []string) int {
			name := viper.GetString(cst.DataName)
			if name == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
				name = args[0]
			}
			if name == "" {
				name = viper.GetString(cst.Path)
			}
			if name == "" {
				return cli.RunResultHelp
			}
			return handleSiemRead(vcli, args)
		},
	})
}

func GetSiemCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSiem, cst.Create},
		SynopsisText: "Create a new SIEM endpoint",
		HelpText: `
Usage:
   • siem create
`,
		RunFunc: handleSiemCreate,
	})
}

func GetSiemUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSiem, cst.Update},
		SynopsisText: "Update an existing SIEM endpoint",
		HelpText: fmt.Sprintf(`
Usage:
   • siem update %[1]s
   • siem update --name %[1]s
`, cst.ExampleSIEM),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataName, Usage: "Name of existing SIEM"},
			{Name: cst.Path, Usage: "Path to existing SIEM", Hidden: true},
		},
		RunFunc: handleSiemUpdate,
	})
}

func GetSiemReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSiem, cst.Read},
		SynopsisText: "Read an existing SIEM endpoint",
		HelpText: fmt.Sprintf(`
Usage:
   • siem read %[1]s
   • siem read --name %[1]s
`, cst.ExampleSIEM),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataName, Usage: "Name of existing SIEM"},
			{Name: cst.Path, Usage: "Path to existing SIEM", Hidden: true},
		},
		MinNumberArgs: 1,
		RunFunc:       handleSiemRead,
	})
}

func GetSiemDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSiem, cst.Delete},
		SynopsisText: "Delete an existing SIEM endpoint",
		HelpText: fmt.Sprintf(`
Usage:
   • siem delete %[1]s
   • siem delete --name %[1]s
`, cst.ExampleSIEM),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataName, Usage: "Name of existing SIEM"},
			{Name: cst.Path, Usage: "Path to existing SIEM", Hidden: true},
		},
		MinNumberArgs: 1,
		RunFunc:       handleSiemDelete,
	})
}

func GetSiemSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSiem, cst.Search},
		SynopsisText: `Search for SIEM endpoints`,
		HelpText: fmt.Sprintf(`Usage:
   • %[1]s %[2]s %[3]s
   • %[1]s %[2]s --query %[3]s
`, cst.NounSiem, cst.Search, cst.ExampleSiemSearch),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Query, Shorthand: "q", Usage: fmt.Sprintf("Filter %s of items to fetch (required)", cst.Query)},
			{Name: cst.Limit, Shorthand: "l", Usage: cst.LimitHelpMessage},
			{Name: cst.Cursor, Usage: cst.CursorHelpMessage},
		},
		RunFunc: handleSiemSearchCmd,
	})
}

func handleSiemSearchCmd(vcli vaultcli.CLI, args []string) int {
	query := viper.GetString(cst.Query)
	limit := viper.GetString(cst.Limit)
	cursor := viper.GetString(cst.Cursor)
	if query == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		query = args[0]
	}

	data, apiErr := siemSearch(vcli, &siemSearchParams{query: query, limit: limit, cursor: cursor})
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleSiemCreate(vcli vaultcli.CLI, args []string) int {
	var name string
	namePrompt := &survey.Input{Message: "Name of SIEM endpoint:"}
	survErr := survey.AskOne(namePrompt, &name, survey.WithValidator(vaultcli.SurveyRequiredName))
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}
	name = strings.TrimSpace(name)

	answers, err := promptSiemData(vcli)
	if err != nil {
		vcli.Out().WriteResponse(nil, errors.New(err))
		return utils.GetExecStatus(err)
	}
	data, apiError := siemCreate(vcli, &siemCreateRequest{
		Name:            name,
		SIEMType:        answers.SIEMType,
		Host:            answers.Host,
		Port:            answers.Port,
		Protocol:        answers.Protocol,
		Endpoint:        answers.Endpoint,
		LoggingFormat:   answers.LoggingFormat,
		AuthMethod:      answers.AuthMethod,
		Auth:            answers.Auth,
		SendToEngine:    answers.SendToEngine,
		Pool:            answers.Pool,
		AllowSelfSigned: answers.AllowSelfSigned,
	})
	vcli.Out().WriteResponse(data, apiError)
	return utils.GetExecStatus(apiError)
}

func handleSiemUpdate(vcli vaultcli.CLI, args []string) int {
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name = args[0]
	}
	if name == "" {
		name = viper.GetString(cst.Path)
	}
	if name == "" {
		namePrompt := &survey.Input{Message: "Name of SIEM endpoint:"}
		survErr := survey.AskOne(namePrompt, &name, survey.WithValidator(vaultcli.SurveyRequiredName))
		if survErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(survErr))
			return utils.GetExecStatus(survErr)
		}
	}
	existedSiem, apiErr := siemRead(vcli, name)
	if apiErr != nil {
		vcli.Out().WriteResponse(nil, apiErr)
		return utils.GetExecStatus(apiErr)
	}
	vcli.Out().WriteResponse(existedSiem, nil)

	answers, err := promptSiemData(vcli)
	if err != nil {
		vcli.Out().WriteResponse(nil, errors.New(err))
		return utils.GetExecStatus(err)
	}

	data, apiErr := siemUpdate(vcli, name, answers)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func promptSiemData(vcli vaultcli.CLI) (*siemUpdateRequest, error) {
	selectType := &survey.Select{
		Message: "Select SIEM type:",
		Options: []string{
			"syslog",
			"splunk",
			"json",
			"cef",
		},
	}
	var siemType string
	survErr := survey.AskOne(selectType, &siemType)
	if survErr != nil {
		return nil, errors.New(survErr)
	}

	protocolOptions, loggingFormat := []string{}, ""
	switch siemType {
	case "syslog":
		protocolOptions = append(protocolOptions, "tls", "tcp", "udp")
		loggingFormat = "rfc5424"
	case "splunk":
		protocolOptions = append(protocolOptions, "https")
		loggingFormat = "json"
	case "json":
		protocolOptions = append(protocolOptions, "http", "https", "udp", "tcp")
		loggingFormat = "json"
	case "cef":
		protocolOptions = append(protocolOptions, "tcp", "tls", "udp")
		loggingFormat = "cef"
	default:
		return nil, fmt.Errorf("unknown siem type")
	}

	selectProtocol := &survey.Select{
		Message: fmt.Sprintf("Select protocol for %s SIEM type:", siemType),
		Options: protocolOptions,
	}
	var protocol string
	survErr = survey.AskOne(selectProtocol, &protocol)
	if survErr != nil {
		return nil, errors.New(survErr)
	}

	questionHost := &survey.Question{
		Name:      "Host",
		Prompt:    &survey.Input{Message: "Host:"},
		Validate:  vaultcli.SurveyRequired,
		Transform: vaultcli.SurveyTrimSpace,
	}
	questionPort := &survey.Question{
		Name:     "Port",
		Prompt:   &survey.Input{Message: "Port:"},
		Validate: vaultcli.SurveyRequiredPortNumber,
		Transform: func(ans interface{}) (newAns interface{}) {
			answer := strings.TrimSpace(ans.(string))
			val, _ := strconv.Atoi(answer)
			return val
		},
	}
	questionEndpoint := &survey.Question{
		Name:      "Endpoint",
		Prompt:    &survey.Input{Message: "Endpoint:"},
		Transform: vaultcli.SurveyTrimSpace,
	}
	questionAuth := &survey.Question{
		Name: "AuthMethod",
		Prompt: &survey.Select{
			Message: "Select authentication method:",
			Options: []string{"token"},
		},
	}
	questionLoggingFormat := &survey.Question{
		Name: "LoggingFormat",
		Prompt: &survey.Select{
			Message: "Select logging format:",
			Options: []string{loggingFormat},
		},
	}
	questionSendToEngine := &survey.Question{
		Name: "SendToEngine",
		Prompt: &survey.Confirm{
			Message: "Route through DSV engine:",
			Default: false,
		},
	}
	questionAllowSelfSigned := &survey.Question{
		Name: "AllowSelfSigned",
		Prompt: &survey.Confirm{
			Message: "Allow self signed (for https):",
			Default: false,
		},
	}
	qs := []*survey.Question{questionHost, questionPort}
	if protocol == "http" || protocol == "https" {
		qs = append(qs, questionEndpoint)
	}
	qs = append(qs, questionAuth, questionLoggingFormat, questionSendToEngine)
	if protocol == "https" {
		qs = append(qs, questionAllowSelfSigned)
	}
	answers := siemUpdateRequest{SIEMType: siemType, Protocol: protocol}
	survErr = survey.Ask(qs, &answers)
	if survErr != nil {
		return nil, errors.New(survErr)
	}
	if answers.SendToEngine {
		poolPrompt := &survey.Input{Message: "Engine pool:"}
		survErr := survey.AskOne(poolPrompt, &answers.Pool, survey.WithValidator(vaultcli.SurveyRequiredName))
		if survErr != nil {
			return nil, errors.New(survErr)
		}
	}
	return &answers, nil
}

func handleSiemRead(vcli vaultcli.CLI, args []string) int {
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name = args[0]
	}
	if name == "" {
		name = viper.GetString(cst.Path)
	}
	if name == "" {
		vcli.Out().FailF("error: must specify %s", cst.DataName)
		return 1
	}
	data, apiErr := siemRead(vcli, name)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleSiemDelete(vcli vaultcli.CLI, args []string) int {
	name := viper.GetString(cst.DataName)
	if name == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name = args[0]
	}
	if name == "" {
		name = viper.GetString(cst.Path)
	}
	if name == "" {
		vcli.Out().FailF("error: must specify %s", cst.DataName)
		return 1
	}
	data, apiErr := siemDelete(vcli, name)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

// API callers:

type siemCreateRequest struct {
	SIEMType        string `json:"siemType"`
	Name            string `json:"name"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Protocol        string `json:"protocol"`
	Endpoint        string `json:"endpoint"`
	LoggingFormat   string `json:"loggingFormat"`
	AuthMethod      string `json:"authMethod"`
	Auth            string `json:"auth"`
	SendToEngine    bool   `json:"sendToEngine"`
	Pool            string `json:"pool"`
	AllowSelfSigned bool   `json:"allowSelfSigned"`
}

func siemCreate(vcli vaultcli.CLI, body *siemCreateRequest) ([]byte, *errors.ApiError) {
	basePath := strings.Join([]string{cst.Config, cst.NounSiem}, "/")
	uri := paths.CreateURI(basePath, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
}

type siemUpdateRequest struct {
	SIEMType        string `json:"siemType"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Protocol        string `json:"protocol"`
	Endpoint        string `json:"endpoint"`
	LoggingFormat   string `json:"loggingFormat"`
	AuthMethod      string `json:"authMethod"`
	Auth            string `json:"auth"`
	SendToEngine    bool   `json:"sendToEngine"`
	Pool            string `json:"pool"`
	AllowSelfSigned bool   `json:"allowSelfSigned"`
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

type siemSearchParams struct {
	query  string
	limit  string
	cursor string
}

func siemSearch(vcli vaultcli.CLI, p *siemSearchParams) ([]byte, *errors.ApiError) {
	queryParams := map[string]string{}
	if p.query != "" {
		queryParams[cst.SearchKey] = p.query
	}
	if p.limit != "" {
		queryParams[cst.Limit] = p.limit
	}
	if p.cursor != "" {
		queryParams[cst.Cursor] = p.cursor
	}
	baseType := strings.Join([]string{cst.Config, cst.NounSiem}, "/")
	uri := paths.CreateResourceURI(baseType, "", "", false, queryParams)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}
