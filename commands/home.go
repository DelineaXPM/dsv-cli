package cmd

import (
	"fmt"

	cst "thy/constants"
	"thy/paths"
	preds "thy/predictors"
	"thy/requests"
	"thy/store"

	"github.com/posener/complete"
	"github.com/spf13/viper"
	"github.com/thycotic-rd/cli"
)

type home struct {
	Secret
}

func GetHomeCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounHome},
		RunFunc: func(args []string) int {
			path := viper.GetString(cst.Path)
			if path == "" {
				path = paths.GetPath(args)
			}
			if path == "" {
				return cli.RunResultHelp
			}
			return home{Secret{
				requests.NewHttpClient(),
				nil,
				store.GetStore, nil, cst.NounHome}}.handleHomeRead(args)
		},
		SynopsisText: "home (<path> | --path|-r)",
		HelpText: fmt.Sprintf(`Work with secrets in a personal user space

Usage:
   • home %[3]s
   • home --path %[3]s
		`, cst.NounHome, cst.ProductName, cst.ExamplePath),
		FlagsPredictor: GetNoDataOpWrappers(cst.NounSecret),
		MinNumberArgs:  1,
	})
}

func GetHomeReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounHome, cst.Read},
		RunFunc: home{Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil, cst.NounHome}}.handleHomeRead,
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounHome, cst.Read),
		HelpText: fmt.Sprintf(`Read a a secret in %[2]s 
Usage:
	• home %[1]s %[4]s 
	• home %[1]s --path %[4]s`, cst.Read, cst.NounHome, cst.ProductName, cst.ExamplePath),
		FlagsPredictor:    GetNoDataOpWrappers(cst.NounSecret),
		ArgsPredictorFunc: preds.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
	})
}

func GetHomeCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounHome, cst.Create},
		RunFunc: home{Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil, cst.NounHome}}.handleHomeCreate,
		SynopsisText: "Create a secret in home",
		HelpText: fmt.Sprintf(`Create a secret in %[2]s
Usage:
	• %[2]s %[1]s %[4]s %[5]s
	• %[2]s %[1]s --path %[4]s --data %[5]s
	• %[2]s %[1]s --path %[4]s --data %[6]s
				`, cst.Create, cst.NounHome, cst.ProductName, cst.ExamplePath, cst.ExampleDataJSON, cst.ExampleDataPath),
		FlagsPredictor: GetDataOpWrappers(cst.NounSecret),
		MinNumberArgs:  2,
	})
}

func GetHomeDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounHome, cst.Delete},
		RunFunc: home{Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil, cst.NounHome}}.handleHomeDelete,
		SynopsisText: "Delete a secret from home",
		HelpText: fmt.Sprintf(`
Usage:
• %[1]s %[2]s %[4]s
• %[1]s %[2]s --%[3]s %[4]s
`, cst.NounHome, cst.Delete, cst.Path, cst.ExamplePath),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):  cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounSecret)}), false},
			preds.LongFlag(cst.Force): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Force, Shorthand: "", Usage: fmt.Sprintf("Immediately delete %s", cst.NounSecret), Global: false, ValueType: "bool"}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetHomeRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounHome, cst.Restore},
		RunFunc: home{Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil, cst.NounHome}}.handleHomeRestore,
		SynopsisText: fmt.Sprintf("Restore a soft-deleted secret in %s", cst.NounHome),
		HelpText: fmt.Sprintf(`
Usage:
	• %[1]s %[2]s %[3]s

				`, cst.NounHome, cst.Restore, cst.ExamplePath),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path): cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounHome)}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetHomeUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounHome, cst.Update},
		RunFunc: home{Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil, cst.NounHome}}.handleHomeUpdate,
		SynopsisText: "Update a secret in home",
		HelpText: fmt.Sprintf(`Create a secret in %[2]s
Usage:
	• %[2]s %[1]s %[4]s %[5]s
	• %[2]s %[1]s --path %[4]s --data %[5]s
	• %[2]s %[1]s --path %[4]s --data %[6]s
				`, cst.Create, cst.NounHome, cst.ProductName, cst.ExamplePath, cst.ExampleDataJSON, cst.ExampleDataPath),
		FlagsPredictor: GetDataOpWrappers(cst.NounSecret),
		MinNumberArgs:  1,
	})
}

func GetHomeRollbackCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounHome, cst.Rollback},
		RunFunc: home{Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil, cst.NounHome}}.handleHomeRollback,
		SynopsisText: fmt.Sprintf("Rollback a home secret to a previous version %s", cst.NounHome),
		HelpText: fmt.Sprintf(`Rollback a %[2]s in %[3]s
Usage:
	• %[3]s %[1]s %[4]s --%[5]s 4
	• %[3]s %[1]s --path %[4]s
				`, cst.Rollback, cst.NounSecret, cst.NounHome, cst.ExamplePath, cst.Version),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):    cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s secret (required)", cst.Path, cst.NounHome)}), false},
			preds.LongFlag(cst.Version): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Version, Shorthand: "", Usage: "The version to which to rollback"}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetHomeSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounHome, cst.Search},
		RunFunc: home{Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil, cst.NounHome}}.handleHomeSearch,
		FlagsPredictor: GetSearchOpWrappers(),
		SynopsisText:   "Search for secrets in home",
		HelpText: fmt.Sprintf(`Search for a %[2]s from %[3]s

Usage:
    • %[2]s %[1]s %[4]s
    • %[2]s %[1]s --query %[4]s
    • %[2]s %[1]s --query aws:base:secret --search-links
    • %[2]s %[1]s --query aws --search-field attributes.type
    • %[2]s %[1]s --query 900 --search-field attributes.ttl --search-type number
    • %[2]s %[1]s --query production --search-field attributes.stage --search-comparison equal
            `, cst.Search, cst.NounHome, cst.ProductName, cst.ExampleUserSearch),
		MinNumberArgs: 1,
	})
}

func GetHomeDescribeCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounHome, cst.Describe},
		RunFunc: home{Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, nil, cst.NounHome}}.handleHomeDescribe,
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path): cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounSecret)}), false},
		},
		SynopsisText: "Get a secret description in home",
		HelpText: fmt.Sprintf(`
Usage:
• %[1]s %[2]s %[4]s
• %[1]s %[2]s --%[3]s %[4]s
`, cst.NounHome, cst.Describe, cst.Path, cst.ExamplePath),
		MinNumberArgs: 1,
	})
}

func GetHomeEditCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounHome, cst.Edit},
		RunFunc: home{Secret{
			requests.NewHttpClient(),
			nil,
			store.GetStore, EditData, cst.NounHome}}.handleHomeEdit,
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path): cli.PredictorWrapper{preds.NewSecretPathPredictorDefault(), preds.NewFlagValue(preds.Params{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounSecret)}), false},
		},
		SynopsisText: "Edit a secret in home",
		HelpText: fmt.Sprintf(`
Usage:
• %[1]s %[2]s %[4]s
• %[1]s %[2]s --%[3]s %[4]s
`, cst.NounHome, cst.Edit, cst.Path, cst.ExamplePath),
		MinNumberArgs: 1,
	})
}

func (h home) handleHomeRead(args []string) int {
	return h.handleReadCmd(args)
}

func (h home) handleHomeCreate(args []string) int {
	return h.handleUpsertCmd(args)
}

func (h home) handleHomeDelete(args []string) int {
	return h.handleDeleteCmd(args)
}

func (h home) handleHomeRestore(args []string) int {
	return h.handleRestoreCmd(args)
}

func (h home) handleHomeSearch(args []string) int {
	return h.handleSecretSearchCmd(args)
}

func (h home) handleHomeUpdate(args []string) int {
	return h.handleUpsertCmd(args)
}

func (h home) handleHomeRollback(args []string) int {
	return h.handleRollbackCmd(args)
}

func (h home) handleHomeDescribe(args []string) int {
	return h.handleDescribeCmd(args)
}

func (h home) handleHomeEdit(args []string) int {
	return h.handleEditCmd(args)
}
