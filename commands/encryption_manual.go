package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

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

type manualKeyEncryption struct {
	request   requests.Client
	outClient format.OutClient
}

func GetCryptoManualCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounEncryption, cst.Manual},
		RunFunc: func(args []string) int {
			return cli.RunResultHelp
		},
		SynopsisText:  "Encryption-as-a-Service with a Manual Key",
		HelpText:      "Encryption-as-a-Service with a Manual Key",
		MinNumberArgs: 0,
	})
}

func GetManualKeyUploadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Manual, cst.NounKey + "-" + cst.Upload},
		RunFunc:      manualKeyEncryption{requests.NewHttpClient(), nil}.handleUploadManualKey,
		SynopsisText: "Upload a new manual encryption/decryption key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s mykeys/key1 --%[5]s %[6]s --%[7]s %[8]s --%[9]s %[10]s
		`, cst.NounEncryption, cst.Manual, cst.NounKey+"-"+cst.Upload, cst.Path,
			cst.Scheme, "symmetric", cst.PrivateKey, cst.ExamplePrivateKey, cst.Nonce, cst.ExampleNonce),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):       cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey)}), false},
			preds.LongFlag(cst.PrivateKey): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.PrivateKey, Usage: "Private key base64-encoded (required)"}), false},
			preds.LongFlag(cst.Scheme):     cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Scheme, Usage: "Encryption scheme (required)"}), false},
			preds.LongFlag(cst.Nonce):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Nonce, Usage: "Nonce base64-encoded (optional)"}), false},
			preds.LongFlag(cst.Metadata):   cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Metadata, Usage: "Metadata as a JSON object (optional)"}), false},
		},
		MinNumberArgs: 6,
	})
}

func GetManualKeyUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Manual, cst.NounKey + "-" + cst.Update},
		RunFunc:      manualKeyEncryption{requests.NewHttpClient(), nil}.handleUpdateManualKey,
		SynopsisText: "Update an existing encryption key for encryption/decryption",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s mykeys/key1 --%[5]s %[6]s
		`, cst.NounEncryption, cst.Manual, cst.NounKey+"-"+cst.Update, cst.Path,
			cst.PrivateKey, cst.ExamplePrivateKey),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):       cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey)}), false},
			preds.LongFlag(cst.PrivateKey): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.PrivateKey, Usage: "Private key base64-encoded (required)"}), false},
			preds.LongFlag(cst.Nonce):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Nonce, Usage: "Nonce base64-encoded (optional)"}), false},
			preds.LongFlag(cst.Metadata):   cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Metadata, Usage: "Metadata as a JSON object (optional)"}), false},
		},
		MinNumberArgs: 4,
	})
}

func GetManualKeyReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Manual, cst.NounKey + "/" + cst.Read},
		RunFunc:      manualKeyEncryption{requests.NewHttpClient(), nil}.handleReadManualKey,
		SynopsisText: "Read an existing manual encryption key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s mykeys/key1
		`, cst.NounEncryption, cst.NounKey, cst.Read, cst.Path),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path): cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey)}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetManualKeyDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Manual, cst.NounKey + "/" + cst.Delete},
		RunFunc:      manualKeyEncryption{requests.NewHttpClient(), nil}.handleDeleteManualKey,
		SynopsisText: "Delete an existing manual encryption key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s mykeys/key1
		`, cst.NounEncryption, cst.Manual, cst.Delete, cst.Path),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):  cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey)}), false},
			preds.LongFlag(cst.Force): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Force, Shorthand: "", Usage: fmt.Sprintf("Immediately delete %s and all its versions", cst.NounKey), Global: false, ValueType: "bool"}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetManualKeyRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.NounKey + "/" + cst.Restore},
		RunFunc:      manualKeyEncryption{requests.NewHttpClient(), nil}.handleRestoreManualKey,
		SynopsisText: "Restore a previously soft-deleted manual encryption key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s mykeys/key1
		`, cst.NounEncryption, cst.NounKey, cst.Restore, cst.Path),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path): cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey)}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetManualKeyEncryptCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Encrypt},
		RunFunc:      manualKeyEncryption{requests.NewHttpClient(), nil}.handleManualKeyEncrypt,
		SynopsisText: "Encrypt data using a manual key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s 'hello world' --%[5]s mykeys/key1
   • %[1]s %[2]s %[3]s --%[4]s @mysecret.txt --%[5]s mykeys/key1
		`, cst.NounEncryption, cst.Manual, cst.Encrypt, cst.Data, cst.Path),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):    cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey)}), false},
			preds.LongFlag(cst.Version): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Version, Shorthand: "", Usage: fmt.Sprintf("Version of the %s used for encryption/decryption", cst.NounKey)}), false},
			preds.LongFlag(cst.Data):    cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: "A plaintext string or path to a @file with data to be encrypted"}), false},
			preds.LongFlag(cst.Output):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Output, Shorthand: "", Usage: "Output file for encrypted value and metadata"}), false},
		},
		MinNumberArgs: 4,
	})
}

func GetManualKeyDecryptCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Decrypt},
		RunFunc:      manualKeyEncryption{requests.NewHttpClient(), nil}.handleManualKeyDecrypt,
		SynopsisText: "Decrypt data using a manual key that had performed encryption",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s 'hello world' --%[5]s mykeys/key1
   • %[1]s %[2]s %[3]s --%[4]s @mysecret.txt"
		`, cst.NounEncryption, cst.Manual, cst.Decrypt, cst.Data, cst.Path, cst.Version),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):    cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s", cst.Path, cst.NounKey)}), false},
			preds.LongFlag(cst.Version): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Version, Shorthand: "", Usage: fmt.Sprintf("Version of the %s used for encryption/decryption", cst.NounKey)}), false},
			preds.LongFlag(cst.Data):    cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: "A ciphertext string or path to a @file with data to be decrypted"}), false},
			preds.LongFlag(cst.Output):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Output, Shorthand: "", Usage: "Output file for decrypted value and metadata"}), false},
		},
		MinNumberArgs: 2,
	})
}

func (e manualKeyEncryption) handleUploadManualKey(args []string) int {
	if e.outClient == nil {
		e.outClient = format.NewDefaultOutClient()
	}
	path := viper.GetString(cst.Path)
	if path == "" {
		e.outClient.FailF("%s is required", cst.Path)
		return 1
	}

	privateKey := viper.GetString(cst.PrivateKey)
	if privateKey == "" {
		e.outClient.FailF("%s is required", cst.PrivateKey)
		return 1
	}

	scheme := viper.GetString(cst.Scheme)
	if scheme == "" {
		e.outClient.FailF("%s is required", cst.Scheme)
		return 1
	}

	nonce := viper.GetString(cst.Nonce)
	metadata := viper.GetString(cst.Metadata)

	uri, err := e.makeKeyURL(path, nil)
	if err != nil {
		e.outClient.Fail(err)
		return utils.GetExecStatus(err)
	}

	body := manualKeyData{
		PrivateKey: privateKey,
		Nonce:      nonce,
		Scheme:     scheme,
	}

	if metadata != "" {
		var metadataMap map[string]interface{}
		if err := json.Unmarshal([]byte(metadata), &metadataMap); err != nil {
			e.outClient.FailS("failed to parse the metadata")
		} else {
			body.Metadata = metadataMap
		}
	}

	resp, apiError := e.request.DoRequest(http.MethodPost, uri, body)
	e.outClient.WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func (e manualKeyEncryption) handleUpdateManualKey(args []string) int {
	if e.outClient == nil {
		e.outClient = format.NewDefaultOutClient()
	}
	path := viper.GetString(cst.Path)
	if path == "" {
		e.outClient.FailF("%s is required", cst.Path)
		return 1
	}

	privateKey := viper.GetString(cst.PrivateKey)
	if privateKey == "" {
		e.outClient.FailF("%s is required", cst.PrivateKey)
		return 1
	}

	nonce := viper.GetString(cst.Nonce)
	metadata := viper.GetString(cst.Metadata)

	uri, err := e.makeKeyURL(path, nil)
	if err != nil {
		e.outClient.Fail(err)
		return utils.GetExecStatus(err)
	}

	body := manualKeyData{
		PrivateKey: privateKey,
		Nonce:      nonce,
	}

	if metadata != "" {
		var metadataMap map[string]interface{}
		if err := json.Unmarshal([]byte(metadata), &metadataMap); err != nil {
			e.outClient.FailS("failed to parse the metadata")
		} else {
			body.Metadata = metadataMap
		}
	}

	resp, apiError := e.request.DoRequest(http.MethodPut, uri, body)
	e.outClient.WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func (e manualKeyEncryption) handleReadManualKey(args []string) int {
	if e.outClient == nil {
		e.outClient = format.NewDefaultOutClient()
	}

	path := viper.GetString(cst.Path)
	if path == "" {
		path = paths.GetPath(args)
	}

	uri, err := e.makeKeyURL(path, nil)
	if err != nil {
		e.outClient.Fail(err)
		return utils.GetExecStatus(err)
	}

	resp, apiError := e.request.DoRequest(http.MethodGet, uri, nil)
	e.outClient.WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func (e manualKeyEncryption) handleDeleteManualKey(args []string) int {
	if e.outClient == nil {
		e.outClient = format.NewDefaultOutClient()
	}

	path := viper.GetString(cst.Path)
	if path == "" {
		path = paths.GetPath(args)
	}

	force := viper.GetBool(cst.Force)
	query := map[string]string{"force": strconv.FormatBool(force)}
	uri, err := e.makeKeyURL(path, query)
	if err != nil {
		e.outClient.Fail(err)
		return utils.GetExecStatus(err)
	}

	resp, apiError := e.request.DoRequest(http.MethodDelete, uri, nil)
	e.outClient.WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func (e manualKeyEncryption) handleRestoreManualKey(args []string) int {
	if e.outClient == nil {
		e.outClient = format.NewDefaultOutClient()
	}

	path := viper.GetString(cst.Path)
	if path == "" {
		path = paths.GetPath(args)
	}

	uri, err := e.makeKeyURL(path, nil)
	if err != nil {
		e.outClient.Fail(err)
		return utils.GetExecStatus(err)
	}

	resp, apiError := e.request.DoRequest(http.MethodPut, uri+"/restore", nil)
	e.outClient.WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func (e manualKeyEncryption) handleManualKeyEncrypt(args []string) int {
	if e.outClient == nil {
		e.outClient = format.NewDefaultOutClient()
	}
	path := viper.GetString(cst.Path)
	if path == "" {
		e.outClient.FailF("%s is required", cst.Path)
		return 1
	}

	data := viper.GetString(cst.Data)
	if data == "" {
		e.outClient.FailF("Please provide a value for %s. Either a string in quotes or a path to a file (@myfile.txt).", cst.Data)
		return 1
	}

	filename := paths.GetFilenameFromArgs(args)
	isDataInFile := filename != ""
	if isDataInFile {
		data = base64Encode(data)
	}
	if len(data) > MaxPayloadSizeBytes {
		e.outClient.Fail(ErrPayloadTooLarge)
		return 1
	}

	body := encryptionRequest{Path: path, Plaintext: data, Version: viper.GetString(cst.Version)}
	basePath := strings.Join([]string{"crypto/manual", cst.Encrypt}, "/")
	uri := paths.CreateURI(basePath, nil)
	resp, apiError := e.request.DoRequest(http.MethodPost, uri, body)
	if apiError != nil {
		e.outClient.FailE(apiError)
		return 1
	}

	if isDataInFile {
		var newFileName string
		output := viper.GetString(cst.Output)
		if output != "" {
			newFileName = output
		} else {
			info, _ := os.Stat(filename)
			newFileName = info.Name() + ".enc"
		}

		err := ioutil.WriteFile(newFileName, resp, 0664)
		if err != nil {
			e.outClient.Fail(err)
			return 1
		}
		e.outClient.WriteResponse([]byte(fmt.Sprintf("Ciphertext with metadata successfully saved in %s.", newFileName)), nil)
		return 0
	}

	e.outClient.WriteResponse(resp, nil)
	return utils.GetExecStatus(apiError)
}

func (e manualKeyEncryption) handleManualKeyDecrypt(args []string) int {
	if e.outClient == nil {
		e.outClient = format.NewDefaultOutClient()
	}

	data := viper.GetString(cst.Data)
	if data == "" {
		e.outClient.FailF("Please provide a value for %s. Either a string in quotes or a path to a file (@myfile.txt).", cst.Data)
		return 1
	}

	filename := paths.GetFilenameFromArgs(args)
	isDataInFile := filename != ""

	var body decryptionRequest
	if isDataInFile {
		var dr decryptionRequest
		err := json.Unmarshal([]byte(data), &dr)
		if err != nil {
			e.outClient.Fail(err)
			return 1
		}
		body = dr
	} else {
		path := viper.GetString(cst.Path)
		if path == "" {
			e.outClient.FailF("%s is required", cst.Path)
			return 1
		}
		body.Path = path
		body.Ciphertext = data
		body.Version = viper.GetString(cst.Version)
	}

	if len(body.Ciphertext) > MaxPayloadSizeBytes {
		e.outClient.Fail(ErrPayloadTooLarge)
		return 1
	}

	basePath := strings.Join([]string{"crypto/manual", cst.Decrypt}, "/")
	uri := paths.CreateURI(basePath, nil)

	resp, apiError := e.request.DoRequest(http.MethodPost, uri, body)
	if apiError != nil {
		e.outClient.FailE(apiError)
		return 1
	}

	if isDataInFile {
		var dr decryptionResponse
		err := json.Unmarshal(resp, &dr)
		if err != nil {
			e.outClient.Fail(err)
			return 1
		}

		resp, _ = json.Marshal(dr)

		var newFileName string
		output := viper.GetString(cst.Output)
		if output != "" {
			newFileName = output
		} else {
			info, _ := os.Stat(filename)
			newFileName = info.Name() + ".txt"
		}

		err = ioutil.WriteFile(newFileName, resp, 0664)
		if err != nil {
			e.outClient.Fail(err)
			return 1
		}
		e.outClient.WriteResponse([]byte(fmt.Sprintf("Decrypted data with metadata successfully saved in %s.", newFileName)), nil)
		return 0
	}
	e.outClient.WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func (e manualKeyEncryption) makeKeyURL(path string, query map[string]string) (string, *errors.ApiError) {
	return paths.GetResourceURIFromResourcePath("crypto/manual/key", path, "", "", query, false)
}

type manualKeyData struct {
	Scheme     string                 `json:"scheme"`
	PrivateKey string                 `json:"privateKey"`
	Nonce      string                 `json:"nonce"`
	Metadata   map[string]interface{} `json:"metadata"`
}
