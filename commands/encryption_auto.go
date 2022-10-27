package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/internal/predictor"
	"github.com/DelineaXPM/dsv-cli/paths"
	"github.com/DelineaXPM/dsv-cli/utils"
	"github.com/DelineaXPM/dsv-cli/vaultcli"

	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

const MaxPayloadSizeBytes = 2097152

var ErrPayloadTooLarge = fmt.Errorf("payload is too large, maximum size is %dMB", MaxPayloadSizeBytes/1000000)

func GetCryptoCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption},
		SynopsisText: "Encryption-as-a-Service",
		HelpText:     "Encryption-as-a-Service",
		NoConfigRead: true,
		NoPreAuth:    true,
	})
}

func GetAutoKeyCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.NounKey + "/" + cst.Create},
		SynopsisText: "Create a new auto key for encryption/decryption",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s mykeys/key1
`, cst.NounEncryption, cst.NounKey, cst.Create, cst.Path),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey), Predictor: predictor.NewSecretPathPredictorDefault()},
		},
		MinNumberArgs: 1,
		RunFunc:       handleCreateAutoKey,
	})
}

func GetEncryptionRotateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Rotate},
		SynopsisText: "Rotate existing data with a later or new version of the key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s "mykeys/key1 --%[4]s '$fh9d87g' --%[5]s 4"
   • %[1]s %[2]s --%[3]s "mykeys/key1 --%[4]s @cipher.enc --%[5]s 0 --%[6]s 3"
`, cst.NounEncryption, cst.Rotate, cst.Path, cst.Data, cst.VersionStart, cst.VersionEnd),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey), Predictor: predictor.NewSecretPathPredictorDefault()},
			{Name: cst.Data, Shorthand: "d", Usage: "Ciphertext to be re-encrypted. Pass a string literal in quotes or specify a filepath prefixed with '@' (required)", Predictor: predictor.NewPrefixFilePredictor("*")},
			{Name: cst.VersionStart, Usage: "Starting version of the auto key (required)"},
			{Name: cst.VersionEnd, Usage: "Ending version of the auto key"},
			{Name: cst.Output, Usage: "Output file for encrypted value and metadata"},
		},
		MinNumberArgs: 5,
		RunFunc:       handleRotate,
	})
}

func GetAutoKeyReadMetadataCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.NounKey + "/" + cst.Read},
		SynopsisText: "Read metadata of an existing auto key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s mykeys/key1
`, cst.NounEncryption, cst.NounKey, cst.Read, cst.Path),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey), Predictor: predictor.NewSecretPathPredictorDefault()},
		},
		MinNumberArgs: 1,
		RunFunc:       handleReadAutoKey,
	})
}

func GetAutoKeyDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.NounKey + "/" + cst.Delete},
		SynopsisText: "Delete an existingn auto key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s mykeys/key1
`, cst.NounEncryption, cst.NounKey, cst.Delete, cst.Path),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey), Predictor: predictor.NewSecretPathPredictorDefault()},
			{Name: cst.Force, Usage: fmt.Sprintf("Immediately delete %s and all its versions", cst.NounKey), ValueType: "bool"},
		},
		MinNumberArgs: 1,
		RunFunc:       handleDeleteAutoKey,
	})
}

func GetAutoKeyRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.NounKey + "/" + cst.Restore},
		SynopsisText: "Restore a previously soft-deletedn auto key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s --%[4]s mykeys/key1
`, cst.NounEncryption, cst.NounKey, cst.Restore, cst.Path),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey), Predictor: predictor.NewSecretPathPredictorDefault()},
		},
		MinNumberArgs: 1,
		RunFunc:       handleRestoreAutoKey,
	})
}

func GetEncryptCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Encrypt},
		SynopsisText: "Encrypt data using an auto key",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s 'hello world' --%[4]s mykeys/key1
   • %[1]s %[2]s --%[3]s @mysecret.txt --%[4]s mykeys/key1
`, cst.NounEncryption, cst.Encrypt, cst.Data, cst.Path),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounKey), Predictor: predictor.NewSecretPathPredictorDefault()},
			{Name: cst.Version, Usage: fmt.Sprintf("Version of the %s used for encryption/decryption", cst.NounKey)},
			{Name: cst.Data, Shorthand: "d", Usage: "A plaintext string or path to a @file with data to be encrypted", Predictor: predictor.NewPrefixFilePredictor("*")},
			{Name: cst.Output, Usage: "Output file for encrypted value and metadata"},
		},
		MinNumberArgs: 4,
		RunFunc:       handleEncrypt,
	})
}

func GetDecryptCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounEncryption, cst.Decrypt},
		SynopsisText: "Decrypt data using an auto key that had performed encryption",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s --%[3]s 'hello world' --%[4]s mykeys/key1
   • %[1]s %[2]s --%[3]s @mysecret.txt"
`, cst.NounEncryption, cst.Decrypt, cst.Data, cst.Path, cst.Version),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s", cst.Path, cst.NounKey), Predictor: predictor.NewSecretPathPredictorDefault()},
			{Name: cst.Version, Usage: fmt.Sprintf("Version of the %s used for encryption/decryption", cst.NounKey)},
			{Name: cst.Data, Shorthand: "d", Usage: "A ciphertext string or path to a @file with data to be decrypted", Predictor: predictor.NewPrefixFilePredictor("*")},
			{Name: cst.Output, Usage: "Output file for decrypted value and metadata"},
		},
		MinNumberArgs: 2,
		RunFunc:       handleDecrypt,
	})
}

func handleCreateAutoKey(vcli vaultcli.CLI, args []string) int {
	path := viper.GetString(cst.Path)
	if path == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		path = args[0]
	}
	uri, err := makeKeyURL(path, nil)
	if err != nil {
		vcli.Out().Fail(err)
		return utils.GetExecStatus(err)
	}

	resp, apiError := vcli.HTTPClient().DoRequest(http.MethodPost, uri, nil)
	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func handleReadAutoKey(vcli vaultcli.CLI, args []string) int {
	path := viper.GetString(cst.Path)
	if path == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		path = args[0]
	}

	uri, err := makeKeyURL(path, nil)
	if err != nil {
		vcli.Out().Fail(err)
		return utils.GetExecStatus(err)
	}

	resp, apiError := vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func handleDeleteAutoKey(vcli vaultcli.CLI, args []string) int {
	path := viper.GetString(cst.Path)
	if path == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		path = args[0]
	}

	force := viper.GetBool(cst.Force)
	query := map[string]string{"force": strconv.FormatBool(force)}
	uri, err := makeKeyURL(path, query)
	if err != nil {
		vcli.Out().Fail(err)
		return utils.GetExecStatus(err)
	}

	resp, apiError := vcli.HTTPClient().DoRequest(http.MethodDelete, uri, nil)
	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func handleRestoreAutoKey(vcli vaultcli.CLI, args []string) int {
	path := viper.GetString(cst.Path)
	if path == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		path = args[0]
	}

	uri, err := makeKeyURL(path, nil)
	if err != nil {
		vcli.Out().Fail(err)
		return utils.GetExecStatus(err)
	}

	resp, apiError := vcli.HTTPClient().DoRequest(http.MethodPut, uri+"/restore", nil)
	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func handleRotate(vcli vaultcli.CLI, args []string) int {
	path := viper.GetString(cst.Path)
	if path == "" {
		vcli.Out().FailF("%s is required", cst.Path)
		return 1
	}

	versionStart := viper.GetString(cst.VersionStart)
	if versionStart == "" {
		vcli.Out().FailF("%s is required", cst.VersionStart)
		return 1
	}

	data := viper.GetString(cst.Data)
	if data == "" {
		vcli.Out().FailF("Please provide a value for %s. Either a string in quotes or a path to a file (@myfile.txt).", cst.Data)
		return 1
	}

	filename := vaultcli.GetFilenameFromArgs(args)
	isDataInFile := filename != ""
	var body rotationRequest
	if isDataInFile {
		var dr rotationRequest
		err := json.Unmarshal([]byte(data), &dr)
		if err != nil {
			vcli.Out().Fail(err)
			return 1
		}
		body = dr
	} else {
		body.Ciphertext = data
		body.Path = path
	}

	if len(body.Ciphertext) > MaxPayloadSizeBytes {
		vcli.Out().Fail(ErrPayloadTooLarge)
		return 1
	}

	body.StartingVersion = versionStart
	body.EndingVersion = viper.GetString(cst.VersionEnd)

	basePath := strings.Join([]string{"crypto", cst.Rotate}, "/")
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
			newFileName = info.Name()
		}

		err := os.WriteFile(newFileName, resp, 0664)
		if err != nil {
			vcli.Out().Fail(err)
			return 1
		}
		vcli.Out().WriteResponse([]byte(fmt.Sprintf("Re-encrypted data with metadata successfully saved in %s.", newFileName)), nil)
		return 0
	}

	vcli.Out().WriteResponse(resp, nil)
	return utils.GetExecStatus(apiError)
}

func handleEncrypt(vcli vaultcli.CLI, args []string) int {
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
	basePath := strings.Join([]string{"crypto", cst.Encrypt}, "/")
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

func handleDecrypt(vcli vaultcli.CLI, args []string) int {
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

	basePath := strings.Join([]string{"crypto", cst.Decrypt}, "/")
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

func makeKeyURL(path string, query map[string]string) (string, *errors.ApiError) {
	return paths.GetResourceURIFromResourcePath("crypto/key", path, "", "", query)
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
