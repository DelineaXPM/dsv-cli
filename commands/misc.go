package cmd

import (
	"fmt"
	"strings"

	"github.com/DelineaXPM/dsv-cli/auth"
	"github.com/DelineaXPM/dsv-cli/vaultcli"

	cst "github.com/DelineaXPM/dsv-cli/constants"

	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

func GetWhoAmICmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounWhoAmI},
		SynopsisText: "Show current identity",
		HelpText:     fmt.Sprintf("%s returns the current user identity, accounting for config, env, and flags", cst.NounWhoAmI),
		NoPreAuth:    true,
		RunFunc:      handleWhoAmICmd,
	})
}

func GetEvaluateFlagCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.EvaluateFlag},
		SynopsisText: "Inspect environment and configuration values",
		HelpText:     fmt.Sprintf("%s returns the value of the variable, accounting for config, env, and flags", cst.EvaluateFlag),
		NoPreAuth:    true,
		RunFunc:      handleEvaluateFlag,
	})
}

func handleWhoAmICmd(vcli vaultcli.CLI, args []string) int {
	subject, err := auth.GetCurrentIdentity()
	if err == nil {
		vcli.Out().WriteResponse([]byte(subject), nil)
	} else {
		vcli.Out().FailS("Failed to parse the subject from the auth token, try re-authenticating")
	}
	return 0
}

func handleEvaluateFlag(vcli vaultcli.CLI, args []string) int {
	if len(args) < 1 {
		return cli.RunResultHelp
	}
	arg := args[0]
	arg = strings.TrimPrefix(arg, "--")
	arg = strings.ReplaceAll(arg, "-", ".")
	arg = strings.ReplaceAll(arg, "_", ".")

	data := []byte(viper.GetString(arg))
	vcli.Out().WriteResponse(data, nil)
	return 0
}

func hasFlag(args []string, flagName string) bool {
	for _, fn := range args {
		if strings.HasPrefix(fn, flagName) {
			return true
		}
	}
	return false
}
