package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"thy/auth"
	cst "thy/constants"
	"thy/errors"
	"thy/format"
	"thy/internal/prompt"
	"thy/paths"
	preds "thy/predictors"
	"thy/requests"
	"thy/store"
	"thy/utils"

	"github.com/posener/complete"
	"github.com/spf13/viper"
	"github.com/thycotic-rd/cli"
)

type Secret struct {
	request    requests.Client
	outClient  format.OutClient
	getStore   func(stString string) (store.Store, *errors.ApiError)
	edit       func([]byte, dataFunc, *errors.ApiError, bool) ([]byte, *errors.ApiError)
	secretType paths.SecretType
}

func GetDataOpWrappers(targetEntity string) cli.PredictorWrappers {
	return cli.PredictorWrappers{
		preds.LongFlag(cst.Data):            cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("%s to be stored in a %s. Prefix with '@' to denote filepath (required)", strings.Title(cst.Data), targetEntity)}), false},
		preds.LongFlag(cst.Path):            cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, targetEntity)}), false},
		preds.LongFlag(cst.DataDescription): cli.PredictorWrapper{complete.PredictNothing, preds.NewFlagValue(preds.Params{Name: cst.DataDescription, Usage: fmt.Sprintf("Description of a %s", targetEntity)}), false},
		preds.LongFlag(cst.DataAttributes):  cli.PredictorWrapper{complete.PredictNothing, preds.NewFlagValue(preds.Params{Name: cst.DataAttributes, Usage: fmt.Sprintf("Attributes of a %s", targetEntity)}), false},
		preds.LongFlag(cst.Overwrite):       cli.PredictorWrapper{complete.PredictNothing, preds.NewFlagValue(preds.Params{Name: cst.Overwrite, Usage: fmt.Sprintf("Overwrite all the contents of %s data", targetEntity), Global: false, ValueType: "bool"}), false},
	}
}
func GetNoDataOpWrappers(targetEntity string) cli.PredictorWrappers {
	return cli.PredictorWrappers{
		preds.LongFlag(cst.Path):    cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, targetEntity)}), false},
		preds.LongFlag(cst.ID):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.ID, Shorthand: "i", Usage: fmt.Sprintf("Target %s for a %s", cst.ID, targetEntity)}), false},
		preds.LongFlag(cst.Version): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Version, Shorthand: "", Usage: "List the current and last (n) versions"}), false},
	}
}

func GetSearchOpWrappers() cli.PredictorWrappers {
	return cli.PredictorWrappers{
		preds.LongFlag(cst.Query):            cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Query, Shorthand: "q", Usage: fmt.Sprintf("%s of %ss to fetch (required)", strings.Title(cst.Query), cst.NounSecret)}), false},
		preds.LongFlag(cst.SearchLinks):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.SearchLinks, Shorthand: "", Usage: "Find secrets that link to the secret path in the query", Global: false, ValueType: "bool"}), false},
		preds.LongFlag(cst.Limit):            cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Limit, Shorthand: "l", Usage: fmt.Sprint("Maximum number of results per cursor (optional)")}), false},
		preds.LongFlag(cst.Cursor):           cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Cursor, Usage: cst.CursorHelpMessage}), false},
		preds.LongFlag(cst.SearchField):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.SearchField, Shorthand: "", Usage: "Advanced search on a secret field (optional)", Global: false}), false},
		preds.LongFlag(cst.SearchType):       cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.SearchType, Shorthand: "", Usage: "Specify the value type for advanced field searching, can be 'number' or 'string' (optional)", Global: false}), false},
		preds.LongFlag(cst.SearchComparison): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.SearchComparison, Shorthand: "", Usage: "Specify the operator for advanced field searching, can be 'contains' or 'equal' (optional)", Global: false}), false},
		preds.LongFlag(cst.Sort):             cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Sort, Usage: "Change result sorting order (asc|desc) [default: desc] when search field is specified (optional)"}), false},
	}
}

func GetSecretCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounSecret},
		RunFunc: func(args []string) int {
			id := viper.GetString(cst.ID)
			path := viper.GetString(cst.Path)
			if path == "" && id == "" {
				path = paths.GetPath(args)
			}
			if path == "" && id == "" {
				return cli.RunResultHelp
			}
			return Secret{
				requests.NewHttpClient(),
				nil,
				store.GetStore, nil, cst.NounSecret}.handleReadCmd(args)
		},
		SynopsisText: "secret (<path> | --path|-r)",
		HelpText: fmt.Sprintf(`Execute an action on a %s from %s

Usage:
   • secret %[3]s -b
   • secret --path %[3]s -b
`, cst.NounSecret, cst.ProductName, cst.ExamplePath),
		FlagsPredictor: GetNoDataOpWrappers(cst.NounSecret),
		MinNumberArgs:  1,
	})
}

func GetReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounSecret, cst.Read},
		RunFunc: Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil, cst.NounSecret}.handleReadCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounSecret, cst.Read),
		HelpText: fmt.Sprintf(`Read a %[2]s from %[3]s

Usage:
   • secret %[1]s %[4]s -b
   • secret %[1]s --path %[4]s -bf data.Data.Key
   • secret %[1]s --version
`, cst.Read, cst.NounSecret, cst.ProductName, cst.ExamplePath),
		FlagsPredictor:    GetNoDataOpWrappers(cst.NounSecret),
		ArgsPredictorFunc: preds.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
	})
}

func GetDescribeCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounSecret, cst.Describe},
		RunFunc: Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil, cst.NounSecret}.handleDescribeCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounSecret, cst.Describe),
		HelpText: fmt.Sprintf(`Describe a %[2]s from %[3]s

Usage:
   • secret %[1]s %[4]s
   • secret %[1]s --path %[4]s -f id
`, cst.Describe, cst.NounSecret, cst.ProductName, cst.ExamplePath),
		FlagsPredictor:    GetNoDataOpWrappers(cst.NounSecret),
		ArgsPredictorFunc: preds.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
	})
}

func GetDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounSecret, cst.Delete},
		RunFunc: Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil, cst.NounSecret}.handleDeleteCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounSecret, cst.Delete),
		HelpText: fmt.Sprintf(`Delete a %[2]s from %[3]s

Usage:
   • secret %[1]s %[4]s
   • secret %[1]s --path %[4]s --force
`, cst.Delete, cst.NounSecret, cst.ProductName, cst.ExamplePath),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):  cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounSecret)}), false},
			preds.LongFlag(cst.ID):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.ID, Shorthand: "i", Usage: fmt.Sprintf("Target %s for a %s", cst.ID, cst.NounSecret)}), false},
			preds.LongFlag(cst.Force): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Force, Shorthand: "", Usage: fmt.Sprintf("Immediately delete %s", cst.NounSecret), Global: false, ValueType: "bool"}), false},
		},
		ArgsPredictorFunc: preds.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
	})
}

func GetSecretRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounSecret, cst.Read},
		RunFunc: Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil, cst.NounSecret}.handleRestoreCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounSecret, cst.Restore),
		HelpText: fmt.Sprintf(`Restore a deleted %[2]s from %[3]s

Usage:
   • secret %[1]s %[4]s
`, cst.Restore, cst.NounSecret, cst.ProductName, cst.ExamplePath),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path): cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounSecret)}), false},
			preds.LongFlag(cst.ID):   cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.ID, Shorthand: "i", Usage: fmt.Sprintf("Target %s for a %s", cst.ID, cst.NounSecret)}), false},
		},
		ArgsPredictorFunc: preds.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
	})
}

func GetUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounSecret, cst.Update},
		RunFunc: Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil, cst.NounSecret}.handleUpsertCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> <data> | (--path|-r) (--data|-d))", cst.NounSecret, cst.Update),
		HelpText: fmt.Sprintf(`Update a %[2]s in %[3]s

Usage:
   • secret %[1]s %[4]s %[5]s
   • secret %[1]s --path %[4]s --data %[5]s
   • secret %[1]s --path %[4]s --data %[6]s
`, cst.Update, cst.NounSecret, cst.ProductName, cst.ExamplePath, cst.ExampleDataJSON, cst.ExampleDataPath),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Data):            cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("%s to be stored in a %s. Prefix with '@' to denote filepath (required)", strings.Title(cst.Data), cst.NounSecret)}), false},
			preds.LongFlag(cst.Path):            cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounSecret)}), false},
			preds.LongFlag(cst.DataDescription): cli.PredictorWrapper{complete.PredictNothing, preds.NewFlagValue(preds.Params{Name: cst.DataDescription, Usage: fmt.Sprintf("Description of a %s", cst.NounSecret)}), false},
			preds.LongFlag(cst.DataAttributes):  cli.PredictorWrapper{complete.PredictNothing, preds.NewFlagValue(preds.Params{Name: cst.DataAttributes, Usage: fmt.Sprintf("Attributes of a %s", cst.NounSecret)}), false},
			preds.LongFlag(cst.Overwrite):       cli.PredictorWrapper{complete.PredictNothing, preds.NewFlagValue(preds.Params{Name: cst.Overwrite, Usage: fmt.Sprintf("Overwrite all the contents of %s data", cst.NounSecret), Global: false, ValueType: "bool"}), false},
			preds.LongFlag(cst.ID):              cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.ID, Shorthand: "i", Usage: fmt.Sprintf("Target %s for a %s", cst.ID, cst.NounSecret)}), false},
		},
		ArgsPredictorFunc: preds.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     0,
	})
}

func GetRollbackCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounSecret, cst.Rollback},
		RunFunc: Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil, cst.NounSecret}.handleRollbackCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> | (--path|-r))", cst.NounSecret, cst.Rollback),
		HelpText: fmt.Sprintf(`Rollback a %[2]s in %[3]s

Usage:
   • secret %[1]s %[4]s --%[5]s 1
   • secret %[1]s --path %[4]s
`, cst.Rollback, cst.NounSecret, cst.ProductName, cst.ExamplePath, cst.Version),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):    cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounSecret)}), false},
			preds.LongFlag(cst.ID):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.ID, Shorthand: "i", Usage: fmt.Sprintf("Target %s for a %s", cst.ID, cst.NounSecret)}), false},
			preds.LongFlag(cst.Version): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Version, Shorthand: "", Usage: "The version to which to rollback"}), false},
		},
		ArgsPredictorFunc: preds.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
	})
}

func GetEditCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounSecret, cst.Update},
		RunFunc: Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, EditData, cst.NounSecret}.handleEditCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> <data> | (--path|-r))", cst.NounSecret, cst.Edit),
		HelpText: fmt.Sprintf(`Edit a %[2]s in %[3]s

Usage:
   • secret %[1]s %[4]s
   • secret %[1]s --path %[4]s
`, cst.Edit, cst.NounSecret, cst.ProductName, cst.ExamplePath),
		FlagsPredictor:    GetNoDataOpWrappers(cst.NounSecret),
		ArgsPredictorFunc: preds.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
	})
}

func GetCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounSecret, cst.Create},
		RunFunc: Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil, cst.NounSecret}.handleUpsertCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> <data> | (--path|-r) (--data|-d))", cst.NounSecret, cst.Create),
		HelpText: fmt.Sprintf(`Create a %[2]s in %[3]s

Usage:
   • secret %[1]s %[4]s --data %[5]s
   • secret %[1]s --path %[4]s --data %[5]s
   • secret %[1]s --path %[4]s --data %[6]s
`, cst.Create, cst.NounSecret, cst.ProductName, cst.ExamplePath, cst.ExampleDataJSON, cst.ExampleDataPath),
		FlagsPredictor: GetDataOpWrappers(cst.NounSecret),
		MinNumberArgs:  0,
	})
}

func GetBustCacheCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounSecret, cst.BustCache},
		RunFunc: Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil, cst.NounSecret}.handleBustCacheCmd,
		SynopsisText: fmt.Sprintf("%s %s", cst.NounSecret, cst.BustCache),
		HelpText: `Bust secret cache

Usage:
   • secret bustcache
`,
	})
}

func GetSecretSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounSecret, cst.Search},
		RunFunc: Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil, cst.NounSecret}.handleSecretSearchCmd,
		SynopsisText: fmt.Sprintf("%s (<query> | --query) --limit[default:25] --cursor --search-type[default:string] --search-comparison[default:contains] --search-field[default:path] --search-links[default:false])", cst.Search),
		HelpText: fmt.Sprintf(`Search for a %[2]s from %[3]s

Usage:
    • %[2]s %[1]s %[4]s
    • %[2]s %[1]s --query %[4]s
    • %[2]s %[1]s --query aws:base:secret --search-links
    • %[2]s %[1]s --query aws --search-field attributes.type
    • %[2]s %[1]s --query 900 --search-field attributes.ttl --search-type number
    • %[2]s %[1]s --query production --search-field attributes.stage --search-comparison equal
`, cst.Search, cst.NounSecret, cst.ProductName, cst.ExampleUserSearch),
		FlagsPredictor: GetSearchOpWrappers(),
		MinNumberArgs:  1,
	})
}

func (se Secret) handleBustCacheCmd(args []string) int {
	var err *errors.ApiError
	var s store.Store
	st := viper.GetString(cst.StoreType)
	if s, err = se.getStore(st); err == nil {
		err = s.Wipe(cst.SecretRoot)
		err = s.Wipe(cst.SecretDescriptionRoot).And(err)
	}
	if err == nil {
		log.Print("Successfully cleared local cache")
	}
	outClient := se.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}

	outClient.WriteResponse(nil, err)
	return 0
}

func (se Secret) handleDescribeCmd(args []string) int {
	id := viper.GetString(cst.ID)
	path := viper.GetString(cst.Path)
	if path == "" {
		path = viper.GetString(cst.ID)
		id = ""
	}
	if path == "" {
		path = paths.GetPath(args)
	}
	resp, err := se.getSecret(path, id, false, cst.SuffixDescription)
	outClient := se.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}

	outClient.WriteResponse(resp, err)

	return 0
}

func (se Secret) handleReadCmd(args []string) int {
	id := viper.GetString(cst.ID)
	path := viper.GetString(cst.Path)
	if path == "" {
		path = viper.GetString(cst.ID)
		id = ""
	}
	if path == "" {
		path = paths.GetPath(args)
	}
	version := viper.GetString(cst.Version)
	if strings.TrimSpace(version) != "" {
		path = fmt.Sprint(path, "/", cst.Version, "/", version)
	}
	resp, err := se.getSecret(path, id, false, "")
	if se.outClient == nil {
		se.outClient = format.NewDefaultOutClient()
	}

	se.outClient.WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func (se Secret) handleRestoreCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	if se.outClient == nil {
		se.outClient = format.NewDefaultOutClient()
	}

	path := viper.GetString(cst.Path)
	if path == "" {
		path = viper.GetString(cst.ID)
	}
	if path == "" {
		path = paths.GetPath(args)
	}

	rc, rerr := getResourceConfig(path, string(se.secretType))
	if rerr != nil {
		se.outClient.Fail(rerr)
		return utils.GetExecStatus(rerr)
	}
	path = rc.path
	uri := paths.CreateResourceURI(rc.resourceType, path, "/restore", true, nil, rc.pluralize)
	data, err = se.request.DoRequest(http.MethodPut, uri, nil)

	se.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (se Secret) handleSecretSearchCmd(args []string) int {
	if se.outClient == nil {
		se.outClient = format.NewDefaultOutClient()
	}
	var err *errors.ApiError
	var data []byte
	query := viper.GetString(cst.Query)
	limit := viper.GetString(cst.Limit)
	cursor := viper.GetString(cst.Cursor)
	searchLinks := viper.GetBool(cst.SearchLinks)
	searchType := viper.GetString(cst.SearchType)
	searchComparison := viper.GetString(cst.SearchComparison)
	searchField := viper.GetString(cst.SearchField)
	sort := viper.GetString(cst.Sort)
	if query == "" && len(args) > 0 {
		query = args[0]
	}
	if query == "" {
		err = errors.NewS("error: must specify " + cst.Query)
	} else {
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
		rc, rerr := getResourceConfig("", string(se.secretType))
		if rerr != nil {
			se.outClient.Fail(rerr)
			return utils.GetExecStatus(err)
		}
		uri := paths.CreateResourceURI(rc.resourceType, "", "", false, queryParams, rc.pluralize)
		data, err = se.request.DoRequest(http.MethodGet, uri, nil)
	}

	se.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (se Secret) handleDeleteCmd(args []string) int {
	outClient := se.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	var err *errors.ApiError
	var resp []byte
	id := viper.GetString(cst.ID)
	path := viper.GetString(cst.Path)
	force := viper.GetBool(cst.Force)
	if path == "" {
		path = viper.GetString(cst.ID)
		id = ""
	}
	if path == "" {
		path = paths.GetPath(args)
	}

	query := map[string]string{"force": strconv.FormatBool(force)}
	rc, rerr := getResourceConfig(path, string(se.secretType))
	if rerr != nil {
		outClient.Fail(err)
		return utils.GetExecStatus(err)
	}
	path = rc.path
	uri, err := paths.GetResourceURIFromResourcePath(rc.resourceType, path, id, "", query, rc.pluralize)
	resp, err = se.request.DoRequest(http.MethodDelete, uri, nil)

	outClient.WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func (se Secret) handleRollbackCmd(args []string) int {
	var apiError *errors.ApiError
	var resp []byte
	if se.outClient == nil {
		se.outClient = format.NewDefaultOutClient()
	}

	path := viper.GetString(cst.Path)
	if path == "" {
		path = viper.GetString(cst.ID)
	}
	if path == "" {
		path = paths.GetPath(args)
	}
	version := viper.GetString(cst.Version)
	rc, err := getResourceConfig(path, string(se.secretType))
	if err != nil {
		se.outClient.Fail(err)
		return utils.GetExecStatus(err)
	}

	path = rc.path

	// If version is not provided, get the secret's description and parse the version from it.
	// Submit a request for a version that's previous relative to the one found.
	if version == "" {
		id := viper.GetString(cst.ID)
		resp, apiError = se.getSecret(path, id, false, cst.SuffixDescription)
		if apiError != nil {
			se.outClient.WriteResponse(resp, apiError)
			return utils.GetExecStatus(apiError)
		}

		v, err := utils.GetPreviousVersion(resp)
		if err != nil {
			se.outClient.Fail(err)
			return 1
		}
		version = v
	}
	if strings.TrimSpace(version) != "" {
		path = fmt.Sprint(path, "/rollback/", version)
	}
	uri := paths.CreateResourceURI(rc.resourceType, path, "", true, nil, rc.pluralize)
	resp, apiError = se.request.DoRequest(http.MethodPut, uri, nil)

	se.outClient.WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func (se Secret) handleUpsertCmd(args []string) int {
	if se.outClient == nil {
		se.outClient = format.NewDefaultOutClient()
	}

	action := strings.ToLower(viper.GetString(cst.LastCommandKey))

	if OnlyGlobalArgs(args) {
		ui := &cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		}
		var (
			resp []byte
			err  *errors.ApiError
		)
		switch action {
		case cst.Create:
			resp, err = se.handleCreateWorkflow(ui)
		case cst.Update:
			resp, err = se.handleUpdateWorkflow(ui)
		}
		se.outClient.WriteResponse(resp, err)
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

	rc, rerr := getResourceConfig(path, string(se.secretType))
	if rerr != nil {
		se.outClient.Fail(rerr)
		return utils.GetExecStatus(rerr)
	}
	path = rc.path
	uri, err := paths.GetResourceURIFromResourcePath(rc.resourceType, path, id, "", nil, rc.pluralize)
	if err != nil {
		se.outClient.FailE(err)
		return utils.GetExecStatus(err)
	}

	if data == "" && desc == "" && len(attributes) == 0 {
		se.outClient.FailF("Please provide a properly formed value for at least --%s, or --%s, or --%s.",
			cst.Data, cst.DataDescription, cst.DataAttributes)
		return 1
	}

	dataMap := make(map[string]interface{})
	if data != "" {
		parseErr := json.Unmarshal([]byte(data), &dataMap)
		if parseErr != nil {
			se.outClient.FailF("Failed to parse passed in secret data: %v", parseErr)
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
	resp, err := se.request.DoRequest(reqMethod, uri, &postData)

	se.outClient.WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func (se Secret) handleCreateWorkflow(ui cli.Ui) ([]byte, *errors.ApiError) {
	dataMap := make(map[string]interface{})
	attrMap := make(map[string]interface{})

	path, err := prompt.Ask(ui, "Path:")
	if err != nil {
		return nil, errors.New(err)
	}
	desc, err := prompt.AskDefault(ui, "Description:", "")
	if err != nil {
		return nil, errors.New(err)
	}
	attrAction, err := prompt.Choose(
		ui, "Add Attributes?:",
		prompt.Option{"skip", "Skip"},
		prompt.Option{"kv", "Add key/value pairs"},
		prompt.Option{"json", "Define as a json string"},
	)
	if err != nil {
		return nil, errors.New(err)
	}

	switch attrAction {
	case "kv":
		for {
			attrKey, err := prompt.Ask(ui, "Key:")
			if err != nil {
				return nil, errors.New(err)
			}
			attrVal, err := prompt.Ask(ui, "Value:")
			if err != nil {
				return nil, errors.New(err)
			}
			attrMap[attrKey] = attrVal

			yes, err := prompt.YesNo(ui, "Add more?", false)
			if err != nil {
				return nil, errors.New(err)
			}
			if !yes {
				break
			}
		}
	case "json":
		for {
			attr, err := prompt.Ask(ui, "Attributes:")
			if err != nil {
				return nil, errors.New(err)
			}
			if len(attr) > 0 {
				err := json.Unmarshal([]byte(attr), &attrMap)
				if err != nil {
					ui.Output(fmt.Sprintf("Invalid JSON: %v", err))
					continue
				}
			}
			break
		}
	}

	dataAction, err := prompt.Choose(
		ui, "Add Data?:",
		prompt.Option{"skip", "Skip"},
		prompt.Option{"kv", "Add key/value pairs"},
		prompt.Option{"json", "Define as a json string"},
	)
	if err != nil {
		return nil, errors.New(err)
	}
	switch dataAction {
	case "kv":
		for {
			dataKey, err := prompt.Ask(ui, "Key:")
			if err != nil {
				return nil, errors.New(err)
			}
			dataVal, err := prompt.Ask(ui, "Value:")
			if err != nil {
				return nil, errors.New(err)
			}
			dataMap[dataKey] = dataVal

			yes, err := prompt.YesNo(ui, "Add more?", false)
			if err != nil {
				return nil, errors.New(err)
			}
			if !yes {
				break
			}
		}
	case "json":
		for {
			data, err := prompt.Ask(ui, "Data:")
			if err != nil {
				return nil, errors.New(err)
			}
			if len(data) > 0 {
				err := json.Unmarshal([]byte(data), &dataMap)
				if err != nil {
					ui.Output(fmt.Sprintf("Invalid JSON: %v", err))
					continue
				}
			}
			break
		}
	}

	rc, err := getResourceConfig(path, string(se.secretType))
	if err != nil {
		return nil, errors.New(err)
	}

	postData := secretUpsertBody{
		Description: desc,
		Data:        dataMap,
		Attributes:  attrMap,
	}

	uri, apiErr := paths.GetResourceURIFromResourcePath(rc.resourceType, path, "", "", nil, rc.pluralize)
	if apiErr != nil {
		return nil, apiErr
	}

	return se.request.DoRequest(http.MethodPost, uri, &postData)
}

func (se Secret) handleUpdateWorkflow(ui cli.Ui) ([]byte, *errors.ApiError) {
	path, rerr := prompt.Ask(ui, "Path:")
	if rerr != nil {
		return nil, errors.New(rerr)
	}

	rc, rerr := getResourceConfig(path, string(se.secretType))
	if rerr != nil {
		return nil, errors.New(rerr)
	}
	path = rc.path
	uri, err := paths.GetResourceURIFromResourcePath(rc.resourceType, path, "", "", nil, rc.pluralize)
	if err != nil {
		return nil, err
	}

	isSecretRetrieved := false
	secretResp := &secretGetResponse{}

	secretBytes, err := se.request.DoRequest(http.MethodGet, uri, nil)
	if err != nil {
		httpResp := err.HttpResponse()
		if httpResp == nil {
			return nil, err
		}
		switch httpResp.StatusCode {
		case http.StatusNotFound:
			return nil, errors.NewS("Secret under that path cannot be found.")
		case http.StatusForbidden:
			yes, err := prompt.YesNo(ui, "You are not allowed to read secret under that path. Do you want to continue?", true)
			if err != nil {
				return nil, errors.New(err)
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
			ui.Info("Currently description is empty.")
		} else {
			ui.Info(fmt.Sprintf("Current description: %q", secretResp.Description))
		}
	}
	doUpdDescription, rerr := prompt.YesNo(ui, "Update description?", false)
	if rerr != nil {
		return nil, errors.New(rerr)
	}

	var desc string
	if doUpdDescription {
		desc, rerr = prompt.AskDefault(ui, "Description:", "")
		if rerr != nil {
			return nil, errors.New(rerr)
		}
	}

	if isSecretRetrieved {
		ui.Info("Attributes and data defined currently for the secret:")
		// Print attributes and data beautifully :)
		outClient := format.NewDefaultOutClient()
		relativeData, err := json.Marshal(map[string]interface{}{
			"attributes": secretResp.Attributes,
			"data":       secretResp.Data,
		})
		if err == nil {
			outClient.WriteResponse(relativeData, nil)
		}
	}
	overwrite, rerr := prompt.YesNo(ui, "Overwrite existing attributes and data?", false)
	if rerr != nil {
		return nil, errors.New(rerr)
	}

	dataMap := make(map[string]interface{})
	attrMap := make(map[string]interface{})

	var q string
	if overwrite {
		q = "Overwrite attributes?"
	} else {
		q = "Update attributes?"
	}
	attrAction, rerr := prompt.Choose(
		ui, q,
		prompt.Option{"skip", "No, skip"},
		prompt.Option{"kv", "Yes, use key/value pairs"},
		prompt.Option{"json", "Yes, define as a json string"},
	)
	if rerr != nil {
		return nil, errors.New(rerr)
	}

	switch attrAction {
	case "kv":
		for {
			attrKey, err := prompt.Ask(ui, "Key:")
			if err != nil {
				return nil, errors.New(err)
			}
			attrVal, err := prompt.Ask(ui, "Value:")
			if err != nil {
				return nil, errors.New(err)
			}
			attrMap[attrKey] = attrVal

			yes, err := prompt.YesNo(ui, "Add more?", false)
			if err != nil {
				return nil, errors.New(err)
			}
			if !yes {
				break
			}
		}
	case "json":
		for {
			attr, err := prompt.Ask(ui, "Attributes:")
			if err != nil {
				return nil, errors.New(err)
			}
			if len(attr) > 0 {
				err := json.Unmarshal([]byte(attr), &attrMap)
				if err != nil {
					ui.Output(fmt.Sprintf("Invalid JSON: %v", err))
					continue
				}
			}
			break
		}
	}

	if overwrite {
		q = "Overwrite data?"
	} else {
		q = "Update data?"
	}
	dataAction, rerr := prompt.Choose(
		ui, q,
		prompt.Option{"skip", "No, skip"},
		prompt.Option{"kv", "Yes, use key/value pairs"},
		prompt.Option{"json", "Yes, define as a json string"},
	)
	if rerr != nil {
		return nil, errors.New(rerr)
	}
	switch dataAction {
	case "kv":
		for {
			dataKey, err := prompt.Ask(ui, "Key:")
			if err != nil {
				return nil, errors.New(err)
			}
			dataVal, err := prompt.Ask(ui, "Value:")
			if err != nil {
				return nil, errors.New(err)
			}
			dataMap[dataKey] = dataVal

			yes, err := prompt.YesNo(ui, "Add more?", false)
			if err != nil {
				return nil, errors.New(err)
			}
			if !yes {
				break
			}
		}
	case "json":
		for {
			data, err := prompt.Ask(ui, "Data:")
			if err != nil {
				return nil, errors.New(err)
			}
			if len(data) > 0 {
				err := json.Unmarshal([]byte(data), &dataMap)
				if err != nil {
					ui.Output(fmt.Sprintf("Invalid JSON: %v", err))
					continue
				}
			}
			break
		}
	}

	if !doUpdDescription && attrAction == "skip" && dataAction == "skip" {
		ui.Output("Nothing to update. Exiting.")
		return nil, nil
	}

	postData := secretUpsertBody{
		Description: desc,
		Data:        dataMap,
		Attributes:  attrMap,
		Overwrite:   overwrite,
	}

	ui.Output("Sending request...")
	return se.request.DoRequest(http.MethodPut, uri, &postData)
}

func (se Secret) handleEditCmd(args []string) int {
	if se.outClient == nil {
		se.outClient = format.NewDefaultOutClient()
	}

	var resp []byte
	id := viper.GetString(cst.ID)
	path := viper.GetString(cst.Path)
	if path == "" {
		path = viper.GetString(cst.ID)
		id = ""
	}
	if path == "" {
		path = paths.GetPath(args)
	}
	rc, rerr := getResourceConfig(path, string(se.secretType))
	if rerr != nil {
		se.outClient.Fail(rerr)
		return utils.GetExecStatus(rerr)
	}
	path = rc.path
	uri, err := paths.GetResourceURIFromResourcePath(rc.resourceType, path, id, "", nil, rc.pluralize)

	// fetch
	resp, err = se.getSecret(path, id, true, "")
	if err != nil {
		se.outClient.WriteResponse(resp, err)
		return utils.GetExecStatus(err)
	}

	saveFunc := dataFunc(func(data []byte) (resp []byte, err *errors.ApiError) {
		var model secretUpsertBody
		if mErr := json.Unmarshal(data, &model); mErr != nil {
			return nil, errors.New(mErr).Grow("invalid format for secret")
		}
		model.Overwrite = true
		_, err = se.request.DoRequest(http.MethodPut, uri, &model)
		return nil, err
	})
	resp, err = se.edit(resp, saveFunc, nil, false)
	se.outClient.WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func (se Secret) getSecret(path string, id string, edit bool, requestSuffix string) (respData []byte, err *errors.ApiError) {
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
		cacheData, expired = se.getSecretDataFromCache(secretKey, st)
		if !expired && len(cacheData) > 0 {
			return cacheData, nil
		}
	}

	queryTerms := map[string]string{"edit": strconv.FormatBool(edit)}

	rc, rerr := getResourceConfig(path, string(se.secretType))
	if rerr != nil {
		return nil, errors.New(rerr)
	}
	path = rc.path
	uri, err := paths.GetResourceURIFromResourcePath(rc.resourceType, path, id, requestSuffix, queryTerms, rc.pluralize)
	if err != nil {
		return nil, err
	}
	respData, err = se.request.DoRequest(http.MethodGet, uri, nil)
	if err == nil {
		if cacheStrategy != store.Never {
			if s == nil {
				var errStore *errors.ApiError
				s, errStore = se.getStore(st)
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
		cacheData, expired = se.getSecretDataFromCache(secretKey, st)
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

func (se Secret) getSecretDataFromCache(secretKey string, st string) (cacheData []byte, expired bool) {
	if s, err := se.getStore(st); err != nil {
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
	pluralize    bool
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
			pluralize:    false,
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
			resourceType: resourceType,
			pluralize:    true,
			path:         path,
		}
		return rc, nil
	}
}
