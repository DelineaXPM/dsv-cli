package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/DelineaXPM/dsv-cli/auth"
	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/internal/pki"
	"github.com/DelineaXPM/dsv-cli/internal/predictor"
	"github.com/DelineaXPM/dsv-cli/paths"
	"github.com/DelineaXPM/dsv-cli/store"
	"github.com/DelineaXPM/dsv-cli/utils"
	"github.com/DelineaXPM/dsv-cli/vaultcli"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

func GetCliConfigCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounCliConfig},
		SynopsisText: "Manage the CLI configuration",
		HelpText:     "Execute an action on the cli config for " + cst.ProductName,
		NoConfigRead: true,
		NoPreAuth:    true,
	})
}

func GetCliConfigInitCmd() (cli.Command, error) {
	homePath := "$HOME"
	if runtime.GOOS == "windows" {
		homePath = "%USERPROFILE%"
	}
	defaultConfigPath := filepath.Join(homePath, ".dsv.yml")

	return NewCommand(CommandArgs{
		Path:         []string{cst.NounCliConfig, cst.Init},
		SynopsisText: "Initialize or add a new profile to the CLI configuration",
		HelpText: `Command 'init' is an alias for 'cli-config init'.
For interactive mode provide no arguments.

Examples:
- Add a profile that uses username/password for authentication:
    • init \
        --profile prof \
        --tenant demo \
        --domain secretsvaultcloud.com \
        --store-type file \
        --store-path ~/.thy \
        --cache-strategy server \
        --auth-type password \
        --auth-username 'demouser' \
        --auth-password 'demopassword'

- Add a profile that uses client credentials for authentication:
    • init \
        --profile prof \
        --tenant demo \
        --domain secretsvaultcloud.com \
        --store-type file \
        --store-path ~/.thy \
        --cache-strategy server \
        --auth-type clientcred \
        --auth-client-id '11111111-2222-3333-4444-555555555555' \
        --auth-client-secret 'abcdefghijklmnopqrstuvwxyz012345'

- Add a profile that uses thycotic one for authentication:
    • init \
        --profile prof \
        --tenant demo \
        --domain secretsvaultcloud.com \
        --store-type file \
        --store-path ~/.thy \
        --cache-strategy server \
        --auth-type thy-one

- Add a profile that uses certificate for authentication:
    • init \
        --profile prof \
        --tenant demo \
        --domain secretsvaultcloud.com \
        --store-type file \
        --store-path ~/.thy \
        --cache-strategy server \
        --auth-type cert \
        --auth-certificate '@./path/to/certificate' \
        --auth-privateKey '@./path/to/private_key'
`,
		NoConfigRead: true,
		NoPreAuth:    true,
		FlagsPredictor: []*predictor.Params{
			// Configuration path and profile name.
			{Name: cst.Config, Shorthand: "c", Usage: fmt.Sprintf("Set config file path [default:%s]", defaultConfigPath)},
			{Name: cst.Profile, Usage: "Profile name to add to the config file"},

			// Tenant info.
			{Name: cst.Tenant, Usage: "Name of the tenant to connect to"},
			{Name: cst.DomainName, Usage: "Domain name, e.g. 'secretsvaultcloud.com'"},

			// Storing and Caching.
			{Name: cst.StoreType, Usage: "Store type (file|none|pass_linux|wincred)"},
			{Name: cst.StorePath, Usage: "Path to directory where to store. Only if store type is 'file'"},
			{Name: cst.CacheStrategy, Usage: "Cache strategy (server|server.cache|cache.server|cache.server.expired). Only if store type is not 'none'"},
			{Name: cst.CacheAge, Usage: "Cache age in minutes. Only if cache strategy is not 'server'"},

			// Authentication.
			{Name: cst.AuthType, Usage: "Authentication type (password|clientcred|thy-one|aws|azure|gcp|oidc|cert)"},
			{Name: cst.Username, Usage: "Username for 'password' authentication type"},
			{Name: cst.Password, Usage: "Password for 'password' authentication type"},
			{Name: cst.AuthClientID, Usage: "Client ID for 'clientcred' authentication type"},
			{Name: cst.AuthClientSecret, Usage: "Client Secret for 'clientcred' authentication type"},
			{Name: cst.AwsProfile, Usage: "AWS profile name for 'aws' authentication type"},
			{Name: cst.AuthProvider, Usage: "Authentication provider name for 'oidc' authentication type"},
			{Name: cst.AuthCert, Usage: "Certificate for 'cert' auth type. Prefix with '@' to denote filepath"},
			{Name: cst.AuthPrivateKey, Usage: "Private key for 'cert' auth type. Prefix with '@' to denote filepath"},

			{Name: cst.Dev, Hidden: true, Usage: "Specify dev domain upon initialization (ignored when '--domain' is used)"},
		},
		RunFunc: handleCliConfigInitCmd,
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
		RunFunc:       handleCliConfigUpdateCmd,
	})
}

func GetCliConfigClearCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounCliConfig, cst.Clear},
		SynopsisText: strings.Join([]string{cst.NounCliConfig, cst.Clear}, " "),
		HelpText:     "Clear the cli config for " + cst.ProductName,
		NoPreAuth:    true,
		RunFunc:      handleCliConfigClearCmd,
	})
}

func GetCliConfigReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounCliConfig, cst.Read},
		SynopsisText: strings.Join([]string{cst.NounCliConfig, cst.Read}, " "),
		HelpText:     "Read the cli config for " + cst.ProductName,
		NoPreAuth:    true,
		RunFunc:      handleCliConfigReadCmd,
	})
}

func GetCliConfigEditCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounCliConfig, cst.Edit},
		SynopsisText: strings.Join([]string{cst.NounCliConfig, cst.Edit}, " "),
		HelpText:     "Edit the cli config for " + cst.ProductName,
		NoPreAuth:    true,
		RunFunc:      handleCliConfigEditCmd,
	})
}

func GetCliConfigUseProfileCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounCliConfig, cst.UseProfile},
		SynopsisText: "Set a profile name which will be used by default.",
		HelpText: `Usage:
   • cli-config use-profile admin
   • cli-config use-profile role1

For interactive mode provide no arguments:
   • cli-config use-profile`,
		NoPreAuth: true,
		RunFunc:   handleCliConfigUseProfileCmd,
	})
}

func handleCliConfigUseProfileCmd(vcli vaultcli.CLI, args []string) int {
	cfgPath := viper.GetString(cst.Config)
	cf, err := vaultcli.ReadConfigFile(cfgPath)
	if err != nil {
		vcli.Out().FailF("Error: failed to read config: %v.", err)
		return 1
	}
	existingProfiles := cf.ListProfiles()
	if len(existingProfiles) == 1 {
		vcli.Out().FailS("Default profile cannot be changed since only one profile defined in the configuration file.")
		return 0
	}

	log.Printf("Current profile used by default: %q.", cf.DefaultProfile)

	var profile string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		profile = args[0]
		if _, ok := cf.GetProfile(profile); !ok {
			vcli.Out().FailF("Error: profile %q not found in config.", profile)
			return 1
		}
	} else {
		profiles := make([]string, 0, len(existingProfiles))

		for _, prof := range existingProfiles {
			repr := fmt.Sprintf("%s (tenant: %s; auth type: %s)",
				prof.Name, prof.Get(cst.Tenant), prof.Get(cst.NounAuth, cst.Type))

			if prof.Name == cf.DefaultProfile {
				profiles = append(profiles, repr+" <-- [current default]")
			} else {
				profiles = append(profiles, repr)
			}
		}

		var profileID int
		profilePrompt := &survey.Select{
			Message:  "Please select default profile:",
			Options:  profiles,
			PageSize: 10,
		}
		err = survey.AskOne(profilePrompt, &profileID)
		if err != nil {
			vcli.Out().Fail(err)
			return 1
		}
		profile = existingProfiles[profileID].Name
	}

	if profile == cf.DefaultProfile {
		return 0
	}

	cf.DefaultProfile = profile

	err = cf.Save()
	if err != nil {
		vcli.Out().FailF("Error: cannot save configuration: %v.", err)
		return 1
	}
	return 0
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
	cf.SetProfile(prf)
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
	if err != nil {
		err.Grow("Invalid configuration:")
	}

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
			cfgPath = vaultcli.LookupConfigPath(cfgPath)
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
		ui.Warn(fmt.Sprintf("Found an existing cli-config located at '%s'.", cf.GetPath()))
	}

	profile := viper.GetString(cst.Profile)
	if profile != "" {
		lowered := strings.ToLower(profile)
		if lowered != profile {
			profile = lowered
			ui.Warn(fmt.Sprintf("Profile name can only use lowercase letters. The name '%s' will be used instead.", profile))
		}
		if err := vaultcli.ValidateProfile(profile); err != nil {
			ui.Info(err.Error())
			return 1
		}
		if cfgExists {
			// The user specified --profile [name], so the intent is to add a profile to the config.
			if _, ok := cf.GetProfile(profile); ok {
				msg := fmt.Sprintf("Profile %q already exists in the config.", profile)
				ui.Info(msg)
				return 1
			}
		}
	}
	if !cfgExists && profile == "" {
		// If configuration file does not exist and profile name was not provided
		// then use default profile name to have better user experience with less
		// questions to answer.
		profile = cst.DefaultProfile
	}

	if cfgExists && profile == "" {
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
		if actionID == 0 { // "Do nothing".
			ui.Info("Exiting.")
			return 0
		}

		var profilePrompt *survey.Input
		switch actionID {
		case 1: // "Overwrite the config".
			cf, _ = vaultcli.NewConfigFile(cf.GetPath())
			profilePrompt = &survey.Input{Message: "Please enter profile name:", Default: cst.DefaultProfile}

		case 2: // "Add a new profile to the config".
			profilePrompt = &survey.Input{Message: "Please enter profile name:"}
		}

		validationOpt := survey.WithValidator(vaultcli.SurveyRequiredProfileName(cf.ListProfilesNames()))
		survErr = survey.AskOne(profilePrompt, &profile, validationOpt)
		if survErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(survErr))
			return 1
		}
		profile = strings.TrimSpace(profile)
	}
	prf := vaultcli.NewProfile(profile)
	viper.Set(cst.Profile, profile)

	// Tenant name.
	tenant := viper.GetString(cst.Tenant)
	if tenant == "" {
		tenantPrompt := &survey.Input{Message: "Please enter tenant name:"}
		survErr := survey.AskOne(tenantPrompt, &tenant, survey.WithValidator(vaultcli.SurveyRequired))
		if survErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(survErr))
			return utils.GetExecStatus(survErr)
		}
		tenant = strings.TrimSpace(tenant)
		viper.Set(cst.Tenant, tenant)
	}

	prf.Set(tenant, cst.Tenant)

	// Domain.
	domain := viper.GetString(cst.DomainName)

	var isDevDomain bool
	if domain == "" {
		devDomain := viper.GetString(cst.Dev)

		if devDomain != "" {
			isDevDomain = true
			domain = devDomain
		} else {
			domain, err = promptDomain()
			if err != nil {
				vcli.Out().WriteResponse(nil, errors.New(err))
				return 1
			}
		}
		viper.Set(cst.DomainKey, domain)
	}
	prf.Set(domain, cst.DomainKey)

	// Check if tenant has been setup yet.
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

	// Store configuration.
	storeType := viper.GetString(cst.StoreType)
	if storeType == "" {
		storeType, err = promptStoreType()
		if err != nil {
			vcli.Out().WriteResponse(nil, errors.New(err))
			return 1
		}
	}
	if err := store.ValidateCredentialStore(storeType); err != nil {
		ui.Error(fmt.Sprintf("Failed to get store: %v.", err))
		return 1
	}
	isSecureStore := storeType == store.PassLinux || storeType == store.WinCred

	prf.Set(storeType, cst.Store, cst.Type)
	viper.Set(cst.StoreType, storeType)

	if storeType == store.File {
		fileStorePath := viper.GetString(cst.StorePath)

		if fileStorePath == "" {
			def := filepath.Join(utils.NewEnvProvider().GetHomeDir(), ".thy")
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
		}

		fileStorePath, err = filepath.Abs(fileStorePath)
		if err != nil {
			vcli.Out().FailF("Failed to resolve absolute path: %v", err)
			return utils.GetExecStatus(err)
		}

		prf.Set(fileStorePath, cst.Store, cst.Path)
		viper.Set(cst.StorePath, fileStorePath)
	}

	// Caching strategy and cache age.
	if storeType != store.None {
		cacheStrategy := viper.GetString(cst.CacheStrategy)
		if cacheStrategy == "" {
			cacheStrategy, err = promptCacheStrategy()
			if err != nil {
				vcli.Out().WriteResponse(nil, errors.New(err))
				return 1
			}
		}

		prf.Set(cacheStrategy, cst.Cache, cst.Strategy)

		if cacheStrategy != cst.CacheStrategyNever {
			cacheAge := viper.GetString(cst.CacheAge)
			if cacheAge == "" {
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
			} else {
				if age, err := strconv.Atoi(cacheAge); err != nil || age <= 0 {
					vcli.Out().FailS("Unable to parse cache age. Must be strictly positive int")
					return 1
				}
			}
			prf.Set(cacheAge, cst.Cache, cst.Age)
		}
	}

	// Authentication type and authentication data.
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

	var user, password, passwordKey, encryptionKeyFileName string

	authProvider := viper.GetString(cst.AuthProvider)

	switch {
	case (storeType != store.None && auth.AuthType(authType) == auth.Password):
		user = viper.GetString(cst.Username)
		password = viper.GetString(cst.Password)

		if user == "" || password == "" {
			if setupRequired {
				user, password, err = promptInitialUsernamePassword(tenant)
			} else {
				user, password, err = promptUsernamePassword(tenant)
			}
			if err != nil {
				vcli.Out().WriteResponse(nil, errors.New(err))
				return 1
			}
			viper.Set(cst.Username, user)
			viper.Set(cst.Password, password)
		}

		prf.Set(user, cst.NounAuth, cst.DataUsername)

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
		clientID := viper.GetString(cst.AuthClientID)
		clientSecret := viper.GetString(cst.AuthClientSecret)

		if clientID == "" || clientSecret == "" {
			clientID, clientSecret, err = promptClientCredentials()
			if err != nil {
				vcli.Out().WriteResponse(nil, errors.New(err))
				return 1
			}
			viper.Set(cst.AuthClientID, clientID)
		}

		prf.Set(clientID, cst.NounAuth, cst.NounClient, cst.ID)

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
		awsProfile := viper.GetString(cst.AwsProfile)
		if awsProfile == "" {
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
			viper.Set(cst.AwsProfile, awsProfile)
		}
		prf.Set(awsProfile, cst.NounAuth, cst.NounAwsProfile)

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
			if isDevDomain {
				if authProvider == "" {
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
			} else {
				authProvider = cst.DefaultThyOneName
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
		clientCert := viper.GetString(cst.AuthCert)
		clientPrivKey := viper.GetString(cst.AuthPrivateKey)

		clientCert, err = parseClientCertificate(clientCert)
		if err != nil {
			vcli.Out().WriteResponse(nil, errors.New(err))
			return 1
		}
		clientPrivKey, err = parseClientPrivKey(clientPrivKey)
		if err != nil {
			vcli.Out().WriteResponse(nil, errors.New(err))
			return 1
		}

		viper.Set(cst.AuthCert, clientCert)
		viper.Set(cst.AuthPrivateKey, clientPrivKey)

		prf.Set(clientCert, cst.NounAuth, cst.NounCert)
		prf.Set(clientPrivKey, cst.NounAuth, cst.NounPrivateKey)
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

	cf.SetProfile(prf)
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

	apiError := authenticator.WipeCachedTokens()
	if apiError != nil {
		return apiError
	}

	_, apiError = authenticator.GetToken()
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

func parseClientCertificate(cert string) (string, error) {
	if cert == "" {
		var err error
		cert, err = promptCertificate()
		if err != nil {
			return "", err
		}
	}

	if strings.HasPrefix(cert, cst.CmdFilePrefix) {
		filePath := strings.TrimPrefix(cert, cst.CmdFilePrefix)
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}

		body := cliCertificateOutput{}
		if err := json.Unmarshal(fileData, &body); err == nil {
			cert = body.Certificate
		} else {
			cert = string(fileData)
		}
	}

	return pki.CertToBase64EncodedPEM(cert)
}

func parseClientPrivKey(pk string) (string, error) {
	if pk == "" {
		var err error
		pk, err = promptPrivateKey()
		if err != nil {
			return "", err
		}
	}

	if strings.HasPrefix(pk, cst.CmdFilePrefix) {
		filePath := strings.TrimPrefix(pk, cst.CmdFilePrefix)
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}

		body := cliCertificateOutput{}
		if err := json.Unmarshal(fileData, &body); err == nil {
			pk = body.PrivateKey
		} else {
			pk = string(fileData)
		}
	}

	return pki.PrivateKeyToBase64EncodedPEM(pk)
}

type cliCertificateOutput struct {
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"privateKey"`
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

	switch actionID {
	case 0: // "Raw certificate".
		var rawCert string
		rawCertPrompt := &survey.Input{Message: "Raw certificate:"}
		survErr := survey.AskOne(rawCertPrompt, &rawCert, survey.WithValidator(vaultcli.SurveyRequired))
		if survErr != nil {
			return "", survErr
		}
		return rawCert, nil

	case 1: // "Certificate file path".
		var filePath string
		filePrompt := &survey.Input{Message: "Certificate file path:"}
		survErr := survey.AskOne(filePrompt, &filePath, survey.WithValidator(vaultcli.SurveyRequired))
		if survErr != nil {
			return "", survErr
		}
		return cst.CmdFilePrefix + filePath, nil

	default:
		return "", fmt.Errorf("Unhandled case for certificate prompting flow (id %d)", actionID)
	}
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

	switch actionID {
	case 0: // "Raw private key".
		var rawPrivKey string
		rawPrivKeyPrompt := &survey.Password{Message: "Private key:"}
		survErr := survey.AskOne(rawPrivKeyPrompt, &rawPrivKey, survey.WithValidator(vaultcli.SurveyRequired))
		if survErr != nil {
			return "", survErr
		}
		return rawPrivKey, nil

	case 1: // "Private key file path".
		var filePath string
		filePrompt := &survey.Input{Message: "Private key file path:"}
		survErr := survey.AskOne(filePrompt, &filePath, survey.WithValidator(vaultcli.SurveyRequired))
		if survErr != nil {
			return "", survErr
		}

		return cst.CmdFilePrefix + filePath, nil

	default:
		return "", fmt.Errorf("Unhandled case for private key prompting flow (id %d)", actionID)
	}
}
