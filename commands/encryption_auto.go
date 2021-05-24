package cmd

import (
	"encoding/json"
	goerrors "errors"
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

const MaxPayloadSizeBytes = 2097152

var ErrPayloadTooLarge = goerrors.New(fmt.Sprintf("payload is too large, maximum size is %dMB", MaxPayloadSizeBytes/1000000))

type encryption struct {
	request   requests.Client
	outClient format.OutClient
}

func GetCryptoCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounEncryption},
		RunFunc: func(args []string) int {
			return cli.RunResultHelp
		},
		SynopsisText:  "Encryption-as-a-Service",
		HelpText:      "Encryption-as-a-Service",
		MinNumberArgs: 0,
	})
}

func GetAutoKeyCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.NounKey + "/" + cst.Create},
		RunFunc:      encryption{requests.NewHttpClient(), nil}.handleCreateAutoKey,
		SynopsisText: "Create a new auto key for encryption/decryption",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s mykeys/key1
		`, cst.NounEncryption, cst.NounKey, cst.Create, cst.Path),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path): cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey)}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetEncryptionRotateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Rotate},
		RunFunc:      encryption{requests.NewHttpClient(), nil}.handleRotate,
		SynopsisText: "Rotate existing data with a later or new version of the key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s "mykeys/key1 --%[4]s '$fh9d87g' --%[5]s 4"
   • %[1]s %[2]s --%[3]s "mykeys/key1 --%[4]s @cipher.enc --%[5]s 0 --%[6]s 3"
		`, cst.NounEncryption, cst.Rotate, cst.Path, cst.Data, cst.VersionStart, cst.VersionEnd),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):         cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey)}), false},
			preds.LongFlag(cst.Data):         cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: "Ciphertext to be re-encrypted. Pass a string literal in quotes or specify a filepath prefixed with '@' (required)"}), false},
			preds.LongFlag(cst.VersionStart): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.VersionStart, Shorthand: "", Usage: "Starting version of the auto key (required)"}), false},
			preds.LongFlag(cst.VersionEnd):   cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.VersionEnd, Shorthand: "", Usage: "Ending version of the auto key"}), false},
			preds.LongFlag(cst.Output):       cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Output, Shorthand: "", Usage: "Output file for encrypted value and metadata"}), false},
		},
		MinNumberArgs: 5,
	})
}

func GetAutoKeyReadMetadataCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.NounKey + "/" + cst.Read},
		RunFunc:      encryption{requests.NewHttpClient(), nil}.handleReadAutoKey,
		SynopsisText: "Read metadata of an existing auto key",
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

func GetAutoKeyDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.NounKey + "/" + cst.Delete},
		RunFunc:      encryption{requests.NewHttpClient(), nil}.handleDeleteAutoKey,
		SynopsisText: "Delete an existingn auto key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s mykeys/key1
		`, cst.NounEncryption, cst.NounKey, cst.Delete, cst.Path),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):  cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey)}), false},
			preds.LongFlag(cst.Force): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Force, Shorthand: "", Usage: fmt.Sprintf("Immediately delete %s and all its versions", cst.NounKey), Global: false, ValueType: "bool"}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetAutoKeyRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.NounKey + "/" + cst.Restore},
		RunFunc:      encryption{requests.NewHttpClient(), nil}.handleRestoreAutoKey,
		SynopsisText: "Restore a previously soft-deletedn auto key",
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

func GetEncryptCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Encrypt},
		RunFunc:      encryption{requests.NewHttpClient(), nil}.handleEncrypt,
		SynopsisText: "Encrypt data using an auto key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s 'hello world' --%[4]s mykeys/key1
   • %[1]s %[2]s --%[3]s @mysecret.txt --%[4]s mykeys/key1
		`, cst.NounEncryption, cst.Encrypt, cst.Data, cst.Path),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):    cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey)}), false},
			preds.LongFlag(cst.Version): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Version, Shorthand: "", Usage: fmt.Sprintf("Version of the %s used for encryption/decryption", cst.NounKey)}), false},
			preds.LongFlag(cst.Data):    cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: "A plaintext string or path to a @file with data to be encrypted"}), false},
			preds.LongFlag(cst.Output):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Output, Shorthand: "", Usage: "Output file for encrypted value and metadata"}), false},
		},
		MinNumberArgs: 4,
	})
}

func GetDecryptCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Decrypt},
		RunFunc:      encryption{requests.NewHttpClient(), nil}.handleDecrypt,
		SynopsisText: "Decrypt data using an auto key that had performed encryption",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s 'hello world' --%[4]s mykeys/key1
   • %[1]s %[2]s --%[3]s @mysecret.txt"
		`, cst.NounEncryption, cst.Decrypt, cst.Data, cst.Path, cst.Version),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):    cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s", cst.Path, cst.NounKey)}), false},
			preds.LongFlag(cst.Version): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Version, Shorthand: "", Usage: fmt.Sprintf("Version of the %s used for encryption/decryption", cst.NounKey)}), false},
			preds.LongFlag(cst.Data):    cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: "A ciphertext string or path to a @file with data to be decrypted"}), false},
			preds.LongFlag(cst.Output):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Output, Shorthand: "", Usage: "Output file for decrypted value and metadata"}), false},
		},
		MinNumberArgs: 2,
	})
}

func (e encryption) handleCreateAutoKey(args []string) int {
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

	resp, apiError := e.request.DoRequest(http.MethodPost, uri, nil)
	e.outClient.WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func (e encryption) handleReadAutoKey(args []string) int {
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

func (e encryption) handleDeleteAutoKey(args []string) int {
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

func (e encryption) handleRestoreAutoKey(args []string) int {
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

func (e encryption) handleRotate(args []string) int {
	if e.outClient == nil {
		e.outClient = format.NewDefaultOutClient()
	}

	path := viper.GetString(cst.Path)
	if path == "" {
		e.outClient.FailF("%s is required", cst.Path)
		return 1
	}

	versionStart := viper.GetString(cst.VersionStart)
	if versionStart == "" {
		e.outClient.FailF("%s is required", cst.VersionStart)
		return 1
	}

	data := viper.GetString(cst.Data)
	if data == "" {
		e.outClient.FailF("Please provide a value for %s. Either a string in quotes or a path to a file (@myfile.txt).", cst.Data)
		return 1
	}

	filename := paths.GetFilenameFromArgs(args)
	isDataInFile := filename != ""
	var body rotationRequest
	if isDataInFile {
		var dr rotationRequest
		err := json.Unmarshal([]byte(data), &dr)
		if err != nil {
			e.outClient.Fail(err)
			return 1
		}
		body = dr
	} else {
		body.Ciphertext = data
		body.Path = path
	}

	if len(body.Ciphertext) > MaxPayloadSizeBytes {
		e.outClient.Fail(ErrPayloadTooLarge)
		return 1
	}

	body.StartingVersion = versionStart
	body.EndingVersion = viper.GetString(cst.VersionEnd)

	basePath := strings.Join([]string{"crypto", cst.Rotate}, "/")
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
			newFileName = info.Name()
		}

		err := ioutil.WriteFile(newFileName, resp, 0664)
		if err != nil {
			e.outClient.Fail(err)
			return 1
		}
		e.outClient.WriteResponse([]byte(fmt.Sprintf("Re-encrypted data with metadata successfully saved in %s.", newFileName)), nil)
		return 0
	}

	e.outClient.WriteResponse(resp, nil)
	return utils.GetExecStatus(apiError)
}

func (e encryption) handleEncrypt(args []string) int {
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
	basePath := strings.Join([]string{"crypto", cst.Encrypt}, "/")
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

func (e encryption) handleDecrypt(args []string) int {
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

	basePath := strings.Join([]string{"crypto", cst.Decrypt}, "/")
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

func (e encryption) makeKeyURL(path string, query map[string]string) (string, *errors.ApiError) {
	return paths.GetResourceURIFromResourcePath("crypto/key", path, "", "", true, query, false)
}

type rotationRequest struct {
	Path            string `json:"path"`
	Ciphertext      string `json:"ciphertext"`
	StartingVersion string `json:"startingVersion"`
	EndingVersion   string `json:"endingVersion"`
}

type encryptionRequest struct {
	Path      string `json:"path"`
	Plaintext string `json:"plaintext"`
	Version   string `json:"version"`
}

type decryptionRequest struct {
	Path       string `json:"path"`
	Ciphertext string `json:"ciphertext"`
	Version    string `json:"version"`
}

type decryptionResponse struct {
	Path    string `json:"path"`
	Data    string `json:"data"`
	Version string `json:"version"`
}
