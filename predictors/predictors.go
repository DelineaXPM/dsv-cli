package predictors

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"thy/auth"
	cst "thy/constants"
	"thy/format"
	"thy/paths"
	"thy/requests"

	"github.com/posener/complete"
	"github.com/spf13/viper"
	"github.com/thycotic-rd/cli"
)

type PathAutocompleteResult struct {
	Data []string `json:"data"`
}

// PredictorWrapper wraps a Predictor with a FlagValue
type PredictorWrapper struct {
	complete.Predictor
	Value          *cli.FlagValue
	PredictNothing bool
}

// PredictorWrappers maps a flags full name to its PredictorWrapper
type PredictorWrappers map[string]PredictorWrapper

// NewStringPredictor gets a new StringPredictor
func NewStringPredictor(flagName string, flagShort string, flagUsageFmt string) cli.PredictorWrappers {
	return cli.PredictorWrappers{
		LongFlag(flagName): cli.PredictorWrapper{complete.PredictAnything, NewFlagValue(Params{Name: flagName, Shorthand: flagShort, Usage: fmt.Sprintf(flagUsageFmt, flagName)}), false},
	}
}

// NewSecretPathPredictorDefault returns a new SecretPathPredictor
func NewSecretPathPredictorDefault() complete.Predictor {
	return &secretPathPredictor{
		authenticatorFunc: auth.NewAuthenticatorDefault,
		requestClientFunc: requests.NewHttpClient,
	}
}

// NewSecretPathPredictor returns a new SecretPathPredictor
func NewSecretPathPredictor(authenticatorFunc func() auth.Authenticator,
	requestClientFunc func() requests.Client) complete.Predictor {
	return &secretPathPredictor{
		authenticatorFunc: authenticatorFunc,
		requestClientFunc: requestClientFunc,
	}
}

type secretPathPredictor struct {
	authenticator     auth.Authenticator
	authenticatorFunc func() auth.Authenticator
	requestClient     requests.Client
	requestClientFunc func() requests.Client
}

func (p *secretPathPredictor) getAuthenticator() auth.Authenticator {
	if p.authenticator == nil {
		a := p.authenticatorFunc()
		p.authenticator = a
	}
	return p.authenticator
}

func (p *secretPathPredictor) getRequestClient() requests.Client {
	if p.requestClient == nil {
		c := p.requestClientFunc()
		p.requestClient = c
	}
	return p.requestClient
}

func (p *secretPathPredictor) Predict(a complete.Args) (prediction []string) {
	// NOTE : This is to prevent predict from predicting past first completed term

	if len(a.All) > 0 && a.LastCompleted != "" {
		index := -1
		for i, arg := range a.All {
			if arg == a.LastCompleted {
				index = i - 1
			}
		}
		if index >= 0 {
			oneBefore := a.All[index]
			if oneBefore == "--path" || oneBefore == "read" {
				return []string{}
			}
		}
	}

	token := viper.Get("token")
	if token == nil {
		if tr, err := p.getAuthenticator().GetToken(); err != nil || tr == nil || tr.Token == "" {
			return []string{}
		} else {
			token = tr.Token
			viper.Set("token", token)
		}

	}
	uri := paths.CreateResourceURI(cst.NounSecret, a.Last, cst.SuffixListPaths, true, nil, true)
	res := PathAutocompleteResult{}
	if err := p.getRequestClient().DoRequestOut("GET", uri, nil, &res); err == nil {
		return preparePaths(res.Data)
	}
	return []string{}
}

func preparePaths(unprepared []string) []string {
	prepared := []string{}
	for _, p := range unprepared {
		friendlyTail := paths.GetURIPathFromInternalPath(p)
		if len(friendlyTail) > 0 {
			prepared = append(prepared, friendlyTail)
		}
	}
	return prepared
}

type AuthTypePredictor struct{}

func (p AuthTypePredictor) Predict(a complete.Args) (prediction []string) {
	return []string{
		string(auth.Password),
		string(auth.ClientCredential),
		string(auth.Certificate),
		string(auth.FederatedAws),
		string(auth.FederatedAzure),
		string(auth.FederatedGcp),
	}
}

type ActionTypePredictor struct{}

func (p ActionTypePredictor) Predict(a complete.Args) (prediction []string) {
	return []string{
		"share",
		"create",
		"update",
		"delete",
		"read",
	}
}

type AuthProviderTypePredictor struct{}

func (p AuthProviderTypePredictor) Predict(a complete.Args) (prediction []string) {
	return []string{
		"aws",
		"azure",
		"gcp",
		"oidc",
	}
}

type EffectTypePredictor struct{}

func (p EffectTypePredictor) Predict(a complete.Args) (prediction []string) {
	return []string{
		"allow",
		"deny",
	}
}

type PrefixFilePredictor struct {
	filePattern string
}

func NewPrefixFilePredictor(pattern string) *PrefixFilePredictor {
	return &PrefixFilePredictor{filePattern: pattern}
}

func (p PrefixFilePredictor) Predict(a complete.Args) (prediction []string) {
	last := a.Last
	if !strings.HasPrefix(last, cst.CmdFilePrefix) {
		return []string{}
	}

	allowFiles := true
	prediction = predictFiles(a, p.filePattern, allowFiles, cst.CmdFilePrefix)

	// if the number of prediction is not 1, we either have many results or
	// have no results, so we return it.
	if len(prediction) != 1 {
		return
	}

	// only try deeper, if the one item is a directory
	if stat, err := os.Stat(prediction[0]); err != nil || !stat.IsDir() {
		return
	}

	a.Last = prediction[0]
	return predictFiles(a, p.filePattern, allowFiles, cst.CmdFilePrefix)
}

func predictFiles(a complete.Args, pattern string, allowFiles bool, prefix string) []string {
	if strings.HasSuffix(a.Last, "/..") {
		return nil
	}

	argsWithRealPath := complete.Args{
		All:           a.All,
		Completed:     a.Completed,
		Last:          a.Last[len(prefix):],
		LastCompleted: a.LastCompleted,
	}
	dir := argsWithRealPath.Directory()
	files := listFiles(dir, pattern, allowFiles, prefix)

	// add dir if match
	files = append(files, dir)

	return complete.PredictFilesSet(files).Predict(a)
}

func listFiles(dir, pattern string, allowFiles bool, prefix string) []string {
	// set of all file names
	m := map[string]bool{}

	// list files
	if files, err := filepath.Glob(filepath.Join(dir, pattern)); err == nil {
		for _, f := range files {
			if stat, err := os.Stat(f); err != nil || stat.IsDir() || allowFiles {
				m[f] = true
			}
		}
	}

	// list directories
	if dirs, err := ioutil.ReadDir(dir); err == nil {
		for _, d := range dirs {
			if d.IsDir() {
				m[filepath.Join(dir, d.Name())] = true
			}
		}
	}

	list := make([]string, 0, len(m))
	for k := range m {
		list = append(list, prefix+k)
	}
	return list
}

type OutputTypePredictor struct{}

// Switch out so we don't predict file paths. If this is to be added back in, need to change
// constants.OutVarPrefix and constants.OutFilePrefix to end with * or , ([:!|=$] do not work)
// func (p OutputTypePredictor) Predict(a complete.Args) (prediction []string) {
// 	last := a.Last
// 	if !strings.HasPrefix(last, cst.OutVarPrefix) && !strings.HasPrefix(last, cst.OutFilePrefix) {
// 		return []string{string(format.StdOut), string(format.ClipBoard), cst.OutFilePrefix, cst.OutVarPrefix}
// 	} else if strings.HasPrefix(last, cst.OutFilePrefix) {
// 		allowFiles := true
// 		prediction = predictFiles(a, "*", allowFiles, cst.OutFilePrefix)

// 		// if the number of prediction is not 1, we either have many results or
// 		// have no results, so we return it.
// 		if len(prediction) != 1 {
// 			return
// 		}

// 		// only try deeper, if the one item is a directory
// 		if stat, err := os.Stat(prediction[0]); err != nil || !stat.IsDir() {
// 			return
// 		}
// 		a.Last = prediction[0]
// 		return predictFiles(a, "*", allowFiles, cst.OutFilePrefix)
// 	}
// 	// env variable name
// 	return []string{}
// }

func (p OutputTypePredictor) Predict(a complete.Args) (prediction []string) {
	return []string{string(format.StdOut), string(format.ClipBoard), cst.OutFilePrefix}
}

type GcpAuthTypePredictor struct{}

func (p GcpAuthTypePredictor) Predict(a complete.Args) (prediction []string) {
	return []string{string(auth.GcpGceAuth), string(auth.GcpIamAuth)}
}

type EncodingTypePredictor struct{}

func (p EncodingTypePredictor) Predict(a complete.Args) (prediction []string) {
	return []string{cst.Json, cst.YamlShort}
}

type PermissionTypePredictor struct{}

func (p PermissionTypePredictor) Predict(a complete.Args) (prediction []string) {
	return []string{cst.NounSecret, cst.NounRole, cst.NounUser, cst.NounClient}
}
