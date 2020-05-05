// this refers to remote config
package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	cst "thy/constants"
	"thy/errors"
	"thy/format"
	preds "thy/predictors"
	"thy/requests"
	"thy/store"
	"thy/utils"

	"github.com/posener/complete"

	"github.com/thycotic-rd/cli"
	"github.com/thycotic-rd/viper"
)

type Config struct {
	request   requests.Client
	outClient format.OutClient
	getStore  func(stString string) (store.Store, *errors.ApiError)
	edit      func([]byte, dataFunc, *errors.ApiError, bool) ([]byte, *errors.ApiError)
}

func GetConfigCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounConfig},
		RunFunc: func(args []string) int {
			id := viper.GetString(cst.ID)
			path := viper.GetString(cst.Path)
			if path == "" {
				path = utils.GetPath(args)
			}
			if path == "" && id == "" {
				return cli.RunResultHelp
			}
			config := Config{requests.NewHttpClient(),
				nil,
				store.GetStore,
				EditData}
			return config.handleConfigReadCmd(args)
		},
		SynopsisText: "config",
		HelpText: fmt.Sprintf(`Execute an action on the %[2]s in %[3]s

Usage:
   • config %[1]s
   • config %[4]s --data %[5]s
		`, cst.Read, cst.NounConfig, cst.ProductName, cst.Update, cst.ExampleConfigPath),
		MinNumberArgs: 1,
	})
}

func GetConfigReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounConfig, cst.Read},
		RunFunc: Config{
			requests.NewHttpClient(),
			nil,
			store.GetStore,
			EditData}.handleConfigReadCmd,
		SynopsisText: fmt.Sprintf("%s %s", cst.NounConfig, cst.Read),
		HelpText: fmt.Sprintf(`Read the %[2]s from %[3]s
Usage:
	• config %[1]s -b
				`, cst.Read, cst.NounConfig, cst.ProductName),
		FlagsPredictor: cli.PredictorWrappers{preds.LongFlag(cst.Version): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Version, Shorthand: "", Usage: "List the current and last (n) versions"}), false}},
		MinNumberArgs:  0,
	})
}

func GetConfigUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounConfig, cst.Update},
		RunFunc: Config{
			requests.NewHttpClient(),
			nil,
			store.GetStore,
			EditData}.handleConfigUpdateCmd,
		SynopsisText: fmt.Sprintf("%s %s", cst.NounConfig, cst.Update),
		HelpText: fmt.Sprintf(`Update the %[2]s in %[3]s
Usage:
	• config %[1]s --data %[4]s --encoding yml
	• config %[1]s @/tmp/conf.json
	• config %[1]s "{\"tenantName: \"ambarco\"...}\" --encoding json
				`, cst.Update, cst.NounConfig, cst.ProductName, cst.ExampleConfigPath),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Data): cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("%s to be stored in the %s. Prefix with '@' to denote filepath (required)", strings.Title(cst.Data), cst.Config)}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetConfigEditCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounConfig, cst.Edit},
		RunFunc: Config{
			requests.NewHttpClient(),
			nil,
			store.GetStore,
			EditData}.handleConfigEditCmd,
		SynopsisText: fmt.Sprintf("%s %s", cst.NounConfig, cst.Edit),
		HelpText: fmt.Sprintf(`Edit the %[2]s in %[3]s
Usage:
	• config %[1]s --data %[4]s --encoding yml
	• config %[1]s @/tmp/conf.json
	• config %[1]s "{\"tenantName: \"ambarco\"...}\" --encoding json
				`, cst.Edit, cst.NounConfig, cst.ProductName, cst.ExampleConfigPath),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Data): cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("%s to be stored in the %s. Prefix with '@' to denote filepath (required)", strings.Title(cst.Data), cst.Config)}), false},
		},
		MinNumberArgs: 0,
	})
}

func (c Config) handleConfigReadCmd(args []string) int {
	config := "config"
	version := viper.GetString(cst.Version)
	if strings.TrimSpace(version) != "" {
		config = fmt.Sprint(config, "/", cst.Version, "/", version)
	}
	uri := utils.CreateURI(config, nil)
	resp, err := c.request.DoRequest("GET", uri, nil)

	outClient := c.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}

	outClient.WriteResponse(resp, err)
	return 0
}

func (c Config) handleConfigUpdateCmd(args []string) int {
	var err *errors.ApiError
	var resp []byte
	uri := utils.CreateURI("config", nil)
	data := viper.GetString(cst.Data)
	encoding := viper.GetString(cst.Encoding)
	var fileName string
	if c.outClient == nil {
		c.outClient = format.NewDefaultOutClient()
	}
	if data == "" {
		c.outClient.FailF("Please provide --%s or -%s and a value for it", cst.Data, string(cst.Data[0]))
		return 1
	}

	if f := utils.GetFilenameFromArgs(args); f != "" {
		fileName = f
	}

	if !utf8.Valid([]byte(data)) {
		data = utf16toString([]byte(data))
	}
	model := PostConfigModel{
		Config:        data,
		Serialization: encoding,
	}
	resp, err = c.request.DoRequest("PUT", uri, &model)
	c.outClient.WriteResponse(resp, err)

	if err == nil && resp != nil && fileName != "" {
		// Overwrite the given file with the most recent version of the config.
		// The version would still be one behind that incremented as a result of this update.
		color := false
		text, _ := format.BeautifyBytes(resp, &color)
		if writeErr := ioutil.WriteFile(fileName, []byte(text), 0644); writeErr != nil {
			c.outClient.Fail(writeErr)
		}
	}

	return utils.GetExecStatus(err)
}

func (c Config) handleConfigEditCmd(args []string) int {
	if c.outClient == nil {
		c.outClient = format.NewDefaultOutClient()
	}

	var err *errors.ApiError
	var resp []byte
	uri := utils.CreateURI("config", nil)
	resp, err = c.request.DoRequest("GET", uri, nil)
	if err != nil {
		c.outClient.WriteResponse(resp, err)
		return utils.GetExecStatus(err)
	}
	saveFunc := dataFunc(func(data []byte) (resp []byte, err *errors.ApiError) {
		encoding := viper.GetString(cst.Encoding)
		model := PostConfigModel{
			Config:        string(data),
			Serialization: encoding,
		}
		_, err = c.request.DoRequest("PUT", uri, &model)
		return nil, err
	})
	resp, err = c.edit(resp, saveFunc, nil, false)
	c.outClient.WriteResponse(resp, err)
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
