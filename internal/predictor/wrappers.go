package predictor

import (
	"os"
	"strings"

	"thy/vaultcli"

	"github.com/posener/complete"
	flag "github.com/spf13/pflag"
)

// Wrapper merges a flag with its predictor.
type Wrapper struct {
	Predictor    complete.Predictor
	Val          *FlagValue
	Name         string
	FriendlyName string
	Shorthand    string
	Usage        string
	Global       bool
	Hidden       bool
}

// FlagValue is a generalized storage for a flag's value.
type FlagValue struct {
	Val          string
	FlagType     string
	DefaultValue string
}

type Params struct {
	Name      string
	Shorthand string
	Usage     string
	Default   string
	ValueType string
	Global    bool
	Hidden    bool
	Predictor complete.Predictor
}

func New(params *Params) *Wrapper {
	pred := params.Predictor
	if pred == nil {
		if params.ValueType == "bool" {
			pred = complete.PredictNothing
		} else {
			pred = complete.PredictAnything
		}
	}

	cmdFriendlyName := vaultcli.ToFlagName(params.Name)
	w := &Wrapper{
		Predictor:    pred,
		Name:         params.Name,
		FriendlyName: cmdFriendlyName,
		Shorthand:    params.Shorthand,
		Usage:        params.Usage,
		Global:       params.Global,
		Hidden:       params.Hidden,
	}

	if f := flag.Lookup(cmdFriendlyName); f != nil {
		w.Val = f.Value.(*FlagValue)

	} else {
		w.Val = &FlagValue{
			FlagType:     params.ValueType,
			DefaultValue: params.Default,
		}
		flag.VarP(w.Val, cmdFriendlyName, params.Shorthand, params.Usage)
		if params.ValueType == "bool" {
			flag.Lookup(cmdFriendlyName).NoOptDefVal = "true"
		}
	}

	if w.Hidden {
		flag.CommandLine.MarkHidden(cmdFriendlyName)
	}

	return w
}

func (f *FlagValue) Set(value string) error {
	if f.FlagType == "" || f.FlagType == "string" {
		if len(value) > 1 && strings.HasPrefix(value, "@") {
			f.FlagType = "file"
			fname := value[1:]
			b, err := os.ReadFile(fname)
			if err != nil {
				return err
			}
			value = string(b)
		}
	}
	f.Val = value
	return nil
}

func (f *FlagValue) Type() string   { return f.FlagType }
func (f *FlagValue) String() string { return f.Val }
