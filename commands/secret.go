package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"thy/auth"
	cst "thy/constants"
	"thy/errors"
	"thy/internal/predictor"
	"thy/paths"
	"thy/store"
	"thy/utils"
	"thy/vaultcli"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

func GetDataOpWrappers(targetEntity string) []*predictor.Params {
	return []*predictor.Params{
		{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("%s to be stored in a %s. Prefix with '@' to denote filepath (required)", strings.Title(cst.Data), targetEntity), Predictor: predictor.NewPrefixFilePredictor("*")},
		{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, targetEntity), Predictor: predictor.NewSecretPathPredictorDefault()},
		{Name: cst.DataDescription, Usage: fmt.Sprintf("Description of a %s", targetEntity)},
		{Name: cst.DataAttributes, Usage: fmt.Sprintf("Attributes of a %s", targetEntity)},
		{Name: cst.Overwrite, Usage: fmt.Sprintf("Overwrite all the contents of %s data", targetEntity), ValueType: "bool"},
	}
}
func GetNoDataOpWrappers(targetEntity string) []*predictor.Params {
	return []*predictor.Params{
		{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, targetEntity), Predictor: predictor.NewSecretPathPredictorDefault()},
		{Name: cst.ID, Shorthand: "i", Usage: fmt.Sprintf("Target %s for a %s", cst.ID, targetEntity)},
		{Name: cst.Version, Usage: "List the current and last (n) versions"},
	}
}

func GetSearchOpWrappers() []*predictor.Params {
	return []*predictor.Params{
		{Name: cst.Query, Shorthand: "q", Usage: fmt.Sprintf("%s of %ss to fetch (optional)", strings.Title(cst.Query), cst.NounSecret)},
		{Name: cst.SearchLinks, Usage: "Find secrets that link to the secret path in the query", ValueType: "bool"},
		{Name: cst.Limit, Shorthand: "l", Usage: "Maximum number of results per cursor (optional)"},
		{Name: cst.Cursor, Usage: cst.CursorHelpMessage},
		{Name: cst.SearchField, Usage: "Advanced search on a secret field (optional)"},
		{Name: cst.SearchType, Usage: "Specify the value type for advanced field searching, can be 'number' or 'string' (optional)"},
		{Name: cst.SearchComparison, Usage: "Specify the operator for advanced field searching, can be 'contains' or 'equal' (optional)"},
		{Name: cst.Sort, Usage: "Change result sorting order (asc|desc) [default: desc] when search field is specified (optional)"},
	}
}

func GetSecretCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSecret},
		SynopsisText: "secret (<path> | --path|-r)",
		HelpText: fmt.Sprintf(`Execute an action on a %s from %s

Usage:
   • secret %[3]s -b
   • secret --path %[3]s -b
`, cst.NounSecret, cst.ProductName, cst.ExamplePath),
		FlagsPredictor: GetNoDataOpWrappers(cst.NounSecret),
		MinNumberArgs:  1,
		RunFunc: func(args []string) int {
			id := viper.GetString(cst.ID)
			path := viper.GetString(cst.Path)
			if path == "" && id == "" && len(args) > 0 {
				path = args[0]
			}
			if path == "" && id == "" {
				return cli.RunResultHelp
			}
			return handleSecretReadCmd(vaultcli.New(), cst.NounSecret, args)
		},
	})
}

func GetReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSecret, cst.Read},
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounSecret, cst.Read),
		HelpText: fmt.Sprintf(`Read a %[2]s from %[3]s

Usage:
   • secret %[1]s %[4]s -b
   • secret %[1]s --path %[4]s -bf data.Data.Key
   • secret %[1]s --version
`, cst.Read, cst.NounSecret, cst.ProductName, cst.ExamplePath),
		FlagsPredictor:    GetNoDataOpWrappers(cst.NounSecret),
		ArgsPredictorFunc: predictor.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
		RunFunc: func(args []string) int {
			return handleSecretReadCmd(vaultcli.New(), cst.NounSecret, args)
		},
	})
}

func GetDescribeCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSecret, cst.Describe},
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounSecret, cst.Describe),
		HelpText: fmt.Sprintf(`Describe a %[2]s from %[3]s

Usage:
   • secret %[1]s %[4]s
   • secret %[1]s --path %[4]s -f id
`, cst.Describe, cst.NounSecret, cst.ProductName, cst.ExamplePath),
		FlagsPredictor:    GetNoDataOpWrappers(cst.NounSecret),
		ArgsPredictorFunc: predictor.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
		RunFunc: func(args []string) int {
			return handleSecretDescribeCmd(vaultcli.New(), cst.NounSecret, args)
		},
	})
}

func GetDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSecret, cst.Delete},
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounSecret, cst.Delete),
		HelpText: fmt.Sprintf(`Delete a %[2]s from %[3]s

Usage:
   • secret %[1]s %[4]s
   • secret %[1]s --path %[4]s --force
`, cst.Delete, cst.NounSecret, cst.ProductName, cst.ExamplePath),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounSecret), Predictor: predictor.NewSecretPathPredictorDefault()},
			{Name: cst.ID, Shorthand: "i", Usage: fmt.Sprintf("Target %s for a %s", cst.ID, cst.NounSecret)},
			{Name: cst.Force, Usage: fmt.Sprintf("Immediately delete %s", cst.NounSecret), ValueType: "bool"},
		},
		ArgsPredictorFunc: predictor.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
		RunFunc: func(args []string) int {
			return handleSecretDeleteCmd(vaultcli.New(), cst.NounSecret, args)
		},
	})
}

func GetSecretRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSecret, cst.Read},
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounSecret, cst.Restore),
		HelpText: fmt.Sprintf(`Restore a deleted %[2]s from %[3]s

Usage:
   • secret %[1]s %[4]s
`, cst.Restore, cst.NounSecret, cst.ProductName, cst.ExamplePath),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounSecret), Predictor: predictor.NewSecretPathPredictorDefault()},
			{Name: cst.ID, Shorthand: "i", Usage: fmt.Sprintf("Target %s for a %s", cst.ID, cst.NounSecret)},
		},
		ArgsPredictorFunc: predictor.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
		RunFunc: func(args []string) int {
			return handleSecretRestoreCmd(vaultcli.New(), cst.NounSecret, args)
		},
	})
}

func GetUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSecret, cst.Update},
		SynopsisText: fmt.Sprintf("%s %s (<path> <data> | (--path|-r) (--data|-d))", cst.NounSecret, cst.Update),
		HelpText: fmt.Sprintf(`Update a %[2]s in %[3]s

Usage:
   • secret %[1]s %[4]s %[5]s
   • secret %[1]s --path %[4]s --data %[5]s
   • secret %[1]s --path %[4]s --data %[6]s
`, cst.Update, cst.NounSecret, cst.ProductName, cst.ExamplePath, cst.ExampleDataJSON, cst.ExampleDataPath),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("%s to be stored in a %s. Prefix with '@' to denote filepath (required)", strings.Title(cst.Data), cst.NounSecret), Predictor: predictor.NewPrefixFilePredictor("*")},
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounSecret), Predictor: predictor.NewSecretPathPredictorDefault()},
			{Name: cst.DataDescription, Usage: fmt.Sprintf("Description of a %s", cst.NounSecret)},
			{Name: cst.DataAttributes, Usage: fmt.Sprintf("Attributes of a %s", cst.NounSecret)},
			{Name: cst.Overwrite, Usage: fmt.Sprintf("Overwrite all the contents of %s data", cst.NounSecret), ValueType: "bool"},
			{Name: cst.ID, Shorthand: "i", Usage: fmt.Sprintf("Target %s for a %s", cst.ID, cst.NounSecret)},
		},
		ArgsPredictorFunc: predictor.NewSecretPathPredictorDefault().Predict,
		RunFunc: func(args []string) int {
			return handleSecretUpsertCmd(vaultcli.New(), cst.NounSecret, args)
		},
	})
}

func GetRollbackCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSecret, cst.Rollback},
		SynopsisText: fmt.Sprintf("%s %s (<path> | (--path|-r))", cst.NounSecret, cst.Rollback),
		HelpText: fmt.Sprintf(`Rollback a %[2]s in %[3]s

Usage:
   • secret %[1]s %[4]s --%[5]s 1
   • secret %[1]s --path %[4]s
`, cst.Rollback, cst.NounSecret, cst.ProductName, cst.ExamplePath, cst.Version),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounSecret), Predictor: predictor.NewSecretPathPredictorDefault()},
			{Name: cst.ID, Shorthand: "i", Usage: fmt.Sprintf("Target %s for a %s", cst.ID, cst.NounSecret)},
			{Name: cst.Version, Usage: "The version to which to rollback"},
		},
		ArgsPredictorFunc: predictor.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
		RunFunc: func(args []string) int {
			return handleSecretRollbackCmd(vaultcli.New(), cst.NounSecret, args)
		},
	})
}

func GetEditCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSecret, cst.Update},
		SynopsisText: fmt.Sprintf("%s %s (<path> <data> | (--path|-r))", cst.NounSecret, cst.Edit),
		HelpText: fmt.Sprintf(`Edit a %[2]s in %[3]s

Usage:
   • secret %[1]s %[4]s
   • secret %[1]s --path %[4]s
`, cst.Edit, cst.NounSecret, cst.ProductName, cst.ExamplePath),
		FlagsPredictor:    GetNoDataOpWrappers(cst.NounSecret),
		ArgsPredictorFunc: predictor.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
		RunFunc: func(args []string) int {
			return handleSecretEditCmd(vaultcli.New(), cst.NounSecret, args)
		},
	})
}

func GetCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSecret, cst.Create},
		SynopsisText: fmt.Sprintf("%s %s (<path> <data> | (--path|-r) (--data|-d))", cst.NounSecret, cst.Create),
		HelpText: fmt.Sprintf(`Create a %[2]s in %[3]s

Usage:
   • secret %[1]s %[4]s --data %[5]s
   • secret %[1]s --path %[4]s --data %[5]s
   • secret %[1]s --path %[4]s --data %[6]s
`, cst.Create, cst.NounSecret, cst.ProductName, cst.ExamplePath, cst.ExampleDataJSON, cst.ExampleDataPath),
		FlagsPredictor: GetDataOpWrappers(cst.NounSecret),
		RunFunc: func(args []string) int {
			return handleSecretUpsertCmd(vaultcli.New(), cst.NounSecret, args)
		},
	})
}

func GetBustCacheCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSecret, cst.BustCache},
		SynopsisText: fmt.Sprintf("%s %s", cst.NounSecret, cst.BustCache),
		HelpText: `Bust secret cache

Usage:
   • secret bustcache
`,
		RunFunc: func(args []string) int {
			return handleBustCacheCmd(vaultcli.New(), args)
		},
	})
}

func GetSecretSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounSecret, cst.Search},
		SynopsisText: fmt.Sprintf("%s (<query> | --query) --limit[default:25] --cursor --search-type[default:string] --search-comparison[default:contains] --search-field[default:path] --search-links[default:false])", cst.Search),
		HelpText: fmt.Sprintf(`Search for a %[2]s from %[3]s

Usage:
    • secret %[1]s %[4]s
    • secret %[1]s --query %[4]s
    • secret %[1]s --query aws:base:secret --search-links
    • secret %[1]s --query aws --search-field attributes.type
    • secret %[1]s --query 900 --search-field attributes.ttl --search-type number
    • secret %[1]s --query production --search-field attributes.stage --search-comparison equal
`, cst.Search, cst.NounSecret, cst.ProductName, cst.ExampleUserSearch),
		FlagsPredictor: GetSearchOpWrappers(),
		RunFunc: func(args []string) int {
			return handleSecretSearchCmd(vaultcli.New(), cst.NounSecret, args)
		},
	})
}

func handleBustCacheCmd(vcli vaultcli.CLI, args []string) int {
	var err *errors.ApiError
	var s store.Store
	st := viper.GetString(cst.StoreType)
	if s, err = vcli.Store(st); err == nil {
		err = s.Wipe(cst.SecretRoot)
		err = s.Wipe(cst.SecretDescriptionRoot).And(err)
	}
	if err == nil {
		log.Print("Successfully cleared local cache")
	}

	vcli.Out().WriteResponse(nil, err)
	return 0
}

func handleSecretDescribeCmd(vcli vaultcli.CLI, secretType string, args []string) int {
	id := viper.GetString(cst.ID)
	path := viper.GetString(cst.Path)
	if path == "" {
		path = viper.GetString(cst.ID)
		id = ""
	}
	if path == "" && len(args) > 0 {
		path = args[0]
	}
	resp, err := getSecret(vcli, secretType, path, id, false, cst.SuffixDescription)

	vcli.Out().WriteResponse(resp, err)
	return 0
}

func handleSecretReadCmd(vcli vaultcli.CLI, secretType string, args []string) int {
	id := viper.GetString(cst.ID)
	path := viper.GetString(cst.Path)
	if path == "" {
		path = viper.GetString(cst.ID)
		id = ""
	}
	if path == "" && len(args) > 0 {
		path = args[0]
	}
	version := viper.GetString(cst.Version)
	if strings.TrimSpace(version) != "" {
		path = fmt.Sprint(path, "/", cst.Version, "/", version)
	}
	resp, err := getSecret(vcli, secretType, path, id, false, "")

	vcli.Out().WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func handleSecretRestoreCmd(vcli vaultcli.CLI, secretType string, args []string) int {
	var err *errors.ApiError
	var data []byte
	path := viper.GetString(cst.Path)
	if path == "" {
		path = viper.GetString(cst.ID)
	}
	if path == "" && len(args) > 0 {
		path = args[0]
	}

	rc, rerr := getResourceConfig(path, secretType)
	if rerr != nil {
		vcli.Out().Fail(rerr)
		return utils.GetExecStatus(rerr)
	}
	path = rc.path
	uri := paths.CreateResourceURI(rc.resourceType, path, "/restore", true, nil)
	data, err = vcli.HTTPClient().DoRequest(http.MethodPut, uri, nil)

	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func handleSecretSearchCmd(vcli vaultcli.CLI, secretType string, args []string) int {
	query := viper.GetString(cst.Query)
	limit := viper.GetString(cst.Limit)
	cursor := viper.GetString(cst.Cursor)
	searchLinks := viper.GetBool(cst.SearchLinks)
	searchType := viper.GetString(cst.SearchType)
	searchComparison := viper.GetString(cst.SearchComparison)
	searchField := viper.GetString(cst.SearchField)
	sort := viper.GetString(cst.Sort)
	if query == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		query = args[0]
	}

	queryParams := map[string]string{
		cst.SearchKey:        query,
		cst.Limit:            limit,
		cst.Cursor:           cursor,
		cst.SearchType:       searchType,
		cst.SearchComparison: searchComparison,
		cst.SearchField:      searchField,
		cst.Sort:             sort,
	}
	if searchLinks {
		//flag just needs to be present
		queryParams[cst.SearchLinks] = strconv.FormatBool(searchLinks)
	}
	rc, rerr := getResourceConfig("", secretType)
	if rerr != nil {
		vcli.Out().Fail(rerr)
		return utils.GetExecStatus(rerr)
	}
	uri := paths.CreateResourceURI(rc.resourceType, "", "", false, queryParams)
	data, err := vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)

	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func handleSecretDeleteCmd(vcli vaultcli.CLI, secretType string, args []string) int {
	var err *errors.ApiError
	var resp []byte
	id := viper.GetString(cst.ID)
	path := viper.GetString(cst.Path)
	force := viper.GetBool(cst.Force)
	if path == "" {
		path = viper.GetString(cst.ID)
		id = ""
	}
	if path == "" && len(args) > 0 {
		path = args[0]
	}

	query := map[string]string{"force": strconv.FormatBool(force)}
	rc, rerr := getResourceConfig(path, secretType)
	if rerr != nil {
		vcli.Out().Fail(err)
		return utils.GetExecStatus(err)
	}
	path = rc.path
	uri, err := paths.GetResourceURIFromResourcePath(rc.resourceType, path, id, "", query)
	resp, err = vcli.HTTPClient().DoRequest(http.MethodDelete, uri, nil)

	vcli.Out().WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func handleSecretRollbackCmd(vcli vaultcli.CLI, secretType string, args []string) int {
	var apiError *errors.ApiError
	var resp []byte

	path := viper.GetString(cst.Path)
	if path == "" {
		path = viper.GetString(cst.ID)
	}
	if path == "" && len(args) > 0 {
		path = args[0]
	}
	version := viper.GetString(cst.Version)
	rc, err := getResourceConfig(path, secretType)
	if err != nil {
		vcli.Out().Fail(err)
		return utils.GetExecStatus(err)
	}

	path = rc.path

	// If version is not provided, get the secret's description and parse the version from it.
	// Submit a request for a version that's previous relative to the one found.
	if version == "" {
		id := viper.GetString(cst.ID)
		resp, apiError = getSecret(vcli, secretType, path, id, false, cst.SuffixDescription)
		if apiError != nil {
			vcli.Out().WriteResponse(resp, apiError)
			return utils.GetExecStatus(apiError)
		}

		v, err := utils.GetPreviousVersion(resp)
		if err != nil {
			vcli.Out().Fail(err)
			return 1
		}
		version = v
	}
	if strings.TrimSpace(version) != "" {
		path = fmt.Sprint(path, "/rollback/", version)
	}
	uri := paths.CreateResourceURI(rc.resourceType, path, "", true, nil)
	resp, apiError = vcli.HTTPClient().DoRequest(http.MethodPut, uri, nil)

	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func handleSecretUpsertCmd(vcli vaultcli.CLI, secretType string, args []string) int {
	action := strings.ToLower(viper.GetString(cst.LastCommandKey))

	if OnlyGlobalArgs(args) {
		var (
			resp []byte
			err  *errors.ApiError
		)
		switch action {
		case cst.Create:
			resp, err = handleCreateWorkflow(vcli, secretType)
		case cst.Update:
			resp, err = handleUpdateWorkflow(vcli, secretType)
		}
		vcli.Out().WriteResponse(resp, err)
		return utils.GetExecStatus(err)
	}

	id := viper.GetString(cst.ID)
	path := viper.GetString(cst.Path)
	overwrite := viper.GetBool(cst.Overwrite)
	data := viper.GetString(cst.Data)
	desc := viper.GetString(cst.DataDescription)
	attributes := viper.GetStringMap(cst.DataAttributes)

	if path == "" {
		path = id
		id = ""
	}
	if path == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		path = args[0]
	}

	rc, rerr := getResourceConfig(path, secretType)
	if rerr != nil {
		vcli.Out().Fail(rerr)
		return utils.GetExecStatus(rerr)
	}
	path = rc.path
	uri, err := paths.GetResourceURIFromResourcePath(rc.resourceType, path, id, "", nil)
	if err != nil {
		vcli.Out().FailE(err)
		return utils.GetExecStatus(err)
	}

	if data == "" && desc == "" && len(attributes) == 0 {
		vcli.Out().FailF("Please provide a properly formed value for at least --%s, or --%s, or --%s.",
			cst.Data, cst.DataDescription, cst.DataAttributes)
		return 1
	}

	dataMap := make(map[string]interface{})
	if data != "" {
		parseErr := json.Unmarshal([]byte(data), &dataMap)
		if parseErr != nil {
			vcli.Out().FailF("Failed to parse passed in secret data: %v", parseErr)
			return 1
		}
	}
	postData := secretUpsertBody{
		Data:        dataMap,
		Description: desc,
		Attributes:  attributes,
		Overwrite:   overwrite,
	}

	var reqMethod string
	if action == cst.Create {
		reqMethod = http.MethodPost
	} else {
		reqMethod = http.MethodPut
	}
	resp, err := vcli.HTTPClient().DoRequest(reqMethod, uri, &postData)

	vcli.Out().WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func handleCreateWorkflow(vcli vaultcli.CLI, secretType string) ([]byte, *errors.ApiError) {
	dataMap := make(map[string]interface{})
	attrMap := make(map[string]interface{})

	qs := []*survey.Question{
		{
			Name:   "Path",
			Prompt: &survey.Input{Message: "Path:"},
			Validate: func(ans interface{}) error {
				answer := strings.TrimSpace(ans.(string))
				if len(answer) == 0 {
					return errors.NewS("Value is required")
				}
				return nil
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				return strings.TrimSpace(ans.(string))
			},
		},
		{
			Name:   "Description",
			Prompt: &survey.Input{Message: "Description:"},
			Transform: func(ans interface{}) (newAns interface{}) {
				return strings.TrimSpace(ans.(string))
			},
		},
	}
	answers := struct {
		Path        string
		Description string
	}{}
	survErr := survey.Ask(qs, &answers)
	if survErr != nil {
		return nil, errors.New(survErr)
	}

	var actionID int
	actionPrompt := &survey.Select{
		Message: "Add Attributes?",
		Options: []string{"Skip", "Add key/value pairs", "Define as a json string"},
	}
	survErr = survey.AskOne(actionPrompt, &actionID)
	if survErr != nil {
		return nil, errors.New(survErr)
	}

	var wizErr *errors.ApiError
	switch actionID {
	case 1:
		attrMap, wizErr = handleKeyValueWizard()
	case 2:
		attrMap, wizErr = handleJSONWizard("Attributes:")
	}
	if wizErr != nil {
		return nil, wizErr
	}

	actionPrompt = &survey.Select{
		Message: "Add Data?",
		Options: []string{"Skip", "Add key/value pairs", "Define as a json string"},
	}
	survErr = survey.AskOne(actionPrompt, &actionID)
	if survErr != nil {
		return nil, errors.New(survErr)
	}

	switch actionID {
	case 1:
		dataMap, wizErr = handleKeyValueWizard()
	case 2:
		dataMap, wizErr = handleJSONWizard("Data:")
	}
	if wizErr != nil {
		return nil, wizErr
	}

	rc, err := getResourceConfig(answers.Path, secretType)
	if err != nil {
		return nil, errors.New(err)
	}

	postData := secretUpsertBody{
		Description: answers.Description,
		Data:        dataMap,
		Attributes:  attrMap,
	}

	uri, apiErr := paths.GetResourceURIFromResourcePath(rc.resourceType, answers.Path, "", "", nil)
	if apiErr != nil {
		return nil, apiErr
	}

	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, &postData)
}

func handleUpdateWorkflow(vcli vaultcli.CLI, secretType string) ([]byte, *errors.ApiError) {
	var path string
	pathPrompt := &survey.Input{Message: "Path:"}
	pathValidation := func(ans interface{}) error {
		answer := strings.TrimSpace(ans.(string))
		if len(answer) == 0 {
			return errors.NewS("Value is required")
		}
		return nil
	}
	survErr := survey.AskOne(pathPrompt, &path, survey.WithValidator(pathValidation))
	if survErr != nil {
		return nil, errors.New(survErr)
	}
	path = strings.TrimSpace(path)

	rc, rerr := getResourceConfig(path, secretType)
	if rerr != nil {
		return nil, errors.New(rerr)
	}
	path = rc.path
	uri, err := paths.GetResourceURIFromResourcePath(rc.resourceType, path, "", "", nil)
	if err != nil {
		return nil, err
	}

	isSecretRetrieved := false
	secretResp := &secretGetResponse{}

	secretBytes, err := vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
	if err != nil {
		httpResp := err.HttpResponse()
		if httpResp == nil {
			return nil, err
		}
		switch httpResp.StatusCode {
		case http.StatusNotFound:
			return nil, errors.NewS("Secret under that path cannot be found.")
		case http.StatusForbidden:
			var yes bool
			continuePrompt := &survey.Confirm{
				Message: "You are not allowed to read secret under that path. Do you want to continue?",
				Default: true,
			}
			survErr = survey.AskOne(continuePrompt, &yes)
			if survErr != nil {
				return nil, errors.New(survErr)
			}
			if !yes {
				return nil, nil
			}
		default:
			log.Printf("Get secret failed. %s: %v", httpResp.Status, err)
		}
	} else {
		rerr = json.Unmarshal(secretBytes, secretResp)
		if rerr != nil {
			log.Printf("error: cannot read Secret returned from API: %v", rerr)
		} else {
			isSecretRetrieved = true
		}
	}

	if isSecretRetrieved {
		if secretResp.Description == "" {
			vcli.Out().WriteResponse([]byte("Currently description is empty."), nil)
		} else {
			vcli.Out().WriteResponse([]byte(fmt.Sprintf("Current description: %q\n", secretResp.Description)), nil)
		}
	}
	var doUpdDescription bool
	doUpdDescriptionPrompt := &survey.Confirm{Message: "Update description?", Default: false}
	survErr = survey.AskOne(doUpdDescriptionPrompt, &doUpdDescription)
	if survErr != nil {
		return nil, errors.New(survErr)
	}

	var desc string
	if doUpdDescription {
		descriptionPrompt := &survey.Input{Message: "Description:"}
		survErr = survey.AskOne(descriptionPrompt, &desc)
		if survErr != nil {
			return nil, errors.New(survErr)
		}
		desc = strings.TrimSpace(desc)
	}

	if isSecretRetrieved {
		vcli.Out().WriteResponse([]byte("Attributes and data defined currently for the secret:"), nil)
		// Print attributes and data beautifully :)
		relativeData, err := json.Marshal(map[string]interface{}{
			"attributes": secretResp.Attributes,
			"data":       secretResp.Data,
		})
		if err == nil {
			vcli.Out().WriteResponse(relativeData, nil)
		}
	}

	var overwrite bool
	overwritePrompt := &survey.Confirm{Message: "Overwrite existing attributes and data?", Default: false}
	survErr = survey.AskOne(overwritePrompt, &overwrite)
	if survErr != nil {
		return nil, errors.New(survErr)
	}

	dataMap := make(map[string]interface{})
	attrMap := make(map[string]interface{})

	var q string
	if overwrite {
		q = "Overwrite attributes?"
	} else {
		q = "Update attributes?"
	}

	var attrActionID int
	actionPrompt := &survey.Select{
		Message: q,
		Options: []string{"No, skip", "Yes, use key/value pairs", "Yes, define as a json string"},
	}
	survErr = survey.AskOne(actionPrompt, &attrActionID)
	if survErr != nil {
		return nil, errors.New(survErr)
	}

	var wizErr *errors.ApiError

	switch attrActionID {
	case 1:
		attrMap, wizErr = handleKeyValueWizard()
	case 2:
		attrMap, wizErr = handleJSONWizard("Attributes:")
	}
	if wizErr != nil {
		return nil, wizErr
	}

	if overwrite {
		q = "Overwrite data?"
	} else {
		q = "Update data?"
	}
	var dataActionID int
	actionPrompt = &survey.Select{
		Message: q,
		Options: []string{"No, skip", "Yes, use key/value pairs", "Yes, define as a json string"},
	}
	survErr = survey.AskOne(actionPrompt, &dataActionID)
	if survErr != nil {
		return nil, errors.New(survErr)
	}
	switch dataActionID {
	case 1:
		dataMap, wizErr = handleKeyValueWizard()
	case 2:
		dataMap, wizErr = handleJSONWizard("Data:")
	}
	if wizErr != nil {
		return nil, wizErr
	}

	if !doUpdDescription && attrActionID == 0 && dataActionID == 0 {
		vcli.Out().WriteResponse([]byte("Nothing to update. Exiting."), nil)
		return nil, nil
	}

	postData := secretUpsertBody{
		Description: desc,
		Data:        dataMap,
		Attributes:  attrMap,
		Overwrite:   overwrite,
	}

	vcli.Out().WriteResponse([]byte("Sending request..."), nil)
	return vcli.HTTPClient().DoRequest(http.MethodPut, uri, &postData)
}

func handleSecretEditCmd(vcli vaultcli.CLI, secretType string, args []string) int {
	var resp []byte
	id := viper.GetString(cst.ID)
	path := viper.GetString(cst.Path)
	if path == "" {
		path = viper.GetString(cst.ID)
		id = ""
	}
	if path == "" && len(args) > 0 {
		path = args[0]
	}
	rc, rerr := getResourceConfig(path, secretType)
	if rerr != nil {
		vcli.Out().Fail(rerr)
		return utils.GetExecStatus(rerr)
	}
	path = rc.path
	uri, err := paths.GetResourceURIFromResourcePath(rc.resourceType, path, id, "", nil)

	// fetch
	resp, err = getSecret(vcli, secretType, path, id, true, "")
	if err != nil {
		vcli.Out().WriteResponse(resp, err)
		return utils.GetExecStatus(err)
	}

	saveFunc := func(data []byte) (resp []byte, err *errors.ApiError) {
		var model secretUpsertBody
		if mErr := json.Unmarshal(data, &model); mErr != nil {
			return nil, errors.New(mErr).Grow("invalid format for secret")
		}
		model.Overwrite = true
		_, err = vcli.HTTPClient().DoRequest(http.MethodPut, uri, &model)
		return nil, err
	}
	resp, err = vcli.Edit(resp, saveFunc)
	vcli.Out().WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func handleKeyValueWizard() (map[string]interface{}, *errors.ApiError) {
	data := make(map[string]interface{})
	for {
		qs := []*survey.Question{
			{
				Name:   "Key",
				Prompt: &survey.Input{Message: "Key:"},
				Validate: func(ans interface{}) error {
					answer := strings.TrimSpace(ans.(string))
					if len(answer) == 0 {
						return errors.NewS("Value is required")
					}
					return nil
				},
				Transform: func(ans interface{}) (newAns interface{}) {
					return strings.TrimSpace(ans.(string))
				},
			},
			{
				Name:   "Value",
				Prompt: &survey.Input{Message: "Value:"},
				Validate: func(ans interface{}) error {
					answer := strings.TrimSpace(ans.(string))
					if len(answer) == 0 {
						return errors.NewS("Value is required")
					}
					return nil
				},
				Transform: func(ans interface{}) (newAns interface{}) {
					return strings.TrimSpace(ans.(string))
				},
			},
			{
				Name:   "More",
				Prompt: &survey.Confirm{Message: "Add more?", Default: false},
			},
		}
		answers := struct {
			Key   string
			Value string
			More  bool
		}{}
		survErr := survey.Ask(qs, &answers)
		if survErr != nil {
			return nil, errors.New(survErr)
		}

		data[answers.Key] = answers.Value

		if !answers.More {
			break
		}
	}
	return data, nil
}

func handleJSONWizard(msg string) (map[string]interface{}, *errors.ApiError) {
	var result string
	validation := func(ans interface{}) error {
		answer := strings.TrimSpace(ans.(string))
		if len(answer) == 0 {
			return nil
		}
		mock := map[string]interface{}{}
		err := json.Unmarshal([]byte(answer), &mock)
		if err != nil {
			return errors.NewF("Invalid JSON: %v", err)
		}
		return nil
	}
	survErr := survey.AskOne(&survey.Input{Message: msg}, &result, survey.WithValidator(validation))
	if survErr != nil {
		return nil, errors.New(survErr)
	}

	data := make(map[string]interface{})
	if len(result) == 0 {
		return data, nil
	}
	err := json.Unmarshal([]byte(result), &data)
	if err != nil {
		// Should not happen because validation above would not pass invalid JSON.
		return nil, errors.NewF("Failed to unmarshal input: %v", err)
	}
	return data, nil
}

func getSecret(vcli vaultcli.CLI, secretType string, path string, id string, edit bool, requestSuffix string) (respData []byte, err *errors.ApiError) {
	var secretDistinguisher string
	var secretKey string
	cacheRoot := cst.SecretRoot
	if requestSuffix == cst.SuffixDescription {
		cacheRoot = cst.SecretDescriptionRoot
	}
	if path != "" {
		secretDistinguisher = path
	} else {
		secretDistinguisher = id
	}
	secretDistinguisher = strings.ReplaceAll(secretDistinguisher, ":", "/")
	secretDistinguisher = strings.ReplaceAll(secretDistinguisher, "-", "<>")
	secretKey = cacheRoot + "-" + secretDistinguisher

	cacheStrategy := store.CacheStrategy(viper.GetString(cst.CacheStrategy))
	var cacheData []byte
	var expired bool
	var s store.Store
	st := viper.GetString(cst.StoreType)
	if cacheStrategy == store.CacheThenServer || cacheStrategy == store.CacheThenServerThenExpired {
		cacheData, expired = getSecretDataFromCache(vcli, secretKey, st)
		if !expired && len(cacheData) > 0 {
			return cacheData, nil
		}
	}

	queryTerms := map[string]string{"edit": strconv.FormatBool(edit)}

	rc, rerr := getResourceConfig(path, secretType)
	if rerr != nil {
		return nil, errors.New(rerr)
	}
	path = rc.path
	uri, err := paths.GetResourceURIFromResourcePath(rc.resourceType, path, id, requestSuffix, queryTerms)
	if err != nil {
		return nil, err
	}
	respData, err = vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
	if err == nil {
		if cacheStrategy != store.Never {
			if s == nil {
				var errStore *errors.ApiError
				s, errStore = vcli.Store(st)
				if errStore != nil {
					log.Printf("Failed to get store to cache secret for store type %s. Error: %s", st, errStore)
				}
			}
			if s != nil {
				if errStore := s.Store(secretKey, secretData{
					Date: time.Now().UTC(),
					Data: respData,
				}); errStore != nil {
					log.Printf("Failed to cache secret for store type %s. Error: %s", st, errStore)
				}
			}
		}
		return respData, nil
	} else if cacheStrategy == store.ServerThenCache {
		cacheData, expired = getSecretDataFromCache(vcli, secretKey, st)
		if !expired {
			return cacheData, nil
		}
		//TODO: is this ever  execute???
	} else if cacheStrategy == store.CacheThenServerThenExpired && len(cacheData) > 0 {
		log.Print("Cache expired but failed to retrieve from server so returning cached data")
		return cacheData, nil
	}

	return nil, err.Or(errors.NewS("run in verbose mode for more information"))
}

func getSecretDataFromCache(vcli vaultcli.CLI, secretKey string, st string) (cacheData []byte, expired bool) {
	if s, err := vcli.Store(st); err != nil {
		log.Printf("Failed to get store of type %s. Error: %s", st, err.Error())
	} else {
		var data secretData
		if err := s.Get(secretKey, &data); err != nil && len(data.Data) > 0 {
			log.Printf("Failed to fetch cached secret from store type %s. Error: %s", st, err.Error())
		} else {
			cacheData = data.Data
			cacheAgeMinutes := viper.GetInt(cst.CacheAge)
			if cacheAgeMinutes > 0 {
				expired = (data.Date.Sub(time.Now().UTC()).Seconds() + float64(cacheAgeMinutes)*60) < 0
			} else {
				log.Printf("Invalid cache age: %d", cacheAgeMinutes)
			}
		}
	}

	return cacheData, expired
}

type secretData struct {
	Date time.Time
	Data []byte
}

// secretGetResponse contains only info that can be updated.
type secretGetResponse struct {
	Description string                 `json:"description"`
	Attributes  map[string]interface{} `json:"attributes"`
	Data        map[string]interface{} `json:"data"`
}

type secretUpsertBody struct {
	Data        map[string]interface{}
	Description string
	Attributes  map[string]interface{}
	Overwrite   bool
}

type resourceConfig struct {
	resourceType string
	path         string
}

func getResourceConfig(path, resourceType string) (*resourceConfig, error) {
	if resourceType == cst.NounHome {
		current, err := auth.GetCurrentIdentity()
		if err != nil {
			return nil, errors.NewS("error: unable to get current identity from access token")
		}
		rc := &resourceConfig{
			resourceType: fmt.Sprintf("%s/%s", cst.NounHome, current),
			path:         path,
		}
		if utils.CheckPrefix(path, "users:", "roles:") {
			p := strings.SplitAfterN(path, "/", 2)
			if len(p) == 2 {
				rc.resourceType = fmt.Sprintf("%s/%s", "home", p[0])
				rc.path = p[1]
			}
		}
		return rc, nil
	} else {
		rc := &resourceConfig{
			resourceType: cst.NounSecrets,
			path:         path,
		}
		return rc, nil
	}
}
