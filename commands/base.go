package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"thy/auth"
	cst "thy/constants"
	"thy/errors"
	"thy/format"
	preds "thy/predictors"
	"thy/utils"
	"thy/version"

	"github.com/posener/complete"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/thycotic-rd/cli"

	"golang.org/x/sys/execabs"
)

func BasePredictorWrappers() cli.PredictorWrappers {
	homePath := "$HOME"
	if utils.NewEnvProvider().GetOs() == "windows" {
		homePath = "%USERPROFILE%"
	}
	return cli.PredictorWrappers{
		preds.LongFlag(cst.Callback):     cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Callback, Usage: fmt.Sprintf("Callback URL for oidc authentication [default: %s", cst.DefaultCallback), Global: true, Hidden: true}), false},
		preds.LongFlag(cst.AuthProvider): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.AuthProvider, Usage: "Authentication provider name for federated authentication", Global: true, Hidden: true}), false},
		preds.LongFlag(cst.Profile):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Profile, Usage: "Configuration Profile [default:default]", Global: true}), false},
		preds.LongFlag(cst.Tenant):       cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Tenant, Shorthand: "t", Usage: "Tenant used for auth", Global: true}), false},
		preds.LongFlag(cst.DomainName):   cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DomainName, Usage: "Domain used for auth", Global: true}), false},
		preds.LongFlag(cst.Encoding):     cli.PredictorWrapper{preds.EncodingTypePredictor{}, preds.NewFlagValue(preds.Params{Name: cst.Encoding, Shorthand: "e", Usage: "Output encoding (json|yaml) [default:json]", Global: true}), false},
		preds.LongFlag(cst.Beautify):     cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Beautify, Shorthand: "b", Usage: "Should beautify output", Global: true, ValueType: "bool", Hidden: true}), false},
		// we could get away with just one of beautify / plain but gets tricky because we want the default to be beautify,
		// unless it's a read operation
		preds.LongFlag(cst.Plain):   cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Plain, Usage: "Should not beautify output (overrides beautify)", Global: true, ValueType: "bool"}), false},
		preds.LongFlag(cst.Verbose): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Verbose, Shorthand: "v", Usage: "Verbose output [default:false]", Global: true, ValueType: "bool"}), false},
		preds.LongFlag(cst.Config):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Config, Shorthand: "c", Usage: fmt.Sprintf("Config file path [default:%s%s.thy.yaml]", homePath, string(os.PathSeparator)), Global: true}), false},
		preds.LongFlag(cst.Filter):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Filter, Shorthand: "f", Usage: "Filter in jq (stedolan.github.io/jq)", Global: true}), false},
		preds.LongFlag(cst.Output):  cli.PredictorWrapper{preds.OutputTypePredictor{}, preds.NewFlagValue(preds.Params{Name: cst.Output, Shorthand: "o", Usage: "Output destination (stdout|clip|file:<fname>) [default:stdout]", Global: true}), false},

		preds.LongFlag(cst.AuthType): cli.PredictorWrapper{preds.AuthTypePredictor{}, preds.NewFlagValue(preds.Params{Name: cst.AuthType, Shorthand: "a", Usage: "Auth Type (" +
			strings.Join([]string{string(auth.Password), string(auth.ClientCredential), string(auth.FederatedAws), string(auth.FederatedAzure), string(auth.FederatedGcp)}, "|") + ")", Global: true}), false},
		preds.LongFlag(cst.AwsProfile):        cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.AwsProfile, Usage: "AWS profile", Global: true}), false},
		preds.LongFlag(cst.Username):          cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Username, Shorthand: "u", Usage: "User", Global: true}), false},
		preds.LongFlag(cst.Password):          cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Password, Shorthand: "p", Usage: "Password", Global: true}), false},
		preds.LongFlag(cst.AuthClientID):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.AuthClientID, Usage: "Client ID", Global: true}), false},
		preds.LongFlag(cst.AuthClientSecret):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.AuthClientSecret, Usage: "Client Secret", Global: true}), false},
		preds.LongFlag(cst.GcpAuthType):       cli.PredictorWrapper{preds.GcpAuthTypePredictor{}, preds.NewFlagValue(preds.Params{Name: cst.GcpAuthType, Usage: "GCP Auth Type (gce|iam)", Global: true}), false},
		preds.LongFlag(cst.GcpServiceAccount): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.GcpServiceAccount, Usage: "GCP Service Account Name", Global: true}), false},
		preds.LongFlag(cst.GcpProject):        cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.GcpProject, Usage: "GCP Project", Global: true}), false},
		preds.LongFlag(cst.GcpToken):          cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.GcpToken, Usage: "GCP OIDC Token", Global: true}), false},
	}
}

type CommandArgs struct {
	Path              []string
	RunFunc           func(args []string) int
	HelpText          string
	SynopsisText      string
	ArgsPredictorFunc func(complete.Args) []string
	FlagsPredictor    map[string]cli.PredictorWrapper
	NoPreAuth         bool
	MinNumberArgs     int
	IsTailCmd         bool
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

	if args.FlagsPredictor != nil {
		if mergedPredictors, err := BasePredictorWrappers().Merge(args.FlagsPredictor, false); err != nil {
			return nil, err
		} else {
			cmd.flagsPredictor = mergedPredictors
		}
	} else {
		cmd.flagsPredictor = BasePredictorWrappers()
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
	flagsPredictor    map[string]cli.PredictorWrapper
	flags             flag.Flag
	authenticatorFunc func() auth.Authenticator
	noPreAuth         bool
	minNumberArgs     int
}

func (c *baseCommand) GetFlagsPredictor() cli.PredictorWrappers {
	return c.flagsPredictor
}

func (c *baseCommand) preRun(args []string) int {
	if len(args) < c.minNumberArgs {
		return cli.RunResultHelp
	}

	name := c.path[len(c.path)-1]
	viper.Set(cst.MainCommand, c.path[0])
	viper.Set(cst.LastCommandKey, name)
	viper.Set(cst.FullCommandKey, strings.Join(c.path, " "))
	c.SetFlags()
	setVerbosity()

	configureFormattingOptions()

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

func configureFormattingOptions() {
	beautify := viper.GetBool(cst.Beautify)
	beautifyComputed := beautify
	if !beautifyComputed {
		method := strings.ToLower(viper.GetString(cst.FullCommandKey))
		if !strings.Contains(method, cst.Read) {
			beautifyComputed = true
		}
	}
	// todo : ideally we break out format logic so they can either do:
	// 1. result as json (but not colorized)
	// 2. result as json and colorized (unix terminal only)
	// 3. result as yaml
	// but currently, converting result to yaml is tied to beautification function
	// result: they can do everything fine except when displaying output as pretty json on *nix terminal, will always be colorized
	encoding := viper.GetString(cst.Encoding)
	encodingComputed := encoding
	if encodingComputed == "" {
		encodingComputed = cst.Json
	} else {
		beautifyComputed = true
	}

	if beautifyComputed && viper.GetBool(cst.Plain) {
		beautifyComputed = false
	}
	if encodingComputed != encoding {
		viper.Set(cst.Encoding, encodingComputed)
	}
	if beautifyComputed != beautify {
		viper.Set(cst.Beautify, beautifyComputed)
	}
}

func setVerbosity() {
	if viper.GetBool(cst.Verbose) {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}
}

func (c *baseCommand) SetFlags() {
	// var cfgFile string
	flag.Parse()
	// if configFlag := flag.Lookup("config"); configFlag != nil && configFlag.Value != nil {
	// 	cfgFile = configFlag.Value.String()
	// }
	//config.InitConfig(cfgFile)
	for _, e := range c.GetFlagsPredictor() {
		w := e.Flag
		v := w.Val
		if v.Name == "" {
			continue
		}
		viperVal := viper.Get(v.Name)
		if v.String() != "" || (viperVal == "" && v.DefaultValue != "") {
			var val string
			if v.String() != "" {
				val = v.String()
			} else {
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

// Help satisfies cli.Command interface
func (c *baseCommand) Help() string {
	return c.helpText
}

// Synopsis satisfies cli.Command interface
func (c *baseCommand) Synopsis() string {
	return c.synopsisText
}

func (c *baseCommand) GetPredictorWrappers() cli.PredictorWrappers {
	return c.flagsPredictor
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
		if !cst.DontAutocompleteGlobals || !v.Flag.Global {
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

func GetFlagName(arg string) string {
	if strings.HasPrefix(arg, "--") {
		return arg[2:]
	}
	if strings.HasPrefix(arg, "-") {
		return arg[1:]
	}
	return ""
}

// OnlyGlobalArgs checks if passed in command line args to a subcommand are only global flags.
// It assumes viper had already set values for relevant flags like profile and config.
func OnlyGlobalArgs(args []string) bool {
	globalFlags := BasePredictorWrappers()
outer:
	for _, arg := range args {
		f := GetFlagName(arg)
		if f == "" || f == "v" {
			continue
		}
		for _, g := range globalFlags {
			if f == g.Flag.Name && g.Flag.Global {
				continue outer
			}
		}
		return false
	}
	return true
}

type dataFunc func(data []byte) (resp []byte, err *errors.ApiError)

func EditData(data []byte, saveFunc dataFunc, startErr *errors.ApiError, retry bool) (edited []byte, runErr *errors.ApiError) {
	viper.Set(cst.Output, string(format.File))
	dataFormatted, errString := format.FormatResponse(data, nil, viper.GetBool(cst.Beautify))
	viper.Set(cst.Output, string(format.StdOut))
	if errString != "" {
		return nil, errors.NewS(errString)
	}
	dataEdited, err := doEditData([]byte(dataFormatted), startErr)
	if err != nil {
		return nil, err
	}
	resp, postErr := saveFunc(dataEdited)
	if retry && postErr != nil {
		return EditData(dataEdited, saveFunc, postErr, true)
	}
	return resp, postErr
}

func doEditData(data []byte, startErr *errors.ApiError) (edited []byte, runErr *errors.ApiError) {
	editorCmd, getErr := getEditorCmd()
	if getErr != nil || editorCmd == "" {
		return nil, getErr
	}
	tmpDir := os.TempDir()
	tmpFile, err := ioutil.TempFile(tmpDir, cst.CmdRoot)
	if err != nil {
		return nil, errors.New(err).Grow("Error while creating temp file to edit data")
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			log.Printf("Warning: failed to remove temporary file: '%s'\n%v", tmpFile.Name(), err)
		}
	}()

	if err := ioutil.WriteFile(tmpFile.Name(), data, 0600); err != nil {
		return nil, errors.New(err).Grow("Error while copying data to temp file")
	}

	// This is necessary for Windows. Opening a file in a parent process and then
	// trying to write it in a child process is not allowed. So we close the file
	// in the parent process first. This does not affect the behavior on Unix.
	if err := tmpFile.Close(); err != nil {
		log.Printf("Warning: failed to close temporary file: '%s'\n%v", tmpFile.Name(), err)
	}

	editorPath, err := execabs.LookPath(editorCmd)
	if err != nil {
		return nil, errors.New(err).Grow(fmt.Sprintf("Error while looking up path to editor %q", editorCmd))
	}
	args := []string{tmpFile.Name()}
	if startErr != nil && (strings.HasSuffix(editorPath, "vim") || strings.HasSuffix(editorPath, "vi")) {
		args = append(args, "-c")
		errMsg := fmt.Sprintf("Error saving to %s. Please correct and save again or exit: %s", cst.ProductName, startErr.String())
		args = append(args, fmt.Sprintf(`:echoerr '%s'`, strings.Replace(errMsg, `'`, `''`, -1)))
	}
	cmd := execabs.Command(editorPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return nil, errors.New(err).Grow(fmt.Sprintf("Command failed to start: '%s %s'", editorCmd, tmpFile.Name()))
	}
	err = cmd.Wait()
	if err != nil {
		return nil, errors.New(err).Grow(fmt.Sprintf("Command failed to return: '%s %s'", editorCmd, tmpFile.Name()))
	}
	edited, err = ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, errors.New(err).Grow(fmt.Sprintf("Failed to read edited file: %q", tmpFile.Name()))
	}
	if utils.SlicesEqual(data, edited) {
		return nil, errors.NewS("Data not modified")
	}
	if len(edited) == 0 {
		return nil, errors.NewS("Cannot save empty file")
	}

	return edited, startErr
}

func getEditorCmd() (string, *errors.ApiError) {
	if utils.NewEnvProvider().GetOs() == "windows" {
		return "notepad.exe", nil
	}
	editor := viper.GetString(cst.Editor)
	// if editor specified in cli-config
	if editor != "" {
		return editor, nil
	}

	// try to find default editor on system
	out, err := execabs.Command("bash", "-c", getDefaultEditorSh).Output()
	editor = strings.TrimSpace(string(out))
	if err != nil || editor == "" {
		return "", errors.New(err).Grow("Failed to find default text editor. Please set 'editor' in the cli-config or make sure $EDITOR, $VISUAL is set on your system.")
	}

	// verbose - let them know why a certain editor is being implicitly chosen
	log.Printf("Using editor '%s' as it is found as default editor on the system. To override, set in cli-config (%s config update editor <EDITOR_NAME>)", editor, cst.CmdRoot)
	return editor, nil
}

const getDefaultEditorSh = `
#!/bin/sh
if [ -n "$VISUAL" ]; then
  echo $VISUAL
elif [ -n "$EDITOR" ]; then
  echo $EDITOR
elif type sensible-editor >/dev/null 2>/dev/null; then
  echo sensible-editor "$@"
elif cmd=$(xdg-mime query default ) 2>/dev/null; [ -n "$cmd" ]; then
  echo "$cmd"
else
  editors='nano joe vi'
  if [ -n "$DISPLAY" ]; then
    editors="gedit kate $editors"
  fi
  for x in $editors; do
    if type "$x" >/dev/null 2>/dev/null; then
	  echo "$x"
	  exit 0
    fi
  done
fi
`
