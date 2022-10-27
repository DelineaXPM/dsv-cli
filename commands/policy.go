package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	cst "thy/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/internal/predictor"
	"github.com/DelineaXPM/dsv-cli/paths"
	"github.com/DelineaXPM/dsv-cli/utils"
	"github.com/DelineaXPM/dsv-cli/vaultcli"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

func GetPolicyCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPolicy},
		SynopsisText: "Manage policies",
		HelpText: fmt.Sprintf(`Execute an action on a policy at a path

Usage:
   • policy %[1]s
   • policy --path %[1]s
`, cst.ExamplePolicyPath),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s", cst.Path, cst.NounPolicy)},
			{Name: cst.Version, Usage: "List the current and last (n) versions"},
		},
		MinNumberArgs: 1,
		RunFunc: func(vcli vaultcli.CLI, args []string) int {
			path := viper.GetString(cst.Path)
			if path == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
				path = args[0]
			}
			if path == "" {
				return cli.RunResultHelp
			}
			return handlePolicyReadCmd(vcli, args)
		},
	})
}

func GetPolicyReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPolicy, cst.Read},
		SynopsisText: "policy read (<path> | (--path | -r) <path>) [--version <n>]",
		HelpText: fmt.Sprintf(`Read a policy

Usage:
   • policy read %[1]s
   • policy read --path %[1]s
`, cst.ExamplePolicyPath),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s", cst.Path, cst.NounPolicy)},
			{Name: cst.Version, Usage: "List the current and last (n) versions"},
		},
		MinNumberArgs: 1,
		RunFunc:       handlePolicyReadCmd,
	})
}

func GetPolicyEditCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPolicy, cst.Edit},
		SynopsisText: "policy edit (<path> | (--path | -r) <path>)",
		HelpText: fmt.Sprintf(`Edit a policy

Usage:
   • policy edit %[1]s
   • policy edit --path %[1]s
`, cst.ExamplePolicyPath),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s", cst.Path, cst.NounPolicy)},
		},
		MinNumberArgs: 1,
		RunFunc:       handlePolicyEditCmd,
	})
}

func GetPolicyDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPolicy, cst.Delete},
		SynopsisText: "policy delete (<path> | (--path | -r) <path>) [--force]",
		HelpText: fmt.Sprintf(`Delete policy

Usage:
   • policy delete %[1]s
   • policy delete --path %[1]s --force
`, cst.ExamplePolicyPath),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s (required)", cst.Path)},
			{Name: cst.Force, Usage: fmt.Sprintf("Immediately delete %s", cst.NounPolicy), ValueType: "bool"},
		},
		MinNumberArgs: 1,
		RunFunc:       handlePolicyDeleteCmd,
	})
}

func GetPolicyRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPolicy, cst.Restore},
		SynopsisText: "policy restore (<path> | (--path | -r) <path>)",
		HelpText: fmt.Sprintf(`Restore a deleted policy

Usage:
   • policy restore %[1]s
`, cst.ExamplePolicyPath),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s", cst.Path, cst.NounPolicy)},
		},
		MinNumberArgs: 1,
		RunFunc:       handlePolicyRestoreCmd,
	})
}

func GetPolicyCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPolicy, cst.Create},
		SynopsisText: "policy create (<path> | --path|-r) ((--data|-d) | --subjects --actions --effect[default:allow] --resources [--desc] [--cidr])",
		HelpText: fmt.Sprintf(`Add a policy

Usage:
   • policy create %[1]s --subjects '<users:kadmin|groups:admin>',users:userA --actions create,update --cidr 10.10.0.15/24
   • policy create --path %[1]s --data %[2]s
`, cst.ExamplePolicyPath, cst.ExampleDataPath),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("%s to be stored in a %s. Prefix with '@' to denote filepath (required)", strings.Title(cst.Data), cst.NounPolicy), Predictor: predictor.NewPrefixFilePredictor("*")},
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounPolicy)},
			{Name: cst.DataAction, Usage: fmt.Sprintf("Policy actions to be stored in a %s (required, regex and list supported)(required)", cst.NounPolicy), Predictor: predictor.ActionTypePredictor{}},
			{Name: cst.DataEffect, Usage: fmt.Sprintf("Policy effect to be stored in a %s. Defaults to allow if not specified", cst.NounPolicy), Default: "allow", Predictor: predictor.EffectTypePredictor{}},
			{Name: cst.DataDescription, Usage: fmt.Sprintf("Policy description to be stored in a %s ", cst.NounPolicy)},
			{Name: cst.DataSubject, Usage: fmt.Sprintf("Policy subjects to be stored in a %s (required, regex and list supported)(required)", cst.NounPolicy)},
			{Name: cst.DataCidr, Usage: fmt.Sprintf("Policy CIDR condition remote IP to be stored in a %s ", cst.NounPolicy)},
			{Name: cst.DataResource, Usage: fmt.Sprintf("Policy resources to be stored in a %s. Defaults to the path plus all paths below (<.*>) ", cst.NounPolicy)},
		},
		RunFunc:    handlePolicyCreateCmd,
		WizardFunc: handlePolicyCreateWizard,
	})
}

func GetPolicyUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPolicy, cst.Update},
		SynopsisText: "policy update (<path> | (--path | -r) <path>) ((--data|-d) | --subjects --actions --effect[default:allow] --resources [--desc] [--cidr])",
		HelpText: fmt.Sprintf(`Update a policy

Policy Updates are all or nothing, so required fields must be included in the update and if optional fields are not included, they are deleted or go to default.

Usage:
   • policy update %[1]s --subjects 'users:<kadmin|groups:admin>',users:userA --actions update --cidr 192.168.0.15/24
   • policy update --path %[1]s --data %[2]s
`, cst.ExamplePolicyPath, cst.ExampleDataPath),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("%s to be stored in a %s. Prefix with '@' to denote filepath (required)", strings.Title(cst.Data), cst.NounPolicy), Predictor: predictor.NewPrefixFilePredictor("*")},
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounPolicy)},
			{Name: cst.DataAction, Usage: fmt.Sprintf("Policy actions to be stored in a %s (required, regex and list supported)(required)", cst.NounPolicy), Predictor: predictor.ActionTypePredictor{}},
			{Name: cst.DataEffect, Usage: fmt.Sprintf("Policy effect to be stored in a %s. Defaults to allow if not specified", cst.NounPolicy), Default: "allow", Predictor: predictor.EffectTypePredictor{}},
			{Name: cst.DataDescription, Usage: fmt.Sprintf("Policy description to be stored in a %s ", cst.NounPolicy)},
			{Name: cst.DataSubject, Usage: fmt.Sprintf("Policy subjects to be stored in a %s (required, regex and list supported)(required)", cst.NounPolicy)},
			{Name: cst.DataCidr, Usage: fmt.Sprintf("Policy CIDR condition remote IP to be stored in a %s ", cst.NounPolicy)},
			{Name: cst.DataResource, Usage: fmt.Sprintf("Policy resources to be stored in a %s. Defaults to the path plus all paths below (<.*>) ", cst.NounPolicy)},
		},
		RunFunc:    handlePolicyUpdateCmd,
		WizardFunc: handlePolicyUpdateWizard,
	})
}

func GetPolicyRollbackCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPolicy, cst.Rollback},
		SynopsisText: "policy rollback (<path> | (--path | -r) <path>) [--version <n>]",
		HelpText: fmt.Sprintf(`Rollback a policy

Usage:
   • policy rollback %[1]s --version 1
   • policy rollback --path %[1]s
`, cst.ExamplePolicyPath),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounPolicy)},
			{Name: cst.Version, Usage: "The version to which to rollback"},
		},
		MinNumberArgs: 1,
		RunFunc:       handlePolicyRollbackCmd,
	})
}

func GetPolicySearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPolicy, cst.Search},
		SynopsisText: "policy search (<query> | (--query | -q) <query>) [(--limit | -l) <n>] [--cursor <cursor>]",
		HelpText: fmt.Sprintf(`Search for a policy

Usage:
   • policy search %[1]s
   • policy search --query %[1]s
`, cst.ExamplePolicySearch),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Query, Shorthand: "q", Usage: fmt.Sprintf("Filter %s of items to fetch (required)", strings.Title(cst.Query))},
			{Name: cst.Limit, Shorthand: "l", Usage: cst.LimitHelpMessage},
			{Name: cst.Cursor, Usage: cst.CursorHelpMessage},
		},
		RunFunc: handlePolicySearchCmd,
	})
}

func handlePolicyReadCmd(vcli vaultcli.CLI, args []string) int {
	path, status := getPolicyParams(args)
	if status != 0 {
		return status
	}

	path = paths.ProcessResource(path)
	version := viper.GetString(cst.Version)
	if strings.TrimSpace(version) != "" {
		path = fmt.Sprint(path, "/", cst.Version, "/", version)
	}

	data, apiErr := policyRead(vcli, path)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handlePolicyEditCmd(vcli vaultcli.CLI, args []string) int {
	path, status := getPolicyParams(args)
	if status != 0 {
		return status
	}
	path = paths.ProcessResource(path)

	data, apiErr := policyRead(vcli, path)
	if apiErr != nil {
		vcli.Out().WriteResponse(data, apiErr)
		return utils.GetExecStatus(apiErr)
	}

	saveFunc := func(data []byte) (resp []byte, err *errors.ApiError) {
		body := &policyUpdateRequest{
			Policy:        string(data),
			Serialization: viper.GetString(cst.Encoding),
		}
		_, err = policyUpdate(vcli, path, body)
		return nil, err
	}
	resp, err := vcli.Edit(data, saveFunc)
	vcli.Out().WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func handlePolicyDeleteCmd(vcli vaultcli.CLI, args []string) int {
	force := viper.GetBool(cst.Force)
	path, status := getPolicyParams(args)
	if status != 0 {
		return status
	}
	data, apiErr := policyDelete(vcli, paths.ProcessResource(path), force)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handlePolicyRestoreCmd(vcli vaultcli.CLI, args []string) int {
	path := viper.GetString(cst.Path)
	if path == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		path = args[0]
	}
	data, apiErr := policyRestore(vcli, paths.ProcessResource(path))
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handlePolicyRollbackCmd(vcli vaultcli.CLI, args []string) int {
	path := viper.GetString(cst.Path)
	if path == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		path = args[0]
	}
	version := viper.GetString(cst.Version)

	// If version is not provided, get the current policy item and parse the version from it.
	// Submit a request for a version that's previous relative to the one found.
	if version == "" {
		data, apiErr := policyRead(vcli, paths.ProcessResource(path))
		if apiErr != nil {
			vcli.Out().WriteResponse(data, apiErr)
			return utils.GetExecStatus(apiErr)
		}

		v, err := utils.GetPreviousVersion(data)
		if err != nil {
			vcli.Out().Fail(err)
			return 1
		}
		version = v
	}

	data, apiErr := policyRollback(vcli, paths.ProcessResource(path), version)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handlePolicyCreateCmd(vcli vaultcli.CLI, args []string) int {
	path, status := getPolicyParams(args)
	if status != 0 {
		return status
	}
	if err := vaultcli.ValidatePath(path); err != nil {
		vcli.Out().FailF("Path %q is invalid: %v", path, err)
		return utils.GetExecStatus(err)
	}

	data := viper.GetString(cst.Data)
	encoding := viper.GetString(cst.Encoding)

	if data == "" {
		var err *errors.ApiError
		data, err = policyBuildFromFlags()
		if err != nil {
			vcli.Out().FailE(err)
			return utils.GetExecStatus(err)
		}
		encoding = cst.Json
	}

	body := &policyCreateRequest{
		Path:          path,
		Policy:        data,
		Serialization: encoding,
	}
	resp, apiErr := policyCreate(vcli, body)
	vcli.Out().WriteResponse(resp, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handlePolicyUpdateCmd(vcli vaultcli.CLI, args []string) int {
	path, status := getPolicyParams(args)
	if status != 0 {
		return status
	}

	data := viper.GetString(cst.Data)
	encoding := viper.GetString(cst.Encoding)

	if data == "" {
		var err *errors.ApiError
		data, err = policyBuildFromFlags()
		if err != nil {
			vcli.Out().FailE(err)
			return utils.GetExecStatus(err)
		}
		encoding = cst.Json
	}

	body := &policyUpdateRequest{
		Policy:        data,
		Serialization: encoding,
	}
	resp, apiErr := policyUpdate(vcli, path, body)
	vcli.Out().WriteResponse(resp, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handlePolicySearchCmd(vcli vaultcli.CLI, args []string) int {
	query := viper.GetString(cst.Query)
	limit := viper.GetString(cst.Limit)
	cursor := viper.GetString(cst.Cursor)
	if query == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		query = args[0]
	}

	data, apiErr := policySearch(vcli, &policySearchParams{query: query, limit: limit, cursor: cursor})
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

// Helpers:

func createPolicy(params []map[string]string) (string, error) {
	permissions := make([]*defaultPolicy, 0, len(params))
	for _, param := range params {
		policy := defaultPolicy{
			Description: param[cst.DataDescription],
			Effect:      param[cst.DataEffect],
			Actions:     utils.StringToSlice(param[cst.DataAction]),
			Subjects:    utils.StringToSlice(param[cst.DataSubject]),
		}
		if resources := param[cst.DataResource]; resources != "" {
			policy.Resources = utils.StringToSlice(param[cst.DataResource])
		}

		if param[cst.DataCidr] != "" {
			if err := setCidrCondition(&policy, param[cst.DataCidr]); err != nil {
				return "", err
			}
		}
		permissions = append(permissions, &policy)
	}

	doc := map[string][]*defaultPolicy{
		"permissionDocument": permissions,
	}
	marshalled, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}

	return string(marshalled), nil
}

func policyBuildFromFlags() (string, *errors.ApiError) {
	effect := viper.GetString(cst.DataEffect)
	if effect == "" {
		effect = "allow"
	}

	params := map[string]string{
		cst.DataDescription: viper.GetString(cst.DataDescription),
		cst.DataEffect:      effect,
		cst.DataAction:      viper.GetString(cst.DataAction),
		cst.DataSubject:     viper.GetString(cst.DataSubject),
		cst.DataResource:    viper.GetString(cst.DataResource),
		cst.DataCidr:        viper.GetString(cst.DataCidr),
	}

	err := ValidateParams(params, []string{cst.DataAction, cst.DataSubject})
	if err != nil {
		return "", err
	}

	data, e := createPolicy([]map[string]string{params})
	if e != nil {
		return "", errors.New(e)
	}
	return data, nil
}

func getPolicyParams(args []string) (path string, status int) {
	status = 0
	path = viper.GetString(cst.Path)
	if path == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		path = args[0]
	}
	if path == "" {
		status = cli.RunResultHelp
	}
	return path, status
}

type CIDRCondition struct {
	CIDR string `json:"cidr"`
}

type jsonCondition struct {
	Type    string          `json:"type"`
	Options json.RawMessage `json:"options"`
}

type defaultPolicy struct {
	ID          string                   `json:"id"`
	Description string                   `json:"description"`
	Subjects    []string                 `json:"subjects"`
	Effect      string                   `json:"effect"`
	Resources   []string                 `json:"resources"`
	Actions     []string                 `json:"actions"`
	Conditions  map[string]jsonCondition `json:"conditions"`
}

func setCidrCondition(policy *defaultPolicy, cidr string) *errors.ApiError {
	if _, _, err := net.ParseCIDR(cidr); err != nil {
		return errors.New(err)
	}
	cidrCondition := CIDRCondition{CIDR: cidr}
	if raw, err := json.Marshal(cidrCondition); err != nil {
		return errors.New(err).Grow("Failed to serialize cidr condition")
	} else {
		jc := jsonCondition{
			Type:    "CIDRCondition",
			Options: json.RawMessage(raw),
		}
		if policy.Conditions == nil {
			policy.Conditions = make(map[string]jsonCondition, 1)
		}
		policy.Conditions["remoteIP"] = jc
		return nil
	}
}

// Wizards:

func handlePolicyCreateWizard(vcli vaultcli.CLI) int {
	pathPrompt := &survey.Input{Message: "Path to policy:"}
	pathValidation := func(ans interface{}) error {
		answer := strings.TrimSpace(ans.(string))
		if len(answer) == 0 {
			return errors.NewS("Value is required")
		}
		if err := vaultcli.ValidatePath(answer); err != nil {
			return err
		}
		answer = paths.ProcessResource(answer)
		_, apiError := policyRead(vcli, answer)
		if apiError == nil {
			return errors.NewS("A policy with this path already exists.")
		}
		return nil
	}
	var path string
	survErr := survey.AskOne(pathPrompt, &path, survey.WithValidator(pathValidation))
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}
	path = paths.ProcessResource(strings.TrimSpace(path))

	policy, e := policyCollectDataWizard()
	if e != nil {
		vcli.Out().Fail(e)
		return utils.GetExecStatus(e)
	}

	body := &policyCreateRequest{
		Path:          path,
		Policy:        policy,
		Serialization: cst.Json,
	}
	resp, apiErr := policyCreate(vcli, body)
	vcli.Out().WriteResponse(resp, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handlePolicyUpdateWizard(vcli vaultcli.CLI) int {
	var policyAtPath []byte
	pathPrompt := &survey.Input{Message: "Path to policy:"}
	pathValidation := func(ans interface{}) error {
		answer := strings.TrimSpace(ans.(string))
		if len(answer) == 0 {
			return errors.NewS("Value is required")
		}
		answer = paths.ProcessResource(answer)
		var apiError *errors.ApiError
		policyAtPath, apiError = policyRead(vcli, answer)
		if apiError != nil &&
			apiError.HttpResponse() != nil &&
			apiError.HttpResponse().StatusCode == http.StatusNotFound {
			return errors.NewS("A policy with this path does not exist.")
		}
		return nil
	}
	var path string
	survErr := survey.AskOne(pathPrompt, &path, survey.WithValidator(pathValidation))
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}
	path = paths.ProcessResource(strings.TrimSpace(path))
	vcli.Out().WriteResponse(policyAtPath, nil)

	policy, e := policyCollectDataWizard()
	if e != nil {
		vcli.Out().Fail(e)
		return utils.GetExecStatus(e)
	}

	body := &policyUpdateRequest{
		Policy:        policy,
		Serialization: cst.Json,
	}
	resp, apiErr := policyUpdate(vcli, path, body)
	vcli.Out().WriteResponse(resp, apiErr)
	return utils.GetExecStatus(apiErr)
}

func policyCollectDataWizard() (string, error) {
	permissions := make([]map[string]string, 0)

	for {
		qs := []*survey.Question{
			{
				Name:      "Description",
				Prompt:    &survey.Input{Message: "Description of policy:"},
				Transform: vaultcli.SurveyTrimSpace,
			},
			{
				Name: "Actions",
				Prompt: &survey.MultiSelect{
					Message: "Actions:",
					Options: []string{"create", "read", "update", "delete", "list", "assign"},
				},
				Validate: vaultcli.SurveySelectAtLeastOne,
			},
			{
				Name:   "Effect",
				Prompt: &survey.Select{Message: "Effect:", Options: []string{"allow", "deny"}},
			},
		}

		answers := struct {
			Description string
			Actions     []string
			Effect      string
		}{}
		survErr := survey.Ask(qs, &answers)
		if survErr != nil {
			return "", survErr
		}

		subjects := []string{}
		for {
			qs := []*survey.Question{
				{
					Name:   "Subject",
					Prompt: &survey.Input{Message: "Subject:"},
					Validate: func(ans interface{}) error {
						answer := strings.TrimSpace(ans.(string))
						if len(answer) == 0 && len(subjects) == 0 {
							return errors.NewS("Must include at least one subject.")
						}
						return nil
					},
					Transform: vaultcli.SurveyTrimSpace,
				},
				{Name: "addMore", Prompt: &survey.Confirm{Message: "Add another subject?", Default: true}},
			}
			answers := struct {
				Subject string
				AddMore bool
			}{}
			survErr := survey.Ask(qs, &answers)
			if survErr != nil {
				return "", survErr
			}
			if answers.Subject != "" {
				subjects = append(subjects, answers.Subject)
			}
			if !answers.AddMore {
				break
			}
		}

		recources := []string{}

		var resource string
		resourcePrompt := &survey.Input{Message: "Resource:"}
		survErr = survey.AskOne(resourcePrompt, &resource)
		if survErr != nil {
			return "", survErr
		}
		if resource != "" {
			recources = append(recources, resource)

			var confirm bool
			confirmPrompt := &survey.Confirm{Message: "Add more?:", Default: true}
			survErr = survey.AskOne(confirmPrompt, &confirm)
			if survErr != nil {
				return "", survErr
			}

			for confirm {
				qs := []*survey.Question{
					{
						Name:      "Resource",
						Prompt:    &survey.Input{Message: "Resource:"},
						Transform: vaultcli.SurveyTrimSpace,
					},
					{Name: "addMore", Prompt: &survey.Confirm{Message: "Add another resource?", Default: true}},
				}
				answers := struct {
					Resource string
					AddMore  bool
				}{}
				survErr := survey.Ask(qs, &answers)
				if survErr != nil {
					return "", survErr
				}
				if answers.Resource != "" {
					recources = append(recources, answers.Resource)
				}
				if !answers.AddMore {
					break
				}
			}
		}

		var cidr string
		cidrPrompt := &survey.Input{Message: "CIDR range to lock down access to a specific IP range:"}
		survErr = survey.AskOne(cidrPrompt, &cidr, survey.WithValidator(vaultcli.SurveyOptionalCIDR))
		if survErr != nil {
			return "", survErr
		}

		params := map[string]string{
			cst.DataDescription: answers.Description,
			cst.DataAction:      strings.Join(answers.Actions, ","),
			cst.DataEffect:      answers.Effect,
			cst.DataSubject:     strings.Join(subjects, ","),
			cst.DataResource:    strings.Join(recources, ","),
			cst.DataCidr:        strings.TrimSpace(cidr),
		}

		permissions = append(permissions, params)

		var addMore bool
		addMorePrompt := &survey.Confirm{Message: "Add another permission?", Default: true}
		survErr = survey.AskOne(addMorePrompt, &addMore)
		if survErr != nil {
			return "", survErr
		}
		if !addMore {
			break
		}
	}

	return createPolicy(permissions)
}

// API callers:

type policyCreateRequest struct {
	Path          string `json:"path"`
	Policy        string `json:"policy"`
	Serialization string `json:"serialization"`
}

func policyCreate(vcli vaultcli.CLI, body *policyCreateRequest) ([]byte, *errors.ApiError) {
	baseType := strings.Join([]string{cst.Config, cst.NounPolicies}, "/")
	uri := paths.CreateResourceURI(baseType, "", "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
}

func policyRead(vcli vaultcli.CLI, path string) ([]byte, *errors.ApiError) {
	baseType := strings.Join([]string{cst.Config, cst.NounPolicies}, "/")
	uri := paths.CreateResourceURI(baseType, path, "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}

type policyUpdateRequest struct {
	Policy        string `json:"policy"`
	Serialization string `json:"serialization"`
}

func policyUpdate(vcli vaultcli.CLI, path string, body *policyUpdateRequest) ([]byte, *errors.ApiError) {
	baseType := strings.Join([]string{cst.Config, cst.NounPolicies}, "/")
	uri := paths.CreateResourceURI(baseType, path, "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPut, uri, body)
}

func policyRollback(vcli vaultcli.CLI, path string, version string) ([]byte, *errors.ApiError) {
	baseType := strings.Join([]string{cst.Config, cst.NounPolicies}, "/")
	path = fmt.Sprintf("%s/rollback/%s", path, version)
	uri := paths.CreateResourceURI(baseType, path, "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPut, uri, nil)
}

func policyDelete(vcli vaultcli.CLI, path string, force bool) ([]byte, *errors.ApiError) {
	query := map[string]string{"force": strconv.FormatBool(force)}
	baseType := strings.Join([]string{cst.Config, cst.NounPolicies}, "/")
	uri := paths.CreateResourceURI(baseType, path, "", true, query)
	return vcli.HTTPClient().DoRequest(http.MethodDelete, uri, nil)
}

func policyRestore(vcli vaultcli.CLI, path string) ([]byte, *errors.ApiError) {
	baseType := strings.Join([]string{cst.Config, cst.NounPolicies}, "/")
	uri := paths.CreateResourceURI(baseType, path, "/restore", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPut, uri, nil)
}

type policySearchParams struct {
	query  string
	limit  string
	cursor string
}

func policySearch(vcli vaultcli.CLI, p *policySearchParams) ([]byte, *errors.ApiError) {
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
	baseType := strings.Join([]string{cst.Config, cst.NounPolicies}, "/")
	uri := paths.CreateResourceURI(baseType, "", "", false, queryParams)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}
