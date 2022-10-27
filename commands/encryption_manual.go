package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	cst "thy/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/internal/predictor"
	"github.com/DelineaXPM/dsv-cli/paths"
	"github.com/DelineaXPM/dsv-cli/utils"
	"github.com/DelineaXPM/dsv-cli/vaultcli"

	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

func GetCryptoManualCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Manual},
		SynopsisText: "Encryption-as-a-Service with a Manual Key",
		HelpText:     "Encryption-as-a-Service with a Manual Key",
		NoConfigRead: true,
		NoPreAuth:    true,
	})
}

func GetManualKeyUploadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Manual, cst.NounKey + "-" + cst.Upload},
		SynopsisText: "Upload a new manual encryption/decryption key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s mykeys/key1 --%[5]s %[6]s --%[7]s %[8]s --%[9]s %[10]s
`, cst.NounEncryption, cst.Manual, cst.NounKey+"-"+cst.Upload, cst.Path,
			cst.Scheme, "symmetric", cst.PrivateKey, cst.ExamplePrivateKey, cst.Nonce, cst.ExampleNonce),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey), Predictor: predictor.NewSecretPathPredictorDefault()},
			{Name: cst.PrivateKey, Usage: "Private key base64-encoded (required)"},
			{Name: cst.Scheme, Usage: "Encryption scheme (required)"},
			{Name: cst.Nonce, Usage: "Nonce base64-encoded (optional)"},
			{Name: cst.Metadata, Usage: "Metadata as a JSON object (optional)"},
		},
		MinNumberArgs: 6,
		RunFunc:       handleUploadManualKey,
	})
}

func GetManualKeyUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Manual, cst.NounKey + "-" + cst.Update},
		SynopsisText: "Update an existing encryption key for encryption/decryption",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s mykeys/key1 --%[5]s %[6]s
`, cst.NounEncryption, cst.Manual, cst.NounKey+"-"+cst.Update, cst.Path,
			cst.PrivateKey, cst.ExamplePrivateKey),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey), Predictor: predictor.NewSecretPathPredictorDefault()},
			{Name: cst.PrivateKey, Usage: "Private key base64-encoded (required)"},
			{Name: cst.Nonce, Usage: "Nonce base64-encoded (optional)"},
			{Name: cst.Metadata, Usage: "Metadata as a JSON object (optional)"},
		},
		MinNumberArgs: 4,
		RunFunc:       handleUpdateManualKey,
	})
}

func GetManualKeyReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Manual, cst.NounKey + "/" + cst.Read},
		SynopsisText: "Read an existing manual encryption key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s mykeys/key1
`, cst.NounEncryption, cst.NounKey, cst.Read, cst.Path),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey), Predictor: predictor.NewSecretPathPredictorDefault()},
		},
		MinNumberArgs: 1,
		RunFunc:       handleReadManualKey,
	})
}

func GetManualKeyDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Manual, cst.NounKey + "/" + cst.Delete},
		SynopsisText: "Delete an existing manual encryption key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s mykeys/key1
`, cst.NounEncryption, cst.Manual, cst.Delete, cst.Path),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey), Predictor: predictor.NewSecretPathPredictorDefault()},
			{Name: cst.Force, Usage: fmt.Sprintf("Immediately delete %s and all its versions", cst.NounKey), ValueType: "bool"},
		},
		MinNumberArgs: 1,
		RunFunc:       handleDeleteManualKey,
	})
}

func GetManualKeyRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.NounKey + "/" + cst.Restore},
		SynopsisText: "Restore a previously soft-deleted manual encryption key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s mykeys/key1
`, cst.NounEncryption, cst.NounKey, cst.Restore, cst.Path),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey), Predictor: predictor.NewSecretPathPredictorDefault()},
		},
		MinNumberArgs: 1,
		RunFunc:       handleRestoreManualKey,
	})
}

func GetManualKeyEncryptCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Encrypt},
		SynopsisText: "Encrypt data using a manual key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s 'hello world' --%[5]s mykeys/key1
   • %[1]s %[2]s %[3]s --%[4]s @mysecret.txt --%[5]s mykeys/key1
`, cst.NounEncryption, cst.Manual, cst.Encrypt, cst.Data, cst.Path),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey), Predictor: predictor.NewSecretPathPredictorDefault()},
			{Name: cst.Version, Usage: fmt.Sprintf("Version of the %s used for encryption/decryption", cst.NounKey)},
			{Name: cst.Data, Shorthand: "d", Usage: "A plaintext string or path to a @file with data to be encrypted", Predictor: predictor.NewPrefixFilePredictor("*")},
			{Name: cst.Output, Usage: "Output file for encrypted value and metadata"},
		},
		MinNumberArgs: 4,
		RunFunc:       handleManualKeyEncrypt,
	})
}

func GetManualKeyDecryptCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Decrypt},
		SynopsisText: "Decrypt data using a manual key that had performed encryption",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s 'hello world' --%[5]s mykeys/key1
   • %[1]s %[2]s %[3]s --%[4]s @mysecret.txt"
`, cst.NounEncryption, cst.Manual, cst.Decrypt, cst.Data, cst.Path, cst.Version),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s", cst.Path, cst.NounKey), Predictor: predictor.NewSecretPathPredictorDefault()},
			{Name: cst.Version, Usage: fmt.Sprintf("Version of the %s used for encryption/decryption", cst.NounKey)},
			{Name: cst.Data, Shorthand: "d", Usage: "A ciphertext string or path to a @file with data to be decrypted", Predictor: predictor.NewPrefixFilePredictor("*")},
			{Name: cst.Output, Usage: "Output file for decrypted value and metadata"},
		},
		MinNumberArgs: 2,
		RunFunc:       handleManualKeyDecrypt,
	})
}

func handleUploadManualKey(vcli vaultcli.CLI, args []string) int {
	path := viper.GetString(cst.Path)
	if path == "" {
		vcli.Out().FailF("%s is required", cst.Path)
		return 1
	}

	privateKey := viper.GetString(cst.PrivateKey)
	if privateKey == "" {
		vcli.Out().FailF("%s is required", cst.PrivateKey)
		return 1
	}

	scheme := viper.GetString(cst.Scheme)
	if scheme == "" {
		vcli.Out().FailF("%s is required", cst.Scheme)
		return 1
	}

	nonce := viper.GetString(cst.Nonce)
	metadata := viper.GetString(cst.Metadata)

	uri, err := makeManualKeyURL(path, nil)
	if err != nil {
		vcli.Out().Fail(err)
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
			vcli.Out().FailS("failed to parse the metadata")
		} else {
			body.Metadata = metadataMap
		}
	}

	resp, apiError := vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func handleUpdateManualKey(vcli vaultcli.CLI, args []string) int {
	path := viper.GetString(cst.Path)
	if path == "" {
		vcli.Out().FailF("%s is required", cst.Path)
		return 1
	}

	privateKey := viper.GetString(cst.PrivateKey)
	if privateKey == "" {
		vcli.Out().FailF("%s is required", cst.PrivateKey)
		return 1
	}

	nonce := viper.GetString(cst.Nonce)
	metadata := viper.GetString(cst.Metadata)

	uri, err := makeManualKeyURL(path, nil)
	if err != nil {
		vcli.Out().Fail(err)
		return utils.GetExecStatus(err)
	}

	body := manualKeyData{
		PrivateKey: privateKey,
		Nonce:      nonce,
	}

	if metadata != "" {
		var metadataMap map[string]interface{}
		if err := json.Unmarshal([]byte(metadata), &metadataMap); err != nil {
			vcli.Out().FailS("failed to parse the metadata")
		} else {
			body.Metadata = metadataMap
		}
	}

	resp, apiError := vcli.HTTPClient().DoRequest(http.MethodPut, uri, body)
	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func handleReadManualKey(vcli vaultcli.CLI, args []string) int {
	path := viper.GetString(cst.Path)
	if path == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		path = args[0]
	}

	uri, err := makeManualKeyURL(path, nil)
	if err != nil {
		vcli.Out().Fail(err)
		return utils.GetExecStatus(err)
	}

	resp, apiError := vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func handleDeleteManualKey(vcli vaultcli.CLI, args []string) int {
	path := viper.GetString(cst.Path)
	if path == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		path = args[0]
	}

	force := viper.GetBool(cst.Force)
	query := map[string]string{"force": strconv.FormatBool(force)}
	uri, err := makeManualKeyURL(path, query)
	if err != nil {
		vcli.Out().Fail(err)
		return utils.GetExecStatus(err)
	}

	resp, apiError := vcli.HTTPClient().DoRequest(http.MethodDelete, uri, nil)
	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func handleRestoreManualKey(vcli vaultcli.CLI, args []string) int {
	path := viper.GetString(cst.Path)
	if path == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		path = args[0]
	}

	uri, err := makeManualKeyURL(path, nil)
	if err != nil {
		vcli.Out().Fail(err)
		return utils.GetExecStatus(err)
	}

	resp, apiError := vcli.HTTPClient().DoRequest(http.MethodPut, uri+"/restore", nil)
	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func handleManualKeyEncrypt(vcli vaultcli.CLI, args []string) int {
	path := viper.GetString(cst.Path)
	if path == "" {
		vcli.Out().FailF("%s is required", cst.Path)
		return 1
	}

	data := viper.GetString(cst.Data)
	if data == "" {
		vcli.Out().FailF("Please provide a value for %s. Either a string in quotes or a path to a file (@myfile.txt).", cst.Data)
		return 1
	}

	filename := vaultcli.GetFilenameFromArgs(args)
	isDataInFile := filename != ""
	if isDataInFile {
		data = base64Encode(data)
	}
	if len(data) > MaxPayloadSizeBytes {
		vcli.Out().Fail(ErrPayloadTooLarge)
		return 1
	}

	body := encryptionRequest{Path: path, Plaintext: data, Version: viper.GetString(cst.Version)}
	basePath := strings.Join([]string{"crypto/manual", cst.Encrypt}, "/")
	uri := paths.CreateURI(basePath, nil)
	resp, apiError := vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
	if apiError != nil {
		vcli.Out().FailE(apiError)
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

		err := os.WriteFile(newFileName, resp, 0664)
		if err != nil {
			vcli.Out().Fail(err)
			return 1
		}
		vcli.Out().WriteResponse([]byte(fmt.Sprintf("Ciphertext with metadata successfully saved in %s.", newFileName)), nil)
		return 0
	}

	vcli.Out().WriteResponse(resp, nil)
	return utils.GetExecStatus(apiError)
}

func handleManualKeyDecrypt(vcli vaultcli.CLI, args []string) int {
	data := viper.GetString(cst.Data)
	if data == "" {
		vcli.Out().FailF("Please provide a value for %s. Either a string in quotes or a path to a file (@myfile.txt).", cst.Data)
		return 1
	}

	filename := vaultcli.GetFilenameFromArgs(args)
	isDataInFile := filename != ""

	var body decryptionRequest
	if isDataInFile {
		var dr decryptionRequest
		err := json.Unmarshal([]byte(data), &dr)
		if err != nil {
			vcli.Out().Fail(err)
			return 1
		}
		body = dr
	} else {
		path := viper.GetString(cst.Path)
		if path == "" {
			vcli.Out().FailF("%s is required", cst.Path)
			return 1
		}
		body.Path = path
		body.Ciphertext = data
		body.Version = viper.GetString(cst.Version)
	}

	if len(body.Ciphertext) > MaxPayloadSizeBytes {
		vcli.Out().Fail(ErrPayloadTooLarge)
		return 1
	}

	basePath := strings.Join([]string{"crypto/manual", cst.Decrypt}, "/")
	uri := paths.CreateURI(basePath, nil)

	resp, apiError := vcli.HTTPClient().DoRequest(http.MethodPost, uri, body)
	if apiError != nil {
		vcli.Out().FailE(apiError)
		return 1
	}

	if isDataInFile {
		var dr decryptionResponse
		err := json.Unmarshal(resp, &dr)
		if err != nil {
			vcli.Out().Fail(err)
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

		err = os.WriteFile(newFileName, resp, 0664)
		if err != nil {
			vcli.Out().Fail(err)
			return 1
		}
		vcli.Out().WriteResponse([]byte(fmt.Sprintf("Decrypted data with metadata successfully saved in %s.", newFileName)), nil)
		return 0
	}
	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func makeManualKeyURL(path string, query map[string]string) (string, *errors.ApiError) {
	return paths.GetResourceURIFromResourcePath("crypto/manual/key", path, "", "", query)
}

type manualKeyData struct {
	Scheme     string                 `json:"scheme"`
	PrivateKey string                 `json:"privateKey"`
	Nonce      string                 `json:"nonce"`
	Metadata   map[string]interface{} `json:"metadata"`
}
