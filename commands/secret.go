package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	cst "thy/constants"
	"thy/errors"
	"thy/format"
	preds "thy/predictors"
	"thy/requests"
	"thy/store"
	"thy/utils"

	"github.com/apex/log"
	"github.com/posener/complete"
	"github.com/thycotic-rd/cli"
	"github.com/thycotic-rd/viper"
)

type Secret struct {
	request   requests.Client
	outClient format.OutClient
	getStore  func(stString string) (store.Store, *errors.ApiError)
	edit      func([]byte, dataFunc, *errors.ApiError, bool) ([]byte, *errors.ApiError)
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
		preds.LongFlag(cst.Version): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Version, Shorthand: "", Usage: "List the current and last (n) versions"}), false},
	}
}

func GetSecretCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounSecret},
		RunFunc: func(args []string) int {
			id := viper.GetString(cst.ID)
			path := viper.GetString(cst.Path)
			if path == "" {
				path = utils.GetPath(args)
			}
			if path == "" && id == "" {
				return cli.RunResultHelp
			}
			return Secret{
				requests.NewHttpClient(),
				nil,
				store.GetStore, nil}.handleReadCmd(args)
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
			store.GetStore, nil}.handleReadCmd,
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
			store.GetStore, nil}.handleDescribeCmd,
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
			store.GetStore, nil}.handleDeleteCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounSecret, cst.Delete),
		HelpText: fmt.Sprintf(`Delete a %[2]s from %[3]s
Usage:
	• secret %[1]s %[4]s
	• secret %[1]s --path %[4]s --force
				`, cst.Delete, cst.NounSecret, cst.ProductName, cst.ExamplePath),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):  cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounSecret)}), false},
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
			store.GetStore, nil}.handleRestoreCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounSecret, cst.Restore),
		HelpText: fmt.Sprintf(`Restore a deleted %[2]s from %[3]s
Usage:
	• secret %[1]s %[4]s

				`, cst.Restore, cst.NounSecret, cst.ProductName, cst.ExamplePath),
		FlagsPredictor:    GetNoDataOpWrappers(cst.NounSecret),
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
			store.GetStore, nil}.handleUpsertCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> <data> | (--path|-r) (--data|-d))", cst.NounSecret, cst.Update),
		HelpText: fmt.Sprintf(`Update a %[2]s in %[3]s
Usage:
	• secret %[1]s %[4]s %[5]s
	• secret %[1]s --path %[4]s --data %[5]s
	• secret %[1]s --path %[4]s --data %[6]s
				`, cst.Update, cst.NounSecret, cst.ProductName, cst.ExamplePath, cst.ExampleDataJSON, cst.ExampleDataPath),
		FlagsPredictor:    GetDataOpWrappers(cst.NounSecret),
		ArgsPredictorFunc: preds.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     2,
	})
}

func GetRollbackCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounSecret, cst.Rollback},
		RunFunc: Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil}.handleRollbackCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> | (--path|-r))", cst.NounSecret, cst.Rollback),
		HelpText: fmt.Sprintf(`Rollback a %[2]s in %[3]s
Usage:
	• secret %[1]s %[4]s --%[5]s 1
	• secret %[1]s --path %[4]s
				`, cst.Rollback, cst.NounSecret, cst.ProductName, cst.ExamplePath, cst.Version),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):    cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounSecret)}), false},
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
			store.GetStore, EditData}.handleEditCmd,
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
			store.GetStore, nil}.handleUpsertCmd,
		SynopsisText: fmt.Sprintf("%s %s (<path> <data> | (--path|-r) (--data|-d))", cst.NounSecret, cst.Create),
		HelpText: fmt.Sprintf(`Create a %[2]s in %[3]s
Usage:
	• secret %[1]s %[4]s %[5]s
	• secret %[1]s --path %[4]s --data %[5]s
	• secret %[1]s --path %[4]s --data %[6]s
				`, cst.Create, cst.NounSecret, cst.ProductName, cst.ExamplePath, cst.ExampleDataJSON, cst.ExampleDataPath),
		FlagsPredictor: GetDataOpWrappers(cst.NounSecret),
		MinNumberArgs:  2,
	})
}

func GetBustCacheCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounSecret, cst.BustCache},
		RunFunc: Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil}.handleBustCacheCmd,
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
			store.GetStore, nil}.handleSecretSearchCmd,
		SynopsisText: fmt.Sprintf("%s (<query> | --query)", cst.Search),
		HelpText: fmt.Sprintf(`Search for a %[2]s from %[3]s

Usage:
    • %[2]s %[1]s %[4]s
    • %[2]s %[1]s --query %[4]s
    • %[2]s %[1]s --query aws:base:secret --search-links
            `, cst.Search, cst.NounSecret, cst.ProductName, cst.ExampleUserSearch),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Query):       cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Query, Shorthand: "q", Usage: fmt.Sprintf("%s of %ss to fetch (required)", strings.Title(cst.Query), cst.NounSecret)}), false},
			preds.LongFlag(cst.SearchLinks): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.SearchLinks, Shorthand: "", Usage: "Find secrets that link to the secret path in the query", Global: false, ValueType: "bool"}), false},
			preds.LongFlag(cst.Limit):       cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Limit, Shorthand: "l", Usage: fmt.Sprint("Maximum number of results per cursor (optional)")}), false},
			preds.LongFlag(cst.Cursor):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Cursor, Usage: fmt.Sprint("Next cursor for additional results (optional)")}), false},
		},
		MinNumberArgs: 1,
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
		log.Info("Successfully cleared local cache")
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
		path = utils.GetPath(args)
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
		path = utils.GetPath(args)
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
		path = utils.GetPath(args)
	}

	uri, err := utils.GetResourceURIFromResourcePath(cst.NounSecret, path, "", "", true, nil)
	if err == nil {
		uri += "/restore"
		data, err = se.request.DoRequest("PUT", uri, nil)
	}

	se.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (se Secret) handleSecretSearchCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	query := viper.GetString(cst.Query)
	limit := viper.GetString(cst.Limit)
	cursor := viper.GetString(cst.Cursor)
	searchLinks := viper.GetBool(cst.SearchLinks)
	if query == "" && len(args) > 0 {
		query = args[0]
	}
	if query == "" {
		err = errors.NewS("error: must specify " + cst.Query)
	} else {
		queryParams := map[string]string{
			cst.SearchKey: query,
			cst.Limit:     limit,
			cst.Cursor:    cursor,
		}
		if searchLinks {
			//flag just needs to be present
			queryParams[cst.SearchLinks] = strconv.FormatBool(searchLinks)
		}
		uri := utils.CreateResourceURI(cst.NounSecret, "", "", false, queryParams, true)
		data, err = se.request.DoRequest("GET", uri, nil)
	}
	outClient := se.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}

	outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (se Secret) handleDeleteCmd(args []string) int {
	var err *errors.ApiError
	var resp []byte
	path := viper.GetString(cst.Path)
	force := viper.GetBool(cst.Force)
	if path == "" {
		path = utils.GetPath(args)
	}

	query := map[string]string{"force": strconv.FormatBool(force)}
	uri := utils.CreateResourceURI(cst.NounSecret, path, "", true, query, true)

	resp, err = se.request.DoRequest("DELETE", uri, nil)
	outClient := se.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
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
		path = utils.GetPath(args)
	}
	version := viper.GetString(cst.Version)

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
	uri := utils.CreateResourceURI(cst.NounSecret, path, "", true, nil, true)
	resp, apiError = se.request.DoRequest("PUT", uri, nil)

	se.outClient.WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func (se Secret) handleUpsertCmd(args []string) int {
	var err *errors.ApiError
	var resp []byte
	id := viper.GetString(cst.ID)
	path := viper.GetString(cst.Path)
	overwrite := viper.GetBool(cst.Overwrite)
	if path == "" {
		path = utils.GetPath(args)
	}
	if se.outClient == nil {
		se.outClient = format.NewDefaultOutClient()
	}
	uri, err := utils.GetResourceURIFromResourcePath(cst.NounSecret, path, id, "", true, nil)
	if err == nil {
		data := viper.GetString(cst.Data)
		desc := viper.GetString(cst.DataDescription)
		attributes := viper.GetStringMap(cst.DataAttributes)
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
		reqMethod := strings.ToLower(viper.GetString(cst.LastCommandKey))
		if reqMethod == cst.Create {
			reqMethod = "POST"
		} else {
			reqMethod = "PUT"
		}
		resp, err = se.request.DoRequest(reqMethod, uri, &postData)
	}

	se.outClient.WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func (se Secret) handleEditCmd(args []string) int {
	if se.outClient == nil {
		se.outClient = format.NewDefaultOutClient()
	}

	var resp []byte
	id := viper.GetString(cst.ID)
	path := viper.GetString(cst.Path)
	if path == "" {
		path = utils.GetPath(args)
	}
	uri, err := utils.GetResourceURIFromResourcePath(cst.NounSecret, path, id, "", true, nil)

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
		_, err = se.request.DoRequest("PUT", uri, &model)
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
	uri, err := utils.GetResourceURIFromResourcePath(cst.NounSecret, path, id, requestSuffix, true, queryTerms)
	if err != nil {
		return nil, err
	}
	respData, err = se.request.DoRequest("GET", uri, nil)
	if err == nil {
		if cacheStrategy != store.Never {
			if s == nil {
				var errStore *errors.ApiError
				s, errStore = se.getStore(st)
				if errStore != nil {
					log.Error(fmt.Sprintf("Failed to get store to cache secret for store type %s. Error: %s", st, errStore))
				}
			}
			if s != nil {
				if errStore := s.Store(secretKey, secretData{
					Date: time.Now().UTC(),
					Data: respData,
				}); errStore != nil {
					log.Error(fmt.Sprintf("Failed to cache secret for store type %s. Error: %s", st, errStore))
				}
			}
		}
		return respData, nil
	} else if cacheStrategy == store.ServerThenCache {
		cacheData, expired = se.getSecretDataFromCache(secretKey, st)
		if !expired {
			return cacheData, nil
		}
	} else if cacheStrategy == store.CacheThenServerThenExpired && len(cacheData) > 0 {
		log.Info("Cache expired but failed to retrieve from server so returning cached data")
		return cacheData, nil
	}

	if expired == true {
		return nil, err.Or(errors.NewS("run in verbose mode for more information"))
	}
	return nil, err.Or(errors.NewS("run in verbose mode for more information"))
}

func (se Secret) getSecretDataFromCache(secretKey string, st string) (cacheData []byte, expired bool) {
	if s, err := se.getStore(st); err != nil {
		log.Error(fmt.Sprintf("Failed to get store of type %s. Error: %s", st, err.Error()))
	} else {
		var data secretData
		if err := s.Get(secretKey, &data); err != nil && len(data.Data) > 0 {
			log.Error(fmt.Sprintf("Failed to fetch cached secret from store type %s. Error: %s", st, err.Error()))
		} else {
			cacheData = data.Data
			cacheAgeMinutes := viper.GetInt(cst.CacheAge)
			if cacheAgeMinutes > 0 {
				expired = (data.Date.Sub(time.Now().UTC()).Seconds() + float64(cacheAgeMinutes)*60) < 0
			} else {
				log.Error(fmt.Sprintf("Invalid cache age: %d", cacheAgeMinutes))
			}
		}
	}

	return cacheData, expired
}

type secretData struct {
	Date time.Time
	Data []byte
}

type secretUpsertBody struct {
	Data        map[string]interface{}
	Description string
	Attributes  map[string]interface{}
	Overwrite   bool
}
