package predictor

import (
	"os"
	"strings"

	"github.com/posener/complete"
	flag "github.com/spf13/pflag"

	"thy/internal/cmd"
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
	Name         string
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

	cmdFriendlyName := cmd.FriendlyName(params.Name)
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
		fv := FlagValue{
			FlagType:     params.ValueType,
			Name:         params.Name,
			DefaultValue: params.Default,
		}
		w.Val = &fv
		f := flag.CommandLine.VarPF(&fv, cmdFriendlyName, params.Shorthand, params.Usage)
		if params.ValueType == "bool" {
			f.NoOptDefVal = "true"
		}
	}

	return w
}

func (f *FlagValue) Set(v string) error {
	if f.FlagType == "" || f.FlagType == "string" {
		if v != "" && len(v) > 1 && strings.HasPrefix(v, "@") {
			f.FlagType = "file"
			fname := v[1:]
			if b, err := os.ReadFile(fname); err != nil {
				return err
			} else {
				f.Val = string(b)
				return nil
			}
		}
	}
	f.Val = v
	return nil
}

func (f *FlagValue) Type() string {
	return f.FlagType
}

func (f *FlagValue) String() string {
	return f.Val
}
