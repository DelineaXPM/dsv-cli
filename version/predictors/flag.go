package predictors

import (
	flag "github.com/spf13/pflag"
	"github.com/thycotic-rd/cli"
	"strings"
)

type Params struct {
	Name      string
	Shorthand string
	Usage     string
	Default   string
	ValueType string
	Global    bool
	Hidden    bool
}

func NewFlagValue(params Params) *cli.FlagWrapper {
	cmdFriendlyName := CmdFriendlyName(params.Name)
	fw := cli.FlagWrapper{Name: params.Name, FriendlyName: cmdFriendlyName, FlagType: params.ValueType,
		Shorthand: params.Shorthand, Usage: params.Usage, Global: params.Global, Hidden: params.Hidden}

	if f := flag.Lookup(cmdFriendlyName); f != nil {
		fw.Val = f.Value.(*cli.FlagValue)

	} else {
		fv := cli.FlagValue{FlagType: params.ValueType, Name: params.Name, DefaultValue: params.Default, Shorthand: params.Shorthand}
		fw.Val = &fv
		f := flag.CommandLine.VarPF(&fv, cmdFriendlyName, params.Shorthand, params.Usage)
		if params.ValueType == "bool" {
			f.NoOptDefVal = "true"
		}
	}

	return &fw
}

func LongFlag(flag string) string {
	return "--" + CmdFriendlyName(flag)
}

func CmdFriendlyName(flag string) string {
	return strings.Replace(flag, ".", "-", -1)
}
