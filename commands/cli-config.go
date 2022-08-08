package cmd

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"thy/auth"
	cst "thy/constants"
	"thy/errors"
	"thy/internal/predictor"
	"thy/paths"
	"thy/store"
	"thy/vaultcli"

	"thy/utils"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

func GetCliConfigCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:      []string{cst.NounCliConfig},
		HelpText:  "Execute an action on the cli config for " + cst.ProductName,
		NoPreAuth: true,
		RunFunc: func(args []string) int {
			return cli.RunResultHelp
		},
	})
}

func GetCliConfigInitCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounCliConfig, cst.Init},
		SynopsisText: strings.Join([]string{cst.NounCliConfig, cst.Init}, " "),
		HelpText:     "Initialize the cli config for " + cst.ProductName,
		NoPreAuth:    true,
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Dev, Hidden: true, Usage: "Specify dev domain upon initialization"},
		},
		RunFunc: func(args []string) int {
			return handleCliConfigInitCmd(vaultcli.New(), args)
		},
	})
}

func GetCliConfigUpdateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounCliConfig, cst.Update},
		SynopsisText: fmt.Sprintf("%s %s (<key> <value> | --key --value)", cst.NounCliConfig, cst.Update),
		HelpText: `Update a cli config setting for the specified profile. The key specifies the path to that setting.

Usage:
   • cli-config update --profile default --key auth.password --value *******
   • cli-config update profile2.auth.type clientcred --profile profile2
		`,
		NoPreAuth: true,
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Key, Usage: "Key of setting to be updated (required)"},
			{Name: cst.Value, Usage: "Value of setting to be udpated (required)"},
		},
		MinNumberArgs: 2,
		RunFunc: func(args []string) int {
			return handleCliConfigUpdateCmd(vaultcli.New(), args)
		},
	})
}

func GetCliConfigClearCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounCliConfig, cst.Clear},
		SynopsisText: strings.Join([]string{cst.NounCliConfig, cst.Clear}, " "),
		HelpText:     "Clear the cli config for " + cst.ProductName,
		NoPreAuth:    true,
		RunFunc: func(args []string) int {
			return handleCliConfigClearCmd(vaultcli.New(), args)
		},
	})
}

func GetCliConfigReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounCliConfig, cst.Read},
		SynopsisText: strings.Join([]string{cst.NounCliConfig, cst.Read}, " "),
		HelpText:     "Read the cli config for " + cst.ProductName,
		NoPreAuth:    true,
		RunFunc: func(args []string) int {
			return handleCliConfigReadCmd(vaultcli.New(), args)
		},
	})
}

func GetCliConfigEditCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounCliConfig, cst.Edit},
		SynopsisText: strings.Join([]string{cst.NounCliConfig, cst.Edit}, " "),
		HelpText:     "Edit the cli config for " + cst.ProductName,
		NoPreAuth:    true,
		RunFunc: func(args []string) int {
			return handleCliConfigEditCmd(vaultcli.New(), args)
		},
	})
}

func handleCliConfigUpdateCmd(vcli vaultcli.CLI, args []string) int {
	key := strings.TrimSpace(viper.GetString(cst.Key))
	val := strings.TrimSpace(viper.GetString(cst.Value))
	if key == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		key = args[0]
	}
	if val == "" && len(args) > 1 && !strings.HasPrefix(args[1], "-") {
		// Make value the same as the key name, as in "user:user", for example.
		val = args[1]
	}

	profile := viper.GetString(cst.Profile)
	if profile == "" {
		profile = cst.DefaultProfile
	}

	if key == cst.Password || key == cst.AuthClientSecret {
		storeType := viper.GetString(cst.StoreType)
		if storeType == store.PassLinux || storeType == store.WinCred {
			key = profile + "." + key
			err := store.StoreSecureSetting(key, val, storeType)
			if err != nil {
				vcli.Out().FailF("Error updating setting in store of type '%s'", string(storeType))
				return 1
			}
			return 0
		}
	}

	cfgPath := viper.GetString(cst.Config)
	cf, readErr := vaultcli.ReadConfigFile(cfgPath)
	if readErr != nil {
		vcli.Out().FailF("Error: failed to read config: %v.", readErr)
		return 1
	}

	prf, ok := cf.GetProfile(profile)
	if !ok {
		vcli.Out().FailF("profile %q not found in configuration file %q", profile, cf.GetPath())
		return 1
	}

	secure := "auth.securePassword"
	if strings.HasSuffix(key, secure) {
		vcli.Out().FailF("Cannot manually change the value of %q.", secure)
		return 1
	}

	if strings.HasSuffix(key, "auth.password") {
		key = secure
		cipherText, err := auth.EncipherPassword(val)
		if err != nil {
			vcli.Out().FailF("Error encrypting password: %v.", err)
			return 1
		}
		// Set in-memory value of plaintext password, so that the CLI can use it to try to authenticate before writing the config.
		viper.Set(cst.Password, val)
		if authError := tryAuthenticate(); authError != nil {
			vcli.Out().FailS("Failed to authenticate, restoring previous config.")
			vcli.Out().FailS("Please check your credentials and try again.")
			return 1
		}
		val = cipherText
	}

	keys := strings.Split(key, ".")
	if val != "" && val != "0" {
		prf.Set(val, keys...)
	} else {
		prf.Del(keys...)
	}
	cf.UpdateProfile(prf)
	saveErr := cf.Save()
	if saveErr != nil {
		vcli.Out().FailF("Error: cannot save configuration: %v", saveErr)
		return 1
	}
	return 0
}

func handleCliConfigReadCmd(vcli vaultcli.CLI, args []string) int {
	var didError int
	var dataOut []byte

	cfgPath := viper.GetString(cst.Config)
	cf, err := vaultcli.ReadConfigFile(cfgPath)
	if err != nil {
		vcli.Out().FailF("Error: failed to read config: %v", err)
		didError = 1
	} else {
		dataOut = []byte(fmt.Sprintf("CLI config (%s):\n", cf.GetPath()))
		dataOut = append(dataOut, cf.Bytes()...)
	}

	storeType := viper.GetString(cst.StoreType)
	if storeType == store.PassLinux || storeType == store.WinCred {
		dataOut = append(dataOut, fmt.Sprintf("\nSecure Store Settings (store type: %s):\n", storeType)...)
		if s, err := store.GetStore(string(storeType)); err != nil {
			vcli.Out().FailF("Error: failed to get store of type '%s'\n", string(storeType))
			didError = 1
		} else if keys, err := s.List(cst.CliConfigRoot); err != nil {
			vcli.Out().FailF("Error: failed to get store of type '%s'\n", string(storeType))
			didError = 1
		} else {
			if len(keys) > 0 {
				dataOut = append(dataOut, []byte("  ")...)
			}
			dataOut = append(dataOut, strings.ReplaceAll(strings.Join(keys, "\n  "), "-", ".")...)
		}
	}

	vcli.Out().WriteResponse(dataOut, nil)
	return didError
}

func handleCliConfigEditCmd(vcli vaultcli.CLI, args []string) int {
	cfgPath := viper.GetString(cst.Config)
	cf, readErr := vaultcli.ReadConfigFile(cfgPath)
	if readErr != nil {
		vcli.Out().FailF("Error: failed to read config: %v", readErr)
		return 1
	}

	saveFunc := func(data []byte) (resp []byte, err *errors.ApiError) {
		updErr := cf.RawUpdate(data)
		if updErr != nil {
			return nil, errors.New(updErr)
		}
		return nil, errors.New(cf.Save())
	}

	_, err := vcli.Edit(cf.Bytes(), saveFunc)

	vcli.Out().WriteResponse(nil, err)
	return utils.GetExecStatus(err)
}

func handleCliConfigClearCmd(vcli vaultcli.CLI, args []string) int {
	var yes bool
	areYouSurePrompt := &survey.Confirm{Message: "Are you sure you want to delete CLI configuration?", Default: false}
	survErr := survey.AskOne(areYouSurePrompt, &yes)
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}
	if !yes {
		vcli.Out().WriteResponse([]byte("Exiting."), nil)
		return 0
	}

	didError := 0
	cfgPath := viper.GetString(cst.Config)
	deleteErr := vaultcli.DeleteConfigFile(cfgPath)
	if deleteErr != nil {
		vcli.Out().FailF("Error deleting CLI config: %v.", deleteErr)
		didError = 1
	}

	st := viper.GetString(cst.StoreType)
	s, err := store.GetStore(st)
	if err != nil {
		vcli.Out().FailF("Failed to get store: %v.", err)
		return 1
	}
	err = s.Wipe("")
	if err != nil {
		vcli.Out().FailF("Failed to clear store: %v.", err)
		return 1
	}

	return didError
}

func handleCliConfigInitCmd(vcli vaultcli.CLI, args []string) int {
	ui := cli.BasicUi{
		Writer:      os.Stdout,
		Reader:      os.Stdin,
		ErrorWriter: os.Stderr,
	}

	cfgPath := viper.GetString(cst.Config)
	if cfgPath != "" {
		// The `dsv init` command is the only command that allows path to a directory in `--config` flag.
		info, err := os.Stat(cfgPath)
		if err == nil && info != nil && info.IsDir() {
			cfgPath = filepath.Join(cfgPath, cst.CliConfigName)
		}
	}

	cf, err := vaultcli.NewConfigFile(cfgPath)
	if err != nil {
		vcli.Out().FailF("Error: %v.", err)
		return 1
	}

	err = cf.Read()
	if err != nil && err != vaultcli.ErrFileNotFound {
		vcli.Out().FailF("Error: could not read file at path %q: %v.", cf.GetPath(), err)
		return 1
	}
	cfgExists := err == nil
	if cfgExists {
		ui.Warn(fmt.Sprintf("Found an existing cli-config located at '%s'", cf.GetPath()))
	}

	profile := strings.ToLower(viper.GetString(cst.Profile))
	if !cfgExists {
		if profile != "" && profile != cst.DefaultProfile {
			// If config does not exist, and the user specified a non-default --profile name, then quit and ask to properly init.
			ui.Info("Initial configuration is needed in order to add a custom profile.")
			ui.Info(fmt.Sprintf("Create CLI config file manually or execute command '%s init' to initiate CLI configuration.", cst.CmdRoot))
			return 1
		}
		profile = cst.DefaultProfile
	} else {
		if profile != "" {
			// The user specified --profile [name], so the intent is to add a profile to the config.

			if err := vaultcli.IsValidProfile(profile); err != nil {
				ui.Info(err.Error())
				return 1
			}
			if _, ok := cf.GetProfile(profile); ok {
				msg := fmt.Sprintf("Profile %q already exists in the config.", profile)
				ui.Info(msg)
				return 1
			}
		} else {
			var actionID int
			actionPrompt := &survey.Select{
				Message: "Select an option:",
				Options: []string{
					"Do nothing",
					"Overwrite the config",
					"Add a new profile to the config",
				},
			}
			survErr := survey.AskOne(actionPrompt, &actionID)
			if survErr != nil {
				vcli.Out().WriteResponse(nil, errors.New(survErr))
				return 1
			}
			switch actionID {
			case 0: // "Do nothing".
				ui.Info("Exiting.")
				return 0

			case 1: // "Overwrite the config".
				cf, _ = vaultcli.NewConfigFile(cf.GetPath())
				profile = cst.DefaultProfile

			case 2: // "Add a new profile to the config".
				profilePrompt := &survey.Input{Message: "Please enter profile name:"}
				profileValidation := func(ans interface{}) error {
					if err := vaultcli.SurveyRequired(ans); err != nil {
						return err
					}
					answer := strings.TrimSpace(ans.(string))
					if err := vaultcli.IsValidProfile(answer); err != nil {
						return errors.New(err)
					}
					_, ok := cf.GetProfile(answer)
					if ok {
						return errors.NewS("Profile with this name already exists in the config.")
					}
					return nil
				}
				survErr := survey.AskOne(profilePrompt, &profile, survey.WithValidator(profileValidation))
				if survErr != nil {
					vcli.Out().WriteResponse(nil, errors.New(survErr))
					return 1
				}
				profile = strings.TrimSpace(profile)
			}
		}
	}
	prf := vaultcli.NewProfile(profile)
	viper.Set(cst.Profile, profile)

	// tenant
	var tenant string
	tenantPrompt := &survey.Input{Message: "Please enter tenant name:"}
	survErr := survey.AskOne(tenantPrompt, &tenant, survey.WithValidator(vaultcli.SurveyRequired))
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}
	tenant = strings.TrimSpace(tenant)

	prf.Set(tenant, cst.Tenant)
	viper.Set(cst.Tenant, tenant)

	// domain
	var domain string
	var isDevDomain bool

	if devDomain := viper.GetString(cst.Dev); devDomain != "" {
		isDevDomain = true
		domain = devDomain
	} else {
		domain, err = promptDomain()
		if err != nil {
			vcli.Out().WriteResponse(nil, errors.New(err))
			return 1
		}
	}
	prf.Set(domain, cst.DomainKey)
	viper.Set(cst.DomainKey, domain)

	// check if tenant has been setup yet
	var setupRequired bool
	heartbeatURI := paths.CreateURI("heartbeat", nil)
	if respData, err := vcli.HTTPClient().DoRequest(http.MethodGet, heartbeatURI, nil); err != nil {
		ui.Error(fmt.Sprintf("Failed to contact %s to determine if initial admin setup is required.", cst.ProductName))
		return 1
	} else {
		var resp heartbeatResponse
		if err := json.Unmarshal(respData, &resp); err != nil {
			ui.Error(fmt.Sprintf("Failed to read the response from %s to determine if initial admin setup is required.", cst.ProductName))
			return 1
		} else if resp.StatusCode == 2 {
			setupRequired = true
		}
	}

	// store
	storeType, err := promptStoreType()
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(err))
		return 1
	}
	if err := store.ValidateCredentialStore(storeType); err != nil {
		ui.Error(fmt.Sprintf("Failed to get store: %v.", err))
		return 1
	}
	isSecureStore := storeType == store.PassLinux || storeType == store.WinCred

	prf.Set(storeType, cst.Store, cst.Type)
	viper.Set(cst.StoreType, storeType)

	if storeType == store.File {
		def := filepath.Join(utils.NewEnvProvider().GetHomeDir(), ".thy")

		var fileStorePath string
		fileStorePathPrompt := &survey.Input{
			Message: "Please enter directory for file store:",
			Default: def,
		}
		survErr := survey.AskOne(fileStorePathPrompt, &fileStorePath, survey.WithValidator(vaultcli.SurveyRequired))
		if survErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(survErr))
			return utils.GetExecStatus(survErr)
		}
		fileStorePath = strings.TrimSpace(fileStorePath)

		if fileStorePath != def {
			prf.Set(fileStorePath, cst.Store, cst.Path)
			viper.Set(cst.StorePath, fileStorePath)
		}
	}

	//cache
	if storeType != store.None {
		cacheStrategy, err := promptCacheStrategy()
		if err != nil {
			vcli.Out().WriteResponse(nil, errors.New(err))
			return 1
		}

		prf.Set(cacheStrategy, cst.Cache, cst.Strategy)

		if cacheStrategy != cst.CacheStrategyNever {
			var cacheAge string
			cacheAgePrompt := &survey.Input{Message: "Please enter cache age (minutes until expiration):"}
			cacheAgeValidation := func(ans interface{}) error {
				if err := vaultcli.SurveyRequired(ans); err != nil {
					return err
				}
				answer := strings.TrimSpace(ans.(string))
				if age, err := strconv.Atoi(answer); err != nil || age <= 0 {
					return errors.NewS("Unable to parse age. Must be strictly positive int")
				}
				return nil
			}
			survErr := survey.AskOne(cacheAgePrompt, &cacheAge, survey.WithValidator(cacheAgeValidation))
			if survErr != nil {
				vcli.Out().WriteResponse(nil, errors.New(survErr))
				return utils.GetExecStatus(survErr)
			}
			cacheAge = strings.TrimSpace(cacheAge)
			prf.Set(cacheAge, cst.Cache, cst.Age)
		}
	}

	//auth
	// allow overriding option with flag
	authType := viper.GetString(cst.AuthType)
	if authType == "" {
		authType, err = promptAuthType()
		if err != nil {
			vcli.Out().WriteResponse(nil, errors.New(err))
			return 1
		}
		viper.Set(cst.AuthType, authType)
	}
	prf.Set(authType, cst.NounAuth, cst.Type)

	var user, password, passwordKey, authProvider, encryptionKeyFileName string
	authProvider = viper.GetString(cst.AuthProvider)

	switch {
	case (storeType != store.None && auth.AuthType(authType) == auth.Password):
		if setupRequired {
			user, password, err = promptInitialUsernamePassword(tenant)
		} else {
			user, password, err = promptUsernamePassword(tenant)
		}
		if err != nil {
			vcli.Out().WriteResponse(nil, errors.New(err))
			return 1
		}

		prf.Set(user, cst.NounAuth, cst.DataUsername)
		viper.Set(cst.Username, user)
		viper.Set(cst.Password, password)

		encryptionKeyFileName = auth.GetEncryptionKeyFilename(viper.GetString(cst.Tenant), user)

		if isSecureStore {
			if err := store.StoreSecureSetting(strings.Join([]string{profile, cst.NounAuth, cst.DataPassword}, "."), password, storeType); err != nil {
				ui.Error(err.Error())
				return 1
			}
		} else {
			encrypted, key, err := auth.StorePassword(encryptionKeyFileName, password)
			if err != nil {
				ui.Error(err.Error())
				return 1
			}
			passwordKey = key
			prf.Set(encrypted, cst.NounAuth, cst.DataSecurePassword)
		}

	case (storeType != store.None && auth.AuthType(authType) == auth.ClientCredential):
		clientID, clientSecret, err := promptClientCredentials()
		if err != nil {
			vcli.Out().WriteResponse(nil, errors.New(err))
			return 1
		}

		prf.Set(clientID, cst.NounAuth, cst.NounClient, cst.ID)
		viper.Set(cst.AuthClientID, clientID)

		if isSecureStore {
			if err := store.StoreSecureSetting(strings.Join([]string{profile, cst.NounAuth, cst.NounClient, cst.NounSecret}, "."), clientSecret, storeType); err != nil {
				ui.Error(err.Error())
				return 1
			}
		} else {
			prf.Set(clientSecret, cst.NounAuth, cst.NounClient, cst.NounSecret)
			viper.Set(cst.AuthClientSecret, clientSecret)
		}

	case auth.AuthType(authType) == auth.FederatedAws:
		var awsProfile string
		awsProfilePrompt := &survey.Input{
			Message: "Please enter aws profile for federated aws auth:",
			Default: "default",
		}
		survErr := survey.AskOne(awsProfilePrompt, &awsProfile, survey.WithValidator(vaultcli.SurveyRequired))
		if survErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(survErr))
			return utils.GetExecStatus(survErr)
		}
		awsProfile = strings.TrimSpace(awsProfile)
		prf.Set(awsProfile, cst.NounAuth, cst.NounAwsProfile)
		viper.Set(cst.AwsProfile, awsProfile)

	case auth.AuthType(authType) == auth.Oidc || auth.AuthType(authType) == auth.FederatedThyOne:
		if auth.AuthType(authType) == auth.Oidc {
			if authProvider == "" {
				authProviderPrompt := &survey.Input{
					Message: "Please enter auth provider name:",
					Default: cst.DefaultThyOneName,
				}
				survErr := survey.AskOne(authProviderPrompt, &authProvider, survey.WithValidator(vaultcli.SurveyRequired))
				if survErr != nil {
					vcli.Out().WriteResponse(nil, errors.New(survErr))
					return utils.GetExecStatus(survErr)
				}
				authProvider = strings.TrimSpace(authProvider)
			}
		} else {
			authProvider = cst.DefaultThyOneName

			if isDevDomain {
				authProviderPrompt := &survey.Input{
					Message: "Thycotic One authentication provider name:",
					Default: cst.DefaultThyOneName,
				}
				survErr := survey.AskOne(authProviderPrompt, &authProvider, survey.WithValidator(vaultcli.SurveyRequired))
				if survErr != nil {
					vcli.Out().WriteResponse(nil, errors.New(survErr))
					return utils.GetExecStatus(survErr)
				}
				authProvider = strings.TrimSpace(authProvider)
			}
		}

		var callback string
		if callback = viper.GetString(cst.Callback); callback == "" {
			callback = cst.DefaultCallback
		}

		prf.Set(authProvider, cst.NounAuth, cst.DataProvider)
		prf.Set(callback, cst.NounAuth, cst.DataCallback)

		viper.Set(cst.AuthProvider, authProvider)
		viper.Set(cst.Callback, callback)

	case auth.AuthType(authType) == auth.Certificate:
		clientCert, err := promptCertificate()
		if err != nil {
			vcli.Out().WriteResponse(nil, errors.New(err))
			return 1
		}
		clientPrivKey, err := promptPrivateKey()
		if err != nil {
			vcli.Out().WriteResponse(nil, errors.New(err))
			return 1
		}

		prf.Set(clientCert, cst.NounAuth, cst.NounCert)
		prf.Set(clientPrivKey, cst.NounAuth, cst.NounPrivateKey)

		viper.Set(cst.AuthCert, clientCert)
		viper.Set(cst.AuthPrivateKey, clientPrivKey)
	}

	if setupRequired {
		initializeURI := paths.CreateURI("initialize", nil)
		body := initializeRequest{
			UserName: user,
			Password: password,
		}
		if _, err := vcli.HTTPClient().DoRequest(http.MethodPost, initializeURI, body); err != nil {
			ui.Error(fmt.Sprintf("Failed to initialize tenant with %s. Please try again. Error:", cst.ProductName))
			vcli.Out().FailE(err)
			return 1
		}
	}

	if storeType != store.None {
		if authError := tryAuthenticate(); authError != nil {
			if isAccountLocked(authError) {
				ui.Output(authError.Error())
				return 1
			}

			ui.Output("Failed to authenticate, restoring previous config.")
			ui.Output("Please check your credentials, or tenant name, or domain name and try again.")
			return 1
		}

		// Store encryption key file (for auth type password).
		if auth.AuthType(authType) == auth.Password {
			st, apiError := store.GetStore(storeType)
			if apiError != nil {
				ui.Error(apiError.Error())
				return 1
			}
			apiError = st.StoreString(encryptionKeyFileName, passwordKey)
			if apiError != nil {
				ui.Error(apiError.Error())
				return 1
			}
		}
	} else {
		ui.Output("Authentication parameters are not checked since 'None (no caching)' was selected for the store type.")
		ui.Output("Each command that calls DSV API will trigger authentication first to get access token.")

		switch auth.AuthType(authType) {
		case auth.Password:
			ui.Output(`To authenticate using username and password use '--auth-username' and '--auth-password' flags.

Example:
	dsv secret search --auth-username "example-username" --auth-password "example-password"`)

		case auth.ClientCredential:
			ui.Output(`To authenticate using client credentials use '--auth-client-id' and '--auth-client-secret' flags.

Example:
	dsv secret search --auth-client-id "a71d...f0d4" --auth-client-secret "R8WzW...jWg"`)

		}
	}

	cf.AddProfile(prf)
	saveErr := cf.Save()
	if saveErr != nil {
		vcli.Out().FailF("Error: could not save configuration at path %q: %v.", cf.GetPath(), err)
		return 1
	}

	ui.Output("\nCLI configuration file successfully saved.")
	return 0
}

// tryAuthenticate attempts to authenticate with the current state of all constants, such as auth type, username, password, etc.
func tryAuthenticate() error {
	authenticator := auth.NewAuthenticatorDefault()
	_, apiError := authenticator.GetToken()
	if apiError != nil {
		return apiError
	}
	return nil
}

func isAccountLocked(err error) bool {
	return strings.Contains(err.Error(), "locked out")
}

type heartbeatResponse struct {
	StatusCode int
	Message    string
}

type initializeRequest struct {
	UserName string
	Password string
}

func promptDomain() (string, error) {
	var domain string
	domainPrompt := &survey.Select{
		Message: "Please choose domain:",
		Options: []string{
			cst.Domain,
			cst.DomainEU,
			cst.DomainAU,
			cst.DomainCA,
		},
	}
	survErr := survey.AskOne(domainPrompt, &domain)
	if survErr != nil {
		return "", survErr
	}
	return domain, nil
}

func promptStoreType() (string, error) {
	var storeTypeID int
	storeTypePrompt := &survey.Select{
		Message: "Please select store type:",
		Options: []string{
			"File store",
			"None (no caching)",
			"Pass (Linux only)",
			"Windows Credential Manager (Windows only)",
		},
	}
	survErr := survey.AskOne(storeTypePrompt, &storeTypeID)
	if survErr != nil {
		return "", survErr
	}
	switch storeTypeID {
	case 0:
		return store.File, nil
	case 1:
		return store.None, nil
	case 2:
		return store.PassLinux, nil
	case 3:
		return store.WinCred, nil
	default:
		return "", errors.NewF("Unhandled case for store type id %d", storeTypeID)
	}
}

func promptCacheStrategy() (string, error) {
	var cacheStrategyID int
	cacheStrategyPrompt := &survey.Select{
		Message: "Please enter cache strategy for secrets:",
		Options: []string{
			"Never",
			"Server then cache",
			"Cache then server",
			"Cache then server, but allow expired cache if server unreachable",
		},
	}
	survErr := survey.AskOne(cacheStrategyPrompt, &cacheStrategyID)
	if survErr != nil {
		return "", survErr
	}
	switch cacheStrategyID {
	case 0:
		return cst.CacheStrategyNever, nil
	case 1:
		return cst.CacheStrategyServerThenCache, nil
	case 2:
		return cst.CacheStrategyCacheThenServer, nil
	case 3:
		return cst.CacheStrategyCacheThenServerThenExpired, nil
	default:
		return "", errors.NewF("Unhandled case for cache strategy id %d", cacheStrategyID)
	}
}

func promptAuthType() (string, error) {
	var authTypeID int
	authTypePrompt := &survey.Select{
		Message: "Please enter auth type:",
		Options: []string{
			"Password (local user)",
			"Client Credential",
			"Thycotic One (federated)",
			"AWS IAM (federated)",
			"Azure (federated)",
			"GCP (federated)",
			"OIDC (federated)",
			"x509 Certificate",
		},
		PageSize: 8,
	}
	survErr := survey.AskOne(authTypePrompt, &authTypeID)
	if survErr != nil {
		return "", survErr
	}
	switch authTypeID {
	case 0:
		return string(auth.Password), nil
	case 1:
		return string(auth.ClientCredential), nil
	case 2:
		return string(auth.FederatedThyOne), nil
	case 3:
		return string(auth.FederatedAws), nil
	case 4:
		return string(auth.FederatedAzure), nil
	case 5:
		return string(auth.FederatedGcp), nil
	case 6:
		return string(auth.Oidc), nil
	case 7:
		return string(auth.Certificate), nil
	default:
		return "", errors.NewF("Unhandled case for auth type id %d", authTypeID)
	}
}

func promptUsernamePassword(tenant string) (string, string, error) {
	answers := struct {
		Username string
		Password string
	}{}

	qs := []*survey.Question{
		{
			Name: "Username",
			Prompt: &survey.Input{
				Message: fmt.Sprintf("Please enter username for tenant %q:", tenant),
			},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:   "Password",
			Prompt: &survey.Password{Message: "Please enter password:"},
		},
	}

	survErr := survey.Ask(qs, &answers)
	if survErr != nil {
		return "", "", survErr
	}
	return answers.Username, answers.Password, nil
}

func promptInitialUsernamePassword(tenant string) (string, string, error) {
	answers := struct {
		Username string
		Password string
		Confirm  string
	}{}

	qs := []*survey.Question{
		{
			Name: "Username",
			Prompt: &survey.Input{
				Message: fmt.Sprintf("Please choose a username for initial local admin for tenant %q:", tenant),
			},
			Validate:  vaultcli.SurveyRequired,
			Transform: vaultcli.SurveyTrimSpace,
		},
		{
			Name:   "Password",
			Prompt: &survey.Password{Message: "Please choose password:"},
		},
		{
			Name: "Confirm",
			Validate: func(ans interface{}) error {
				if answers.Password != ans.(string) {
					return errors.NewS("Passwords do not match.")
				}
				return nil
			},
			Prompt: &survey.Password{Message: "Please choose password (confirm):"},
		},
	}

	survErr := survey.Ask(qs, &answers)
	if survErr != nil {
		return "", "", survErr
	}
	return answers.Username, answers.Password, nil
}

func promptClientCredentials() (string, string, error) {
	answers := struct {
		ClientID     string
		ClientSecret string
	}{}

	qs := []*survey.Question{
		{
			Name:     "ClientID",
			Prompt:   &survey.Input{Message: "Please enter client id for client auth:"},
			Validate: vaultcli.SurveyRequired,
		},
		{
			Name:     "ClientSecret",
			Prompt:   &survey.Password{Message: "Please enter client secret for client auth:"},
			Validate: vaultcli.SurveyRequired,
		},
	}

	survErr := survey.Ask(qs, &answers)
	if survErr != nil {
		return "", "", survErr
	}
	return answers.ClientID, answers.ClientSecret, nil
}

type cliCertificateOutput struct {
	Certificate  string `json:"certificate"`
	PrivateKey   string `json:"privateKey"`
	SSHPublicKey string `json:"sshPublicKey"`
}

func promptCertificate() (string, error) {
	var actionID int
	actionPrompt := &survey.Select{
		Message: "Select an option:",
		Options: []string{
			"Raw certificate",
			"Certificate file path",
		},
	}
	survErr := survey.AskOne(actionPrompt, &actionID)
	if survErr != nil {
		return "", survErr
	}
	answer := struct {
		Cert string
	}{}
	switch actionID {
	case 0: // "Raw certificate".
		qs := []*survey.Question{
			{
				Name:     "Cert",
				Prompt:   &survey.Input{Message: "Raw certificate:"},
				Validate: vaultcli.SurveyRequired,
			},
		}
		survErr := survey.Ask(qs, &answer)
		if survErr != nil {
			return "", survErr
		}
		return parseBase64EncodedCertificate(answer.Cert)

	case 1: // "Certificate file path".
		qs := []*survey.Question{
			{
				Name:     "Cert",
				Prompt:   &survey.Input{Message: "Certificate file path:"},
				Validate: vaultcli.SurveyRequired,
			},
		}
		survErr := survey.Ask(qs, &answer)
		if survErr != nil {
			return "", survErr
		}
		fileData, err := os.ReadFile(answer.Cert)
		if err != nil {
			return "", fmt.Errorf("'%s' file cannot be read", answer.Cert)
		}
		// pem encoded certificate
		if isPem, err := checkIsPem(fileData, "CERTIFICATE"); isPem {
			return base64.StdEncoding.EncodeToString(fileData), nil
		} else if err != nil {
			return "", err
		}
		// der encoded certificate
		cert, err := x509.ParseCertificate(fileData)
		if err == nil {
			out := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
			return base64.StdEncoding.EncodeToString(out), nil
		}
		// cli json output
		body := cliCertificateOutput{}
		if err := json.Unmarshal(fileData, &body); err == nil {
			return parseBase64EncodedCertificate(body.Certificate)
		}
		// base64 encoded data
		res, err := parseBase64EncodedCertificate(string(fileData))
		if err != nil {
			return "", fmt.Errorf("unknown certificate format. Please try .pem/.der/.json")
		}
		return res, nil
	}
	return "", fmt.Errorf("undefined option")
}

func promptPrivateKey() (string, error) {
	var actionID int
	actionPrompt := &survey.Select{
		Message: "Select an option:",
		Options: []string{
			"Raw private key",
			"Private key file path",
		},
	}
	survErr := survey.AskOne(actionPrompt, &actionID)
	if survErr != nil {
		return "", survErr
	}
	answer := struct {
		PrivKey string
	}{}
	switch actionID {
	case 0: // "Raw private key".
		qs := []*survey.Question{
			{
				Name:     "PrivKey",
				Prompt:   &survey.Password{Message: "Private key:"},
				Validate: vaultcli.SurveyRequired,
			},
		}
		survErr := survey.Ask(qs, &answer)
		if survErr != nil {
			return "", survErr
		}
		content, err := base64.StdEncoding.DecodeString(answer.PrivKey)
		if err != nil {
			return "", fmt.Errorf("private key must be base64 encoded")
		}
		key, err := x509.ParsePKCS1PrivateKey(content)
		if err == nil {
			keyBytes := x509.MarshalPKCS1PrivateKey(key)
			out := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes})
			return base64.StdEncoding.EncodeToString(out), nil
		}
		return parseBase64EncodedPrivKey(answer.PrivKey)

	case 1: // "Private key file path".
		qs := []*survey.Question{
			{
				Name:     "PrivKey",
				Prompt:   &survey.Input{Message: "Private key file path:"},
				Validate: vaultcli.SurveyRequired,
			},
		}
		survErr := survey.Ask(qs, &answer)
		if survErr != nil {
			return "", survErr
		}
		fileData, err := os.ReadFile(answer.PrivKey)
		if err != nil {
			return "", fmt.Errorf("'%s' file cannot be read", answer.PrivKey)
		}
		// pem encoded rsa private key
		if isPem, err := checkIsPem(fileData, "RSA PRIVATE KEY"); isPem {
			return base64.StdEncoding.EncodeToString(fileData), nil
		} else if err != nil {
			return "", err
		}
		// der encoded rsa private key
		key, err := x509.ParsePKCS1PrivateKey(fileData)
		if err == nil {
			keyBytes := x509.MarshalPKCS1PrivateKey(key)
			out := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes})
			return base64.StdEncoding.EncodeToString(out), nil
		}
		// cli json output
		body := cliCertificateOutput{}
		if err := json.Unmarshal(fileData, &body); err == nil {
			return parseBase64EncodedPrivKey(body.PrivateKey)
		}
		// base64 encoded data
		res, err := parseBase64EncodedPrivKey(string(fileData))
		if err != nil {
			return "", fmt.Errorf("unknown private key format. Please try .pem/.der/.json")
		}
		return res, nil
	}
	return "", fmt.Errorf("undefined option")
}

func parseBase64EncodedCertificate(cert string) (string, error) {
	content, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		return "", fmt.Errorf("raw certificate must be base64 encoded")
	}
	if isCert, err := x509.ParseCertificate(content); err == nil {
		out := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: isCert.Raw})
		return base64.StdEncoding.EncodeToString(out), nil
	}
	if isPem, err := checkIsPem(content, "CERTIFICATE"); isPem {
		return cert, nil
	} else if err != nil {
		return "", err
	}
	return "", fmt.Errorf("certificate is malformed")
}

func parseBase64EncodedPrivKey(key string) (string, error) {
	content, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", fmt.Errorf("raw private key must be base64 encoded")
	}
	privKey, err := x509.ParsePKCS1PrivateKey(content)
	if err == nil {
		keyBytes := x509.MarshalPKCS1PrivateKey(privKey)
		out := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes})
		return base64.StdEncoding.EncodeToString(out), nil
	}
	if isPem, err := checkIsPem(content, "RSA PRIVATE KEY"); isPem {
		return key, nil
	} else if err != nil {
		return "", err
	}
	return "", fmt.Errorf("private key is malformed")
}

func checkIsPem(content []byte, blockType string) (bool, error) {
	block, _ := pem.Decode(content)
	if block != nil && block.Type == blockType {
		return true, nil
	} else if block != nil {
		return false, fmt.Errorf("expected .pem encoded '%s', got .pem encoded '%s'", blockType, block.Type)
	}
	return false, nil
}
