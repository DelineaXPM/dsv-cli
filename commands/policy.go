package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"thy/constants"
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

type Policy struct {
	request   requests.Client
	outClient format.OutClient
	edit      func([]byte, dataFunc, *errors.ApiError, bool) ([]byte, *errors.ApiError)
}

func GetNoDataOpPolicyWrappers() cli.PredictorWrappers {
	return cli.PredictorWrappers{
		preds.LongFlag(cst.Path):    cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s", cst.Path, cst.NounPolicy)}), false},
		preds.LongFlag(cst.Version): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Version, Shorthand: "", Usage: "List the current and last (n) versions"}), false},
	}
}

func GetPolicyCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounPolicy},
		RunFunc: func(args []string) int {
			path := viper.GetString(cst.Path)
			if path == "" {
				path = paths.GetPath(args)
			}
			if path == "" {
				return cli.RunResultHelp
			}
			return Policy{requests.NewHttpClient(), nil, EditData}.handlePolicyReadCmd(args)
		},
		SynopsisText: "policy (<path> | --path|-r)",
		HelpText: fmt.Sprintf(`Execute an action on a %[1]s at a path

Usage:
   • %[1]s %[2]s
   • %[1]s --path %[2]s
		`, cst.NounPolicy, cst.ExamplePolicyPath),
		FlagsPredictor: GetNoDataOpPolicyWrappers(),
		MinNumberArgs:  1,
	})
}

func GetPolicyReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPolicy, cst.Read},
		RunFunc:      Policy{requests.NewHttpClient(), nil, EditData}.handlePolicyReadCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounPolicy, cst.Read),
		HelpText: fmt.Sprintf(`Read a %[1]s

Usage:
   • %[1]s %[3]s %[2]s
   • %[1]s %[3]s --path %[2]s
		`, cst.NounPolicy, cst.ExamplePolicyPath, cst.Read),
		FlagsPredictor:    GetNoDataOpPolicyWrappers(),
		ArgsPredictorFunc: preds.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
	})
}

func GetPolicyEditCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounPolicy, cst.Edit},
		RunFunc: Policy{
			requests.NewHttpClient(),
			nil,
			EditData}.handlePolicyEditCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounPolicy, cst.Edit),
		HelpText: fmt.Sprintf(`Edit a %[1]s

Usage:
   • %[1]s %[3]s %[2]s
   • %[1]s %[3]s --path %[2]s
		`, cst.NounPolicy, cst.ExamplePolicyPath, cst.Edit),
		FlagsPredictor:    GetNoDataOpPolicyWrappers(),
		ArgsPredictorFunc: preds.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
	})
}

func GetPolicyDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounPolicy, cst.Delete},
		RunFunc: Policy{
			requests.NewHttpClient(),
			nil,
			EditData}.handlePolicyDeleteCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounPolicy, cst.Delete),
		HelpText: fmt.Sprintf(`Delete %[1]s

Usage:
   • %[1]s %[3]s %[2]s
   • %[1]s %[3]s --path %[2]s --force
		`, cst.NounPolicy, cst.ExamplePolicyPath, cst.Delete),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):  cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s (required)", cst.Path)}), false},
			preds.LongFlag(cst.Force): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Force, Shorthand: "", Usage: fmt.Sprintf("Immediately delete %s", cst.NounPolicy), Global: false, ValueType: "bool"}), false},
		},
		ArgsPredictorFunc: preds.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
	})
}

func GetPolicyRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounPolicy, cst.Read},
		RunFunc: Policy{
			requests.NewHttpClient(),
			nil,
			EditData}.handlePolicyRestoreCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounPolicy, cst.Restore),
		HelpText: fmt.Sprintf(`Restore a deleted %[2]s from %[3]s
Usage:
	• policy %[1]s %[4]s

				`, cst.Restore, cst.NounPolicy, cst.ProductName, cst.ExamplePath),
		FlagsPredictor:    GetNoDataOpPolicyWrappers(),
		ArgsPredictorFunc: preds.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
	})
}

func GetPolicyCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPolicy, cst.Create},
		RunFunc:      Policy{requests.NewHttpClient(), nil, EditData}.handlePolicyUpsertCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r) ((--data|-d) | --subjects --actions --effect[default:allow] --desc --cidr  --resources)", cst.NounPolicy, cst.Create),
		HelpText: fmt.Sprintf(`Add a %[1]s

Usage:
   • %[1]s %[3]s %[2]s --subjects 'users:<kadmin|groups:admin>',users:userA --actions create,update --cidr 192.168.0.15/24
   • %[1]s %[3]s --path %[2]s --data %[4]s
		`, cst.NounPolicy, cst.ExamplePolicyPath, cst.Create, cst.ExampleDataPath),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Data):            cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("%s to be stored in a %s. Prefix with '@' to denote filepath (required)", strings.Title(cst.Data), cst.NounPolicy)}), false},
			preds.LongFlag(cst.Path):            cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounPolicy)}), false},
			preds.LongFlag(cst.DataAction):      cli.PredictorWrapper{preds.ActionTypePredictor{}, preds.NewFlagValue(preds.Params{Name: cst.DataAction, Usage: fmt.Sprintf("Policy actions to be stored in a %s (required, regex and list supported)(required)", cst.NounPolicy)}), false},
			preds.LongFlag(cst.DataEffect):      cli.PredictorWrapper{preds.EffectTypePredictor{}, preds.NewFlagValue(preds.Params{Name: cst.DataEffect, Usage: fmt.Sprintf("Policy effect to be stored in a %s. Defaults to allow if not specified", cst.NounPolicy), Default: "allow"}), false},
			preds.LongFlag(cst.DataDescription): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataDescription, Usage: fmt.Sprintf("Policy description to be stored in a %s ", cst.NounPolicy)}), false},
			preds.LongFlag(cst.DataSubject):     cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataSubject, Usage: fmt.Sprintf("Policy subjects to be stored in a %s (required, regex and list supported)(required)", cst.NounPolicy)}), false},
			preds.LongFlag(cst.DataCidr):        cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataCidr, Usage: fmt.Sprintf("Policy CIDR condition remote IP to be stored in a %s ", cst.NounPolicy)}), false},
			preds.LongFlag(cst.DataResource):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataResource, Usage: fmt.Sprintf("Policy resources to be stored in a %s. Defaults to the path plus all paths below (<.*>) ", cst.NounPolicy)}), false},
		},
		ArgsPredictorFunc: preds.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     0,
	})
}

func GetPolicyUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPolicy, cst.Update},
		RunFunc:      Policy{requests.NewHttpClient(), nil, EditData}.handlePolicyUpsertCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r )  ((--data|-d) | --subjects --actions --effect[default:allow] --desc --cidr --resources)", cst.NounPolicy, cst.Update),
		HelpText: fmt.Sprintf(`Update a %[1]s

Usage:
   • Policy Updates are all or nothing, so required fields must be included in the update and if optional fields are not included, they are deleted or go to default
   • %[1]s %[3]s %[2]s --subjects 'users:<kadmin|groups:admin>',users:userA --actions update --cidr 192.168.0.15/24
   • %[1]s %[3]s --path %[2]s --data %[4]s
		`, cst.NounPolicy, cst.ExamplePolicyPath, cst.Update, cst.ExampleDataPath),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Data):            cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("%s to be stored in a %s. Prefix with '@' to denote filepath (required)", strings.Title(cst.Data), cst.NounPolicy)}), false},
			preds.LongFlag(cst.Path):            cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounPolicy)}), false},
			preds.LongFlag(cst.DataAction):      cli.PredictorWrapper{preds.ActionTypePredictor{}, preds.NewFlagValue(preds.Params{Name: cst.DataAction, Usage: fmt.Sprintf("Policy actions to be stored in a %s (required, regex and list supported)(required)", cst.NounPolicy)}), false},
			preds.LongFlag(cst.DataEffect):      cli.PredictorWrapper{preds.EffectTypePredictor{}, preds.NewFlagValue(preds.Params{Name: cst.DataEffect, Usage: fmt.Sprintf("Policy effect to be stored in a %s. Defaults to allow if not specified", cst.NounPolicy), Default: "allow"}), false},
			preds.LongFlag(cst.DataDescription): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataDescription, Usage: fmt.Sprintf("Policy description to be stored in a %s ", cst.NounPolicy)}), false},
			preds.LongFlag(cst.DataSubject):     cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataSubject, Usage: fmt.Sprintf("Policy subjects to be stored in a %s (required, regex and list supported)(required)", cst.NounPolicy)}), false},
			preds.LongFlag(cst.DataCidr):        cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataCidr, Usage: fmt.Sprintf("Policy CIDR condition remote IP to be stored in a %s ", cst.NounPolicy)}), false},
			preds.LongFlag(cst.DataResource):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataResource, Usage: fmt.Sprintf("Policy resources to be stored in a %s. Defaults to the path plus all paths below (<.*>) ", cst.NounPolicy)}), false},
		},
		ArgsPredictorFunc: preds.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     0,
	})
}

func GetPolicyRollbackCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPolicy, cst.Rollback},
		RunFunc:      Policy{requests.NewHttpClient(), nil, EditData}.handlePolicyRollbackCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounPolicy, cst.Rollback),
		HelpText: fmt.Sprintf(`Rollback a %[1]s

Usage:
   • %[1]s %[3]s %[2]s --%[4]s 1
   • %[1]s %[3]s --path %[2]s
		`, cst.NounPolicy, cst.ExamplePolicyPath, cst.Rollback, cst.Version),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):    cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounSecret)}), false},
			preds.LongFlag(cst.Version): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Version, Shorthand: "", Usage: "The version to which to rollback"}), false},
		},
		ArgsPredictorFunc: preds.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
	})
}

func GetPolicySearchCommand() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounPolicy, cst.Search},
		RunFunc:      Policy{requests.NewHttpClient(), nil, EditData}.handlePolicySearchCommand,
		SynopsisText: fmt.Sprintf("%s (<query> | --query)", cst.Search),
		HelpText: fmt.Sprintf(`Search for a %[1]s

		Usage:
		• %[1]s %[2]s %[3]s
		• %[1]s %[2]s --query %[3]s
				`, cst.NounPolicy, cst.Search, cst.ExamplePolicySearch),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Query):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Query, Shorthand: "q", Usage: fmt.Sprintf("Filter %s of items to fetch (required)", strings.Title(cst.Query))}), false},
			preds.LongFlag(cst.Limit):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Limit, Shorthand: "l", Usage: fmt.Sprint("Maximum number of results per cursor (optional)")}), false},
			preds.LongFlag(cst.Cursor): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Cursor, Usage: constants.CursorHelpMessage}), false},
		},
		MinNumberArgs: 0,
	})
}

func (p Policy) handlePolicyReadCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	path, status := getPolicyParams(args)
	if status != 0 {
		return status
	}
	version := viper.GetString(cst.Version)
	if strings.TrimSpace(version) != "" {
		path = fmt.Sprint(path, "/", cst.Version, "/", version)
	}

	baseType := strings.Join([]string{cst.Config, cst.NounPolicies}, "/")
	uri := paths.CreateResourceURI(baseType, paths.ProcessPath(path), "", true, nil, false)

	data, err = p.request.DoRequest("GET", uri, nil)

	outClient := p.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (p Policy) handlePolicyEditCmd(args []string) int {
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}

	var err *errors.ApiError
	var resp []byte

	path, status := getPolicyParams(args)
	if status != 0 {
		return status
	}
	baseType := strings.Join([]string{cst.Config, cst.NounPolicies}, "/")
	uri := paths.CreateResourceURI(baseType, paths.ProcessPath(path), "", true, nil, false)

	resp, err = p.request.DoRequest("GET", uri, nil)
	if err != nil {
		p.outClient.WriteResponse(resp, err)
		return utils.GetExecStatus(err)
	}

	saveFunc := dataFunc(func(data []byte) (resp []byte, err *errors.ApiError) {
		encoding := viper.GetString(cst.Encoding)
		model := postPolicyModel{
			Policy:        string(data),
			Serialization: encoding,
		}
		_, err = p.request.DoRequest("PUT", uri, &model)
		return nil, err
	})
	resp, err = p.edit(resp, saveFunc, nil, false)
	p.outClient.WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func (p Policy) handlePolicyDeleteCmd(args []string) int {
	var err *errors.ApiError
	var resp []byte
	force := viper.GetBool(cst.Force)
	path, status := getPolicyParams(args)
	if status != 0 {
		return status
	}

	baseType := strings.Join([]string{cst.Config, cst.NounPolicies}, "/")
	query := map[string]string{"force": strconv.FormatBool(force)}
	uri := paths.CreateResourceURI(baseType, paths.ProcessPath(path), "", true, query, false)

	resp, err = p.request.DoRequest("DELETE", uri, nil)

	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}
	p.outClient.WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func (p Policy) handlePolicyRestoreCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}
	path := viper.GetString(cst.Path)
	if path == "" {
		path = paths.GetPath(args)
	}

	baseType := strings.Join([]string{cst.Config, cst.NounPolicies}, "/")
	uri := paths.CreateResourceURI(baseType, paths.ProcessPath(path), "/restore", true, nil, false)
	data, err = p.request.DoRequest("PUT", uri, nil)

	p.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (p Policy) handlePolicyRollbackCmd(args []string) int {
	var apiError *errors.ApiError
	var resp []byte
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}
	baseType := strings.Join([]string{cst.Config, cst.NounPolicies}, "/")

	path := viper.GetString(cst.Path)
	if path == "" {
		path = paths.GetPath(args)
	}
	version := viper.GetString(cst.Version)

	// If version is not provided, get the current policy item and parse the version from it.
	// Submit a request for a version that's previous relative to the one found.
	if version == "" {
		uri := paths.CreateResourceURI(baseType, paths.ProcessPath(path), "", true, nil, false)
		resp, apiError = p.request.DoRequest("GET", uri, nil)
		if apiError != nil {
			p.outClient.WriteResponse(resp, apiError)
			return utils.GetExecStatus(apiError)
		}

		v, err := utils.GetPreviousVersion(resp)
		if err != nil {
			p.outClient.Fail(err)
			return 1
		}
		version = v
	}

	if strings.TrimSpace(version) != "" {
		path = fmt.Sprint(path, "/rollback/", version)
	}

	uri := paths.CreateResourceURI(baseType, path, "", true, nil, false)
	resp, apiError = p.request.DoRequest("PUT", uri, nil)

	p.outClient.WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func createPolicy(params map[string]string) (*postPolicyModel, error) {
	policy := defaultPolicy{
		Description: params[cst.DataDescription],
		Subjects:    utils.StringToSlice(params[cst.DataSubject]),
		Effect:      params[cst.DataEffect],
		Actions:     utils.StringToSlice(params[cst.DataAction]),
	}
	if resources := params[cst.DataResource]; resources != "" {
		policy.Resources = utils.StringToSlice(params[cst.DataResource])
	}
	if id := viper.GetString(cst.ID); id != "" {
		policy.ID = id
	}

	if params[cst.DataCidr] != "" {
		if err := setCidrCondition(&policy, params[cst.DataCidr]); err != nil {
			return nil, err
		}
	}
	doc := document{
		PermissionDocument: []*defaultPolicy{
			&policy,
		},
	}
	marshalled, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	return &postPolicyModel{
		Policy:        string(marshalled),
		Path:          params[cst.Path],
		Serialization: "json",
	}, nil
}

func (p Policy) submitPolicy(policy *postPolicyModel) ([]byte, *errors.ApiError) {
	baseType := strings.Join([]string{cst.Config, cst.NounPolicies}, "/")
	var uri string
	reqMethod := strings.ToLower(viper.GetString(cst.LastCommandKey))
	if reqMethod == cst.Create {
		reqMethod = "POST"
		uri = paths.CreateResourceURI(baseType, "", "", true, nil, false)
	} else {
		reqMethod = "PUT"
		uri = paths.CreateResourceURI(baseType, policy.Path, "", true, nil, false)
	}
	return p.request.DoRequest(reqMethod, uri, policy)
}

func (p Policy) handlePolicyUpsertWorkflow(args []string) int {
	params := make(map[string]string)
	var resp []byte
	var err *errors.ApiError
	ui := &PasswordUi{
		cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
	}

	if path, err := getStringAndValidate(
		ui, "Path to policy:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.Path] = paths.ProcessPath(path)
	}

	if viper.GetString(cst.LastCommandKey) == cst.Update {
		code := p.handlePolicyReadCmd([]string{params[cst.Path]})
		if code != 0 {
			return code
		}
	}

	if desc, err := getStringAndValidate(
		ui, "Description of policy (optional):", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.DataDescription] = desc
	}

	if action, err := getStringAndValidate(
		ui, "Allowed actions (comma-delimited strings):", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.DataAction] = action
	}

	if effect, err := getStringAndValidateDefault(
		ui, "Effect of policy (default:allow):", "allow", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.DataEffect] = effect
	}

	if resources, err := getStringAndValidate(
		ui, "Resources of policy (comma-delimited strings):", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.DataResource] = resources
	}

	if subjects, err := getStringAndValidate(
		ui, "Subjects of policy (comma-delimited strings):", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.DataSubject] = subjects
	}

	if cidr, err := getStringAndValidate(
		ui, "CIDR condition remote IP (optional):", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return utils.GetExecStatus(err)
	} else {
		params[cst.DataCidr] = cidr
	}

	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}

	policy, e := createPolicy(params)
	if e != nil {
		p.outClient.Fail(e)
		return utils.GetExecStatus(err)
	}
	resp, err = p.submitPolicy(policy)
	p.outClient.WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func (p Policy) handlePolicyUpsertCmd(args []string) int {
	if OnlyGlobalArgs(args) {
		return p.handlePolicyUpsertWorkflow(args)
	}
	params := map[string]string{}
	var resp []byte
	var err *errors.ApiError

	path, status := getPolicyParams(args)
	if status != 0 {
		return status
	}
	params[cst.Path] = path

	data := viper.GetString(cst.Data)
	if data == "" {
		params[cst.DataAction] = viper.GetString(cst.DataAction)
		params[cst.DataSubject] = viper.GetString(cst.DataSubject)
		params[cst.DataCidr] = viper.GetString(cst.DataCidr)
		params[cst.DataDescription] = viper.GetString(cst.DataDescription)
		params[cst.DataResource] = viper.GetString(cst.DataResource)
		effect := viper.GetString(cst.DataEffect)
		if effect == "" {
			effect = "allow"
		}
		params[cst.DataEffect] = effect
		err = ValidateParams(params, []string{cst.DataAction, cst.DataSubject, cst.DataEffect, cst.Path})
	}
	if p.outClient == nil {
		p.outClient = format.NewDefaultOutClient()
	}
	if err == nil {
		encoding := viper.GetString(cst.Encoding)
		var postData *postPolicyModel
		if data != "" {
			postData = &postPolicyModel{
				Policy:        data,
				Serialization: encoding,
				Path:          params[cst.Path],
			}
		} else {
			data, e := createPolicy(params)
			if e != nil {
				p.outClient.Fail(e)
				return utils.GetExecStatus(err)
			}
			postData = data
		}
		resp, err = p.submitPolicy(postData)
		p.outClient.WriteResponse(resp, err)
		return utils.GetExecStatus(err)
	}
	p.outClient.WriteResponse(nil, err)
	return utils.GetExecStatus(err)
}

func (p Policy) handlePolicySearchCommand(args []string) int {
	baseType := strings.Join([]string{cst.Config, cst.NounPolicies}, "/")
	data, err := handleSearch(args, baseType, p.request)
	outClient := p.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func getPolicyParams(args []string) (path string, status int) {
	status = 0
	path = viper.GetString(cst.Path)
	if path == "" && len(args) > 0 {
		path = args[0]
	}
	if path == "" {
		status = cli.RunResultHelp
	}
	return path, status
}

type postPolicyModel struct {
	Path          string `json:"path"`
	Policy        string `json:"policy"`
	Serialization string `json:"serialization"`
}

type CIDRCondition struct {
	CIDR string `json:"cidr"`
}

type jsonCondition struct {
	Type    string          `json:"type"`
	Options json.RawMessage `json:"options"`
}

type document struct {
	PermissionDocument []*defaultPolicy `json:"permissionDocument"`
	TenantName         string           `json:"tenantName"`
}

type defaultPolicy struct {
	ID          string                   `json:"id" gorethink:"id"`
	Description string                   `json:"description" gorethink:"description"`
	Subjects    []string                 `json:"subjects" gorethink:"subjects"`
	Effect      string                   `json:"effect" gorethink:"effect"`
	Resources   []string                 `json:"resources" gorethink:"resources"`
	Actions     []string                 `json:"actions" gorethink:"actions"`
	Conditions  map[string]jsonCondition `json:"conditions" gorethink:"conditions"`
	//Meta        []byte     `json:"meta" gorethink:"meta"`
}

func setCidrCondition(policy *defaultPolicy, cidr string) *errors.ApiError {
	if _, _, err := net.ParseCIDR(cidr); err != nil {
		return errors.New(err).Grow("Invalid cidr")
	}
	cidrCondition := CIDRCondition{CIDR: cidr}
	if raw, err := json.Marshal(cidrCondition); err != nil {
		return errors.New(err).Grow("Failed to serialized cidr condition")
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
