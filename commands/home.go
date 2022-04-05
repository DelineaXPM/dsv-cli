package cmd

import (
	"fmt"

	cst "thy/constants"
	"thy/internal/predictor"
	"thy/vaultcli"

	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

func GetHomeCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounHome},
		SynopsisText: "home (<path> | --path|-r)",
		HelpText: fmt.Sprintf(`Work with secrets in a personal user space

Usage:
   • home %[3]s
   • home --path %[3]s
		`, cst.NounHome, cst.ProductName, cst.ExamplePath),
		FlagsPredictor: GetNoDataOpWrappers(cst.NounSecret),
		MinNumberArgs:  1,
		RunFunc: func(args []string) int {
			path := viper.GetString(cst.Path)
			if path == "" && len(args) > 0 {
				path = args[0]
			}
			if path == "" {
				return cli.RunResultHelp
			}
			return handleHomeRead(vaultcli.New(), args)
		},
	})
}

func GetHomeReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounHome, cst.Read},
		SynopsisText: fmt.Sprintf("%s %s (<path> | --path|-r)", cst.NounHome, cst.Read),
		HelpText: fmt.Sprintf(`Read a a secret in %[2]s 
Usage:
   • home %[1]s %[4]s 
   • home %[1]s --path %[4]s`, cst.Read, cst.NounHome, cst.ProductName, cst.ExamplePath),
		FlagsPredictor:    GetNoDataOpWrappers(cst.NounSecret),
		ArgsPredictorFunc: predictor.NewSecretPathPredictorDefault().Predict,
		MinNumberArgs:     1,
		RunFunc: func(args []string) int {
			return handleHomeRead(vaultcli.New(), args)
		},
	})
}

func GetHomeCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounHome, cst.Create},
		SynopsisText: "Create a secret in home",
		HelpText: fmt.Sprintf(`Create a secret in %[2]s
Usage:
   • %[2]s %[1]s %[4]s %[5]s
   • %[2]s %[1]s --path %[4]s --data %[5]s
   • %[2]s %[1]s --path %[4]s --data %[6]s
		`, cst.Create, cst.NounHome, cst.ProductName, cst.ExamplePath, cst.ExampleDataJSON, cst.ExampleDataPath),
		FlagsPredictor: GetDataOpWrappers(cst.NounSecret),
		MinNumberArgs:  2,
		RunFunc: func(args []string) int {
			return handleHomeCreate(vaultcli.New(), args)
		},
	})
}

func GetHomeDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounHome, cst.Delete},
		SynopsisText: "Delete a secret from home",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[4]s
   • %[1]s %[2]s --%[3]s %[4]s
`, cst.NounHome, cst.Delete, cst.Path, cst.ExamplePath),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounSecret), Predictor: predictor.NewSecretPathPredictorDefault()},
			{Name: cst.Force, Usage: fmt.Sprintf("Immediately delete %s", cst.NounSecret), ValueType: "bool"},
		},
		MinNumberArgs: 1,
		RunFunc: func(args []string) int {
			return handleHomeDelete(vaultcli.New(), args)
		},
	})
}

func GetHomeRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounHome, cst.Restore},
		SynopsisText: fmt.Sprintf("Restore a soft-deleted secret in %s", cst.NounHome),
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[3]s
		`, cst.NounHome, cst.Restore, cst.ExamplePath),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounHome), Predictor: predictor.NewSecretPathPredictorDefault()},
		},
		MinNumberArgs: 1,
		RunFunc: func(args []string) int {
			return handleHomeRestore(vaultcli.New(), args)
		},
	})
}

func GetHomeUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounHome, cst.Update},
		SynopsisText: "Update a secret in home",
		HelpText: fmt.Sprintf(`Create a secret in %[2]s
Usage:
   • %[2]s %[1]s %[4]s %[5]s
   • %[2]s %[1]s --path %[4]s --data %[5]s
   • %[2]s %[1]s --path %[4]s --data %[6]s
		`, cst.Create, cst.NounHome, cst.ProductName, cst.ExamplePath, cst.ExampleDataJSON, cst.ExampleDataPath),
		FlagsPredictor: GetDataOpWrappers(cst.NounSecret),
		MinNumberArgs:  1,
		RunFunc: func(args []string) int {
			return handleHomeUpdate(vaultcli.New(), args)
		},
	})
}

func GetHomeRollbackCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounHome, cst.Rollback},
		SynopsisText: fmt.Sprintf("Rollback a home secret to a previous version %s", cst.NounHome),
		HelpText: fmt.Sprintf(`Rollback a %[2]s in %[3]s
Usage:
   • %[3]s %[1]s %[4]s --%[5]s 4
   • %[3]s %[1]s --path %[4]s
		`, cst.Rollback, cst.NounSecret, cst.NounHome, cst.ExamplePath, cst.Version),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s secret (required)", cst.Path, cst.NounHome), Predictor: predictor.NewSecretPathPredictorDefault()},
			{Name: cst.Version, Usage: "The version to which to rollback"},
		},
		MinNumberArgs: 1,
		RunFunc: func(args []string) int {
			return handleHomeRollback(vaultcli.New(), args)
		},
	})
}

func GetHomeSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:           []string{cst.NounHome, cst.Search},
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
		RunFunc: func(args []string) int {
			return handleHomeSearch(vaultcli.New(), args)
		},
	})
}

func GetHomeDescribeCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounHome, cst.Describe},
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounSecret), Predictor: predictor.NewSecretPathPredictorDefault()},
		},
		SynopsisText: "Get a secret description in home",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[4]s
   • %[1]s %[2]s --%[3]s %[4]s
`, cst.NounHome, cst.Describe, cst.Path, cst.ExamplePath),
		MinNumberArgs: 1,
		RunFunc: func(args []string) int {
			return handleHomeDescribe(vaultcli.New(), args)
		},
	})
}

func GetHomeEditCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounHome, cst.Edit},
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Shorthand: "r", Usage: fmt.Sprintf("Target %s to a %s (required)", cst.Path, cst.NounSecret), Predictor: predictor.NewSecretPathPredictorDefault()},
		},
		SynopsisText: "Edit a secret in home",
		HelpText: fmt.Sprintf(`
Usage:
   • %[1]s %[2]s %[4]s
   • %[1]s %[2]s --%[3]s %[4]s
`, cst.NounHome, cst.Edit, cst.Path, cst.ExamplePath),
		MinNumberArgs: 1,
		RunFunc: func(args []string) int {
			return handleHomeEdit(vaultcli.New(), args)
		},
	})
}

func handleHomeRead(vcli vaultcli.CLI, args []string) int {
	return handleSecretReadCmd(vcli, cst.NounHome, args)
}

func handleHomeCreate(vcli vaultcli.CLI, args []string) int {
	return handleSecretUpsertCmd(vcli, cst.NounHome, args)
}

func handleHomeDelete(vcli vaultcli.CLI, args []string) int {
	return handleSecretDeleteCmd(vcli, cst.NounHome, args)
}

func handleHomeRestore(vcli vaultcli.CLI, args []string) int {
	return handleSecretRestoreCmd(vcli, cst.NounHome, args)
}

func handleHomeSearch(vcli vaultcli.CLI, args []string) int {
	return handleSecretSearchCmd(vcli, cst.NounHome, args)
}

func handleHomeUpdate(vcli vaultcli.CLI, args []string) int {
	return handleSecretUpsertCmd(vcli, cst.NounHome, args)
}

func handleHomeRollback(vcli vaultcli.CLI, args []string) int {
	return handleSecretRollbackCmd(vcli, cst.NounHome, args)
}

func handleHomeDescribe(vcli vaultcli.CLI, args []string) int {
	return handleSecretDescribeCmd(vcli, cst.NounHome, args)
}

func handleHomeEdit(vcli vaultcli.CLI, args []string) int {
	return handleSecretEditCmd(vcli, cst.NounHome, args)
}
