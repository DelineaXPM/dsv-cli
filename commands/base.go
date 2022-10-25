package cmd

import (
	"bytes"
	stderrors "errors"
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
		{Name: cst.Callback, Usage: fmt.Sprintf("Callback URL for oidc authentication [default: %s]", cst.DefaultCallback), Global: true, Hidden: true},
		{Name: cst.AuthProvider, Usage: "Authentication provider name for federated authentication", Global: true, Hidden: true},
		{Name: cst.Profile, Usage: "Configuration Profile [default:default]", Global: true},
		{Name: cst.Tenant, Shorthand: "t", Usage: "Tenant used for auth", Global: true},
		{Name: cst.DomainName, Usage: "Domain used for auth", Global: true},
		{Name: cst.Encoding, Shorthand: "e", Usage: "Output encoding (json|yaml) [default:json]", Global: true, Predictor: predictor.EncodingTypePredictor{}},
		{Name: cst.Beautify, Shorthand: "b", Usage: "Should beautify output", Global: true, ValueType: "bool", Hidden: true},
		{Name: cst.Plain, Usage: "Should not beautify output", Global: true, ValueType: "bool"},
		{Name: cst.Verbose, Shorthand: "v", Usage: "Verbose output [default:false]", Global: true, ValueType: "bool"},
		{Name: cst.Config, Shorthand: "c", Usage: fmt.Sprintf("Config file path [default:%s%s.dsv.yml]", homePath, string(os.PathSeparator)), Global: true},
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
	Path           []string
	RunFunc        func(vcli vaultcli.CLI, args []string) int
	WizardFunc     func(vcli vaultcli.CLI) int
	HelpText       string
	SynopsisText   string
	ArgsPredictor  complete.Predictor
	FlagsPredictor []*predictor.Params
	NoConfigRead   bool
	NoPreAuth      bool
	MinNumberArgs  int
}

func NewCommand(args CommandArgs) (cli.Command, error) {
	if len(args.Path) == 0 {
		return nil, fmt.Errorf("command path must be defined")
	}

	runFunc := args.RunFunc
	if runFunc == nil {
		// Show help by default.
		runFunc = func(vcli vaultcli.CLI, args []string) int { return cli.RunResultHelp }
	}

	cmd := &baseCommand{
		path:           args.Path,
		runFunc:        runFunc,
		wizardFunc:     args.WizardFunc,
		helpText:       args.HelpText,
		synopsisText:   args.SynopsisText,
		noConfigRead:   args.NoConfigRead,
		noPreAuth:      args.NoPreAuth,
		minNumberArgs:  args.MinNumberArgs,
		argsPredictor:  args.ArgsPredictor,
		flagsPredictor: make(map[string]*predictor.Wrapper),
	}

	for _, v := range BasePredictorWrappers() {
		w := predictor.New(v)
		cmd.flagsPredictor[w.FriendlyName] = w
	}
	for _, v := range args.FlagsPredictor {
		w := predictor.New(v)
		cmd.flagsPredictor[w.FriendlyName] = w
	}

	return cmd, nil
}

type baseCommand struct {
	path           []string
	runFunc        func(vcli vaultcli.CLI, args []string) int
	wizardFunc     func(vcli vaultcli.CLI) int
	helpText       string
	synopsisText   string
	argsPredictor  complete.Predictor
	flagsPredictor map[string]*predictor.Wrapper
	noConfigRead   bool
	noPreAuth      bool
	minNumberArgs  int
}

// Synopsis satisfies cli.Command interface.
func (c *baseCommand) Synopsis() string { return c.synopsisText }

// Run satisfies cli.Command interface.
func (c *baseCommand) Run(args []string) int {
	if len(args) < c.minNumberArgs {
		return cli.RunResultHelp
	}

	onlyGlobalFlags, err := c.parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Flags error: %v.\n", err)
		fmt.Fprintf(os.Stderr, "See %s %s --help.\n", os.Args[0], strings.Join(c.path, " "))
		return 1
	}

	doVerbose := viper.GetBool(cst.Verbose)

	if doVerbose {
		log.SetOutput(os.Stderr)

		fmt.Fprintf(os.Stderr, "DSV CLI version %s\n", version.Version)
		fmt.Fprintf(os.Stderr, "\t- platform:  %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Fprintf(os.Stderr, "\t- gitCommit: %s\n", version.GitCommit)
		fmt.Fprintf(os.Stderr, "\t- buildDate: %s\n\n", version.GetBuildDate())
	}

	vcli := vaultcli.New()

	if !c.noConfigRead {
		err := vaultcli.ViperInit()
		// If tenant is set then probably it is ok to run without configuration file.
		if err != nil && viper.GetString(cst.Tenant) == "" {
			if stderrors.Is(err, vaultcli.ErrFileNotFound) {
				vcli.Out().FailS("Run 'dsv init' to initiate CLI configuration - cannot find config.")
			} else {
				vcli.Out().FailF("Error: %v.", err)
			}
			return 1
		}
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

	if upd, err := version.CheckLatestVersion(); err != nil {
		log.Println(err)
	} else if upd != nil {
		log.SetOutput(os.Stderr)
		log.Println(upd)

		if !doVerbose {
			log.SetOutput(io.Discard)
		}
	}

	if !c.noPreAuth {
		tokenResponse, err := vcli.Authenticator().GetToken()
		if err != nil || tokenResponse == nil || tokenResponse.Token == "" {
			vcli.Out().WriteResponse(nil, err)
			return 1
		}
		viper.Set("token", tokenResponse.Token)
	}

	if onlyGlobalFlags && c.wizardFunc != nil {
		return c.wizardFunc(vcli)
	}
	return c.runFunc(vcli, args)
}

func (c *baseCommand) parseFlags() (bool, error) {
	// Return an error when parsing flags, so it can be handled outside of spf13/pflag.
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		return false, err
	}

	onlyGlobals := true

	for _, flg := range c.flagsPredictor {
		if flg.Name == "" {
			continue
		}

		v := flg.Val
		viperVal := viper.Get(flg.Name)
		flagVal := flg.Val.String()

		if flagVal != "" && !flg.Global {
			onlyGlobals = false
		}

		if flagVal != "" || (viperVal == "" && v.DefaultValue != "") {
			if flagVal == "" {
				flagVal = flg.Val.DefaultValue
			}

			if v.Type() == "bool" {
				if b, err := strconv.ParseBool(flagVal); err == nil {
					viper.Set(flg.Name, b)
				}
			} else {
				viper.Set(flg.Name, flagVal)
			}

			// HACK: There should be a better way to tell authenticator not to read from
			// cache and not to save to cache. This hack uses viper as a global storage
			// and passes configuration to authenticator through it.
			// This hack helps to skip cache when global auth related flag used.
			if flg.Global && strings.HasPrefix(flg.Name, cst.NounAuth) {
				viper.Set(cst.AuthSkipCache, true)
			}
		}
	}

	return onlyGlobals, nil
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

// AutocompleteFlags satisfies cli.CommandAutocomplete interface
func (c *baseCommand) AutocompleteFlags() complete.Flags {
	if c.flagsPredictor == nil {
		return nil
	}

	// Ignore error since not all autocomplete funcs require config.
	_ = vaultcli.ViperInit()

	flags := complete.Flags{}
	for k, v := range c.flagsPredictor {
		flags["--"+k] = v.Predictor
	}
	return flags
}

// AutocompleteArgs satisfies cli.CommandAutocomplete interface
func (c *baseCommand) AutocompleteArgs() complete.Predictor {
	if c.argsPredictor == nil {
		return nil
	}

	// Ignore error since not all autocomplete funcs require config.
	_ = vaultcli.ViperInit()

	return c.argsPredictor
}

func ValidateParams(params map[string]string, requiredKeys []string) *errors.ApiError {
	for _, k := range requiredKeys {
		if val, ok := params[k]; !ok || val == "" {
			return errors.NewF("--%s must be set", k)
		}
	}
	return nil
}
