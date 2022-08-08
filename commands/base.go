package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"text/template"

	"thy/auth"
	cst "thy/constants"
	"thy/errors"
	"thy/format"
	"thy/internal/predictor"
	"thy/utils"
	"thy/vaultcli"
	"thy/version"

	"github.com/mitchellh/cli"
	"github.com/posener/complete"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func BasePredictorWrappers() []*predictor.Params {
	homePath := "$HOME"
	if utils.NewEnvProvider().GetOs() == "windows" {
		homePath = "%USERPROFILE%"
	}
	return []*predictor.Params{
		{Name: cst.Callback, Usage: fmt.Sprintf("Callback URL for oidc authentication [default: %s", cst.DefaultCallback), Global: true, Hidden: true},
		{Name: cst.AuthProvider, Usage: "Authentication provider name for federated authentication", Global: true, Hidden: true},
		{Name: cst.Profile, Usage: "Configuration Profile [default:default]", Global: true},
		{Name: cst.Tenant, Shorthand: "t", Usage: "Tenant used for auth", Global: true},
		{Name: cst.DomainName, Usage: "Domain used for auth", Global: true},
		{Name: cst.Encoding, Shorthand: "e", Usage: "Output encoding (json|yaml) [default:json]", Global: true, Predictor: predictor.EncodingTypePredictor{}},
		{Name: cst.Beautify, Shorthand: "b", Usage: "Should beautify output", Global: true, ValueType: "bool", Hidden: true},
		{Name: cst.Plain, Usage: "Should not beautify output", Global: true, ValueType: "bool"},
		{Name: cst.Verbose, Shorthand: "v", Usage: "Verbose output [default:false]", Global: true, ValueType: "bool"},
		{Name: cst.Config, Shorthand: "c", Usage: fmt.Sprintf("Config file path [default:%s%s.thy.yaml]", homePath, string(os.PathSeparator)), Global: true},
		{Name: cst.Filter, Shorthand: "f", Usage: "Filter in jq (stedolan.github.io/jq)", Global: true},
		{Name: cst.Output, Shorthand: "o", Usage: "Output destination (stdout|clip|file:<fname>) [default:stdout]", Global: true, Predictor: predictor.OutputTypePredictor{}},

		{Name: cst.AuthType, Shorthand: "a", Usage: "Auth Type (" + strings.Join([]string{string(auth.Password), string(auth.ClientCredential), string(auth.FederatedAws), string(auth.FederatedAzure), string(auth.FederatedGcp)}, "|") + ")", Global: true, Predictor: predictor.AuthTypePredictor{}},
		{Name: cst.AwsProfile, Usage: "AWS profile", Global: true},
		{Name: cst.Username, Shorthand: "u", Usage: "User", Global: true},
		{Name: cst.Password, Shorthand: "p", Usage: "Password", Global: true},
		{Name: cst.AuthClientID, Usage: "Client ID", Global: true},
		{Name: cst.AuthClientSecret, Usage: "Client Secret", Global: true},
		{Name: cst.GcpAuthType, Usage: "GCP Auth Type (gce|iam)", Global: true, Predictor: predictor.GcpAuthTypePredictor{}},
		{Name: cst.GcpServiceAccount, Usage: "GCP Service Account Name", Global: true},
		{Name: cst.GcpProject, Usage: "GCP Project", Global: true},
		{Name: cst.GcpToken, Usage: "GCP OIDC Token", Global: true},
	}
}

type CommandArgs struct {
	Path              []string
	RunFunc           func(args []string) int
	HelpText          string
	SynopsisText      string
	ArgsPredictorFunc func(complete.Args) []string
	FlagsPredictor    []*predictor.Params
	NoPreAuth         bool
	MinNumberArgs     int
}

// NewCommand creates new baseCommand
func NewCommand(args CommandArgs) (cli.Command, error) {
	if args.Path == nil || len(args.Path) < 1 {
		return nil, utils.NewMissingArgError(cst.Path)
	}
	cmd := &baseCommand{
		path:              args.Path,
		runFunc:           args.RunFunc,
		helpText:          args.HelpText,
		synopsisText:      args.SynopsisText,
		authenticatorFunc: auth.NewAuthenticatorDefault,
		noPreAuth:         args.NoPreAuth,
		minNumberArgs:     args.MinNumberArgs,
	}

	cmd.flagsPredictor = make(map[string]*predictor.Wrapper)
	for _, v := range BasePredictorWrappers() {
		w := predictor.New(v)
		cmd.flagsPredictor["--"+w.FriendlyName] = w
	}
	for _, v := range args.FlagsPredictor {
		w := predictor.New(v)
		cmd.flagsPredictor["--"+w.FriendlyName] = w
	}

	if args.ArgsPredictorFunc != nil {
		cmd.AddArgsPredictorFunc(args.ArgsPredictorFunc)
	}
	return cmd, nil
}

// AddArgsPredictorFunc adds custom args predictors
func (c *baseCommand) AddArgsPredictorFunc(predictorFunc func(complete.Args) []string) *baseCommand {
	c.argsPredictorFunc = func() complete.Predictor { return complete.PredictFunc(predictorFunc) }
	return c
}

type baseCommand struct {
	path              []string
	runFunc           func(args []string) int
	helpText          string
	synopsisText      string
	argsPredictorFunc func() complete.Predictor
	flagsPredictor    map[string]*predictor.Wrapper
	authenticatorFunc func() auth.Authenticator
	noPreAuth         bool
	minNumberArgs     int
}

func (c *baseCommand) preRun(args []string) int {
	if len(args) < c.minNumberArgs {
		return cli.RunResultHelp
	}

	viper.Set(cst.MainCommand, c.path[0])

	c.SetFlags()
	setVerbosity()

	if viper.GetBool(cst.Verbose) {
		log.Printf("%s CLI", cst.ProductName)
		log.Printf("\t- version:   %s", version.Version)
		log.Printf("\t- platform:  %s/%s", runtime.GOOS, runtime.GOARCH)
		log.Printf("\t- gitCommit: %s", version.GitCommit)
		log.Printf("\t- buildDate: %s", version.GetBuildDate())
	}

	encoding := viper.GetString(cst.Encoding)
	if encoding == "" {
		encoding = cst.Json
	} else {
		encoding = strings.ToLower(encoding)
	}
	viper.Set(cst.Encoding, encoding)

	// The --plain flag overrides the --beautify flag.
	beautify := !viper.GetBool(cst.Plain)
	viper.Set(cst.Beautify, beautify)

	// Set profile name to lower case globally.
	viper.Set(cst.Profile, strings.ToLower(viper.GetString(cst.Profile)))

	if upd, err := version.CheckLatestVersion(); err != nil {
		log.Println(err)
	} else if upd != nil {
		log.SetOutput(os.Stderr)
		log.Println(upd)
		setVerbosity()
	}

	if !c.noPreAuth {
		if tr, err := c.authenticatorFunc().GetToken(); err != nil || tr == nil || tr.Token == "" {
			format.NewDefaultOutClient().WriteResponse(nil, err)
			os.Exit(1)
		} else {
			viper.Set("token", tr.Token)
		}
	}
	return 0
}

func setVerbosity() {
	if viper.GetBool(cst.Verbose) {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(io.Discard)
	}
}

func (c *baseCommand) SetFlags() {
	flag.Parse()

	for _, e := range c.flagsPredictor {
		v := e.Val
		if v.Name == "" {
			continue
		}
		viperVal := viper.Get(v.Name)
		if v.String() != "" || (viperVal == "" && v.DefaultValue != "") {
			val := v.String()
			if val == "" {
				val = v.DefaultValue
			}

			if v.Type() == "bool" {
				if b, err := strconv.ParseBool(val); err == nil {
					viper.Set(v.Name, b)
				}
			} else {
				viper.Set(v.Name, val)
			}
		}
	}
}

// Run satisfies cli.Command interface
func (c *baseCommand) Run(args []string) int {
	sig := c.preRun(args)
	if sig != 0 {
		return sig
	}
	return c.runFunc(args)
}

// Help satisfies cli.Command interface.
func (c *baseCommand) Help() string {
	const helpTemplate = `Cmd: {{.Cmd}}	{{.Synopsis}}{{if ne .Help ""}}

{{.Help}}{{ end}}{{if gt (len .Flags) 0}}
Flags:{{ range $value := .Flags }}
   --{{ $value.FriendlyName }}{{if ne $value.Shorthand ""}}, -{{$value.Shorthand}}{{end}}	{{ $value.Usage }}{{ end }}
{{ end }}{{if gt (len .FlagsGlobal) 0}}
Global:{{ range $value := .FlagsGlobal }}
   --{{ $value.FriendlyName }}{{if ne $value.Shorthand ""}}, -{{$value.Shorthand}}{{end}}	{{ $value.Usage }}{{ end }}{{ end }}`

	flags := []*predictor.Wrapper{}
	flagsGlobal := []*predictor.Wrapper{}
	for _, pw := range c.flagsPredictor {
		if pw.Hidden {
			continue
		}
		if pw.Global {
			flagsGlobal = append(flagsGlobal, pw)
		} else {
			flags = append(flags, pw)
		}
	}
	sort.Slice(flagsGlobal, func(i, j int) bool {
		return flagsGlobal[i].Name < flagsGlobal[j].Name
	})
	sort.Slice(flags, func(i, j int) bool {
		return flags[i].Name < flags[j].Name
	})

	data := map[string]interface{}{
		"Cmd":         strings.Join(c.path, " "),
		"Synopsis":    c.synopsisText,
		"Help":        c.helpText,
		"Flags":       flags,
		"FlagsGlobal": flagsGlobal,
	}

	t, err := template.New("helptext").Parse(helpTemplate)
	if err != nil {
		return fmt.Sprintf("Internal error! Failed to parse command help template: %s\n", err)
	}

	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 8, 8, 8, ' ', 0)
	err = t.Execute(w, data)
	if err != nil {
		return fmt.Sprintf("Internal error! Failed to execute command help template: %s\n", err)
	}

	err = w.Flush()
	if err != nil {
		return fmt.Sprintf("Internal error! Failed to write command help template: %s\n", err)
	}

	return b.String()
}

// Synopsis satisfies cli.Command interface
func (c *baseCommand) Synopsis() string {
	return c.synopsisText
}

// AutocompleteFlags satisfies cli.CommandAutocomplete interface
func (c *baseCommand) AutocompleteFlags() complete.Flags {
	if c.flagsPredictor == nil {
		return nil
	}

	// TODO : THIS IS INEFFICENT. Need to change complete.command.go nil checks becuase they think that
	// a derived type is not nil (complete.PredictNothing) when they are nil
	flags := complete.Flags{}
	for k, v := range c.flagsPredictor {
		if !cst.DontAutocompleteGlobals || !v.Global {
			flags[k] = v.Predictor
		}
	}
	return flags
}

// AutocompleteArgs satisfies cli.CommandAutocomplete interface
func (c *baseCommand) AutocompleteArgs() complete.Predictor {
	if c.argsPredictorFunc == nil {
		return nil
	}
	return c.argsPredictorFunc()
}

func ValidateParams(params map[string]string, requiredKeys []string) *errors.ApiError {
	for _, k := range requiredKeys {
		if val, ok := params[k]; !ok || val == "" {
			return utils.NewMissingArgError(k)
		}
	}
	return nil
}

// IsInit checks if passed in command line args contain an init command. IsInit supports both cli.Args and os.Args.
func IsInit(args []string) bool {
	if len(args) == 0 {
		return false
	}

	var a []string
	if args[0] == cst.CmdRoot {
		a = args[1:]
	} else {
		a = args
	}

	if len(a) == 0 {
		return false
	}
	if a[0] == cst.Init {
		return true
	}
	if len(a) > 1 && strings.Contains(strings.Join(a, " "), cst.NounCliConfig+" "+cst.Init) {
		return true
	}
	return false
}

// IsInstall checks if passed in command line args contain an install command.
func IsInstall(args []string) bool {
	for _, a := range args {
		if a == "--install" || a == "-install" {
			return true
		}
	}
	return false
}

// OnlyGlobalArgs checks if passed in command line args to a subcommand are only global flags.
// It assumes viper had already set values for relevant flags like profile and config.
func OnlyGlobalArgs(args []string) bool {
	globalFlags := BasePredictorWrappers()

	var isGlobal bool
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			continue // skip, not a flag
		}

		f := strings.TrimPrefix(arg, "--")
		f = strings.TrimPrefix(f, "-")
		f = strings.Split(f, "=")[0]

		isGlobal = false
		for _, g := range globalFlags {
			if f == vaultcli.ToFlagName(g.Name) || f == g.Shorthand {
				isGlobal = g.Global
				break
			}
		}
		if !isGlobal {
			return false
		}
	}
	return true
}
