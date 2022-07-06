// this refers to remote config
package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	cst "thy/constants"
	"thy/errors"
	"thy/format"
	"thy/internal/predictor"
	"thy/paths"
	"thy/utils"
	"thy/vaultcli"

	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

func GetConfigCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounConfig},
		SynopsisText: "config",
		HelpText: fmt.Sprintf(`Execute an action on the %[2]s in %[3]s

Usage:
   • config %[1]s
   • config %[4]s --data %[5]s
		`, cst.Read, cst.NounConfig, cst.ProductName, cst.Update, cst.ExampleConfigPath),
		RunFunc: func(args []string) int {
			return handleConfigReadCmd(vaultcli.New(), args)
		},
	})
}

func GetConfigReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounConfig, cst.Read},
		SynopsisText: fmt.Sprintf("%s %s", cst.NounConfig, cst.Read),
		HelpText: fmt.Sprintf(`Read the %[2]s from %[3]s
Usage:
   • config %[1]s
		`, cst.Read, cst.NounConfig, cst.ProductName),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Version, Usage: "List the current and last (n) versions"},
		},
		RunFunc: func(args []string) int {
			return handleConfigReadCmd(vaultcli.New(), args)
		},
	})
}

func GetConfigUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounConfig, cst.Update},
		SynopsisText: fmt.Sprintf("%s %s", cst.NounConfig, cst.Update),
		HelpText: fmt.Sprintf(`Update the %[2]s in %[3]s
Usage:
   • config %[1]s --data %[4]s --encoding yml
   • config %[1]s @/tmp/conf.json
   • config %[1]s "{\"tenantName: \"ambarco\"...}\" --encoding json
		`, cst.Update, cst.NounConfig, cst.ProductName, cst.ExampleConfigPath),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("%s to be stored in the %s. Prefix with '@' to denote filepath (required)", strings.Title(cst.Data), cst.Config), Predictor: predictor.NewPrefixFilePredictor("*")},
		},
		MinNumberArgs: 1,
		RunFunc: func(args []string) int {
			return handleConfigUpdateCmd(vaultcli.New(), args)
		},
	})
}

func GetConfigEditCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounConfig, cst.Edit},
		SynopsisText: fmt.Sprintf("%s %s", cst.NounConfig, cst.Edit),
		HelpText: fmt.Sprintf(`Edit the %[1]s in %[2]s
Usage:
   • config edit
   • config edit --encoding yml
		`, cst.NounConfig, cst.ProductName),
		RunFunc: func(args []string) int {
			return handleConfigEditCmd(vaultcli.New(), args)
		},
	})
}

func handleConfigReadCmd(vcli vaultcli.CLI, args []string) int {
	config := "config"
	version := viper.GetString(cst.Version)
	if strings.TrimSpace(version) != "" {
		config = fmt.Sprint(config, "/", cst.Version, "/", version)
	}
	uri := paths.CreateURI(config, nil)
	resp, err := vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)

	vcli.Out().WriteResponse(resp, err)
	return 0
}

func handleConfigUpdateCmd(vcli vaultcli.CLI, args []string) int {
	data := viper.GetString(cst.Data)
	encoding := viper.GetString(cst.Encoding)
	if data == "" {
		vcli.Out().FailF("Please provide --%s or -%s and a value for it", cst.Data, string(cst.Data[0]))
		return 1
	}

	fileName := vaultcli.GetFilenameFromArgs(args)

	if !utf8.Valid([]byte(data)) {
		data = utf16toString([]byte(data))
	}

	uri := paths.CreateURI("config", nil)
	model := PostConfigModel{
		Config:        data,
		Serialization: encoding,
	}
	resp, err := vcli.HTTPClient().DoRequest(http.MethodPut, uri, &model)
	vcli.Out().WriteResponse(resp, err)

	if err == nil && resp != nil && fileName != "" {
		// Overwrite the given file with the most recent version of the config.
		// The version would still be one behind that incremented as a result of this update.
		color := false
		text, _ := format.BeautifyBytes(resp, &color)
		if writeErr := ioutil.WriteFile(fileName, []byte(text), 0644); writeErr != nil {
			vcli.Out().Fail(writeErr)
		}
	}

	return utils.GetExecStatus(err)
}

func handleConfigEditCmd(vcli vaultcli.CLI, args []string) int {
	uri := paths.CreateURI("config", nil)
	resp, err := vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
	if err != nil {
		vcli.Out().WriteResponse(resp, err)
		return utils.GetExecStatus(err)
	}
	saveFunc := func(data []byte) (resp []byte, err *errors.ApiError) {
		encoding := viper.GetString(cst.Encoding)
		model := PostConfigModel{
			Config:        string(data),
			Serialization: encoding,
		}
		_, err = vcli.HTTPClient().DoRequest(http.MethodPut, uri, &model)
		return nil, err
	}
	resp, err = vcli.Edit(resp, saveFunc)
	vcli.Out().WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

type PostConfigModel struct {
	Config        string
	Serialization string
}

//utf16toString decodes utf16 to utf8
func utf16toString(b []uint8) string {
	if len(b)&1 != 0 {
		return ""
	}
	// Check and remove BOM
	var bom int
	if len(b) >= 2 {
		switch n := int(b[0])<<8 | int(b[1]); n {
		case 0xfffe:
			bom = 1
			fallthrough
		case 0xfeff:
			b = b[2:]
		}
	}

	w := make([]uint16, len(b)/2)
	for i := range w {
		w[i] = uint16(b[2*i+bom&1])<<8 | uint16(b[2*i+(bom+1)&1])
	}
	return string(utf16.Decode(w))
}
