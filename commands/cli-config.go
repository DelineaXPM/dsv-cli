package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"thy/auth"
	config "thy/cli-config"
	cst "thy/constants"
	"thy/errors"
	"thy/internal/predictor"
	"thy/internal/prompt"
	"thy/paths"
	"thy/store"
	"thy/vaultcli"

	"thy/utils"

	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const ProfileNameContainsRestrictedCharacters = "Profile name contains restricted characters. Leading, trailing and middle whitespace are not allowed."

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
	if key == "" && len(args) > 0 {
		key = args[0]
	}
	if val == "" && len(args) > 1 {
		// Make value the same as the key name, as in "user:user", for example.
		val = args[1]
	}

	var valInterface interface{}
	if val != "" && val != "0" {
		if i, err := strconv.Atoi(val); err == nil {
			valInterface = i
		} else {
			valInterface = val
		}
	}

	var storeType string

	storeKeySecure := key == cst.Password || key == cst.AuthClientSecret

	if storeKeySecure {
		storeType = viper.GetString(cst.StoreType)
		if storeType == store.Unset {
			storeType = store.File
		}
		if storeType == store.PassLinux || storeType == store.WinCred {
			storeKeySecure = true
		} else {
			storeKeySecure = false
		}
	}

	cfgPath := config.GetFlagBeforeParse(cst.Config, args)
	if cfgPath == "" {
		cfgPath = config.GetCliConfigPath()
	}

	profile := viper.GetString(cst.Profile)
	if profile == "" {
		profile = cst.DefaultProfile
	}
	key = profile + "." + key

	if !storeKeySecure {
		if cfgPath == "" {
			vcli.Out().FailS("CLI config path could not be resolved. Exiting.")
			return 1
		}

		cfgContent, err := os.ReadFile(cfgPath)
		if err != nil {
			vcli.Out().FailS("Failed to load CLI config from file. Exiting. Error: " + err.Error())
			return 1
		}

		var cfgMap interfaceMap
		err = yaml.Unmarshal(cfgContent, &cfgMap)
		if err != nil || cfgMap == nil {
			vcli.Out().FailS("Failed to unmarshal CLI config. Exiting. Error: " + err.Error())
			return 1
		}

		cfg := jsonish{}
		err = MapToJsonish(cfgMap, &cfg, nil)
		if err != nil {
			vcli.Out().FailS("Failed to convert CLI config to expected format. Exiting. Error: " + err.Error())
			return 1
		}

		keys := strings.Split(key, ".")

		secure := "auth.securePassword"
		if strings.HasSuffix(key, secure) {
			vcli.Out().FailF("Cannot manually change the value of %q.", secure)
			return 1
		}
		if strings.HasSuffix(key, "auth.password") {
			key = fmt.Sprintf("%s.%s", profile, secure)
			keys = strings.Split(key, ".")
			cipherText, err := auth.EncipherPassword(val)
			if err != nil {
				vcli.Out().FailF("Error encrypting password: %v.", err)
				return 1
			}
			valInterface = cipherText
			// Set in-memory value of plaintext password, so that the CLI can use it to try to authenticate before writing the config.
			viper.Set(cst.Password, val)
			if authError := tryAuthenticate(); authError != nil {
				vcli.Out().FailS("Failed to authenticate, restoring previous config.")
				vcli.Out().FailS("Please check your credentials and try again.")
				return 1
			}
		}
		if valInterface == nil {
			RemoveNode(&cfg, keys...)
		} else {
			AddNode(&cfg, jsonish{keys[len(keys)-1]: valInterface}, keys[:len(keys)-1]...)
		}
		if err := WriteCliConfig(cfgPath, cfg, false); err != nil {
			vcli.Out().FailS(err.Error())
			return 1
		}
	} else if err := store.StoreSecureSetting(key, val, storeType); err != nil {
		vcli.Out().FailF("Error updating setting in store of type '%s'", string(storeType))
		return 1
	}
	return 0
}

func handleCliConfigReadCmd(vcli vaultcli.CLI, args []string) int {
	cfgPath := config.GetFlagBeforeParse(cst.Config, args)
	if cfgPath == "" {
		cfgPath = config.GetCliConfigPath()
	}

	var didError int
	dataOut := []byte(fmt.Sprintf("CLI config (%s):\n", cfgPath))
	if cfgPath == "" {
		vcli.Out().FailS("CLI config path could not be resolved. Exiting.")
		didError = 1
	} else if b, err := os.ReadFile(cfgPath); err != nil {
		vcli.Out().FailS("Failed to read CLI config: " + err.Error())
		didError = 1
	} else {
		dataOut = append(dataOut, b...)
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
			dataOut = append(dataOut, strings.Replace(strings.Join(keys, "\n  "), "-", ".", -1)...)
		}
	}

	vcli.Out().WriteResponse(dataOut, nil)
	return didError
}

func handleCliConfigEditCmd(vcli vaultcli.CLI, args []string) int {
	var dataOut []byte
	cfgPath := config.GetFlagBeforeParse(cst.Config, args)
	if cfgPath == "" {
		cfgPath = config.GetCliConfigPath()
	}
	if cfgPath == "" {
		vcli.Out().FailS("CLI config path could not be resolved. Exiting.")
		return 1
	} else if b, err := os.ReadFile(cfgPath); err != nil {
		vcli.Out().FailS("Failed to read CLI config: " + err.Error())
		return 1
	} else {
		dataOut = append(dataOut, b...)
	}

	saveFunc := func(data []byte) (resp []byte, err *errors.ApiError) {
		writeErr := ioutil.WriteFile(cfgPath, data, 0600)
		return nil, errors.New(writeErr)
	}

	_, err := vcli.Edit(dataOut, saveFunc)

	vcli.Out().WriteResponse(nil, err)
	return utils.GetExecStatus(err)
}

func handleCliConfigClearCmd(vcli vaultcli.CLI, args []string) int {
	var ui cli.Ui = &cli.BasicUi{
		Writer:      os.Stdout,
		Reader:      os.Stdin,
		ErrorWriter: os.Stderr,
	}

	if yes, err := prompt.YesNo(ui, "Are you sure you want to delete CLI configuration?", false); err != nil {
		ui.Error(err.Error())
		return 1
	} else if !yes {
		ui.Info("exiting")
		return 0
	}

	cfgPath := config.GetFlagBeforeParse(cst.Config, args)
	if cfgPath == "" {
		cfgPath = config.GetCliConfigPath()
	}
	if cfgPath == "" {
		ui.Warn("CLI config path could not be resolved. Exiting.")
		return 1
	} else if err := os.Remove(cfgPath); err != nil {
		ui.Error("Error deleting CLI config: " + err.Error())
	}
	st := viper.GetString(cst.StoreType)
	if s, err := store.GetStore(st); err == nil {
		err = s.Wipe("")
	} else {
		ui.Error("Failed to clear store: " + err.Error())
	}

	return 0
}

func IsValidProfile(profile string) bool {
	return !strings.Contains(profile, " ")
}

func handleCliConfigInitCmd(vcli vaultcli.CLI, args []string) int {
	ui := &PasswordUi{
		cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
	}
	cfg := jsonish{}

	var isSecureStore bool
	var addProfile bool
	cfgPath := config.GetFlagBeforeParse(cst.Config, args)
	if cfgPath == "" {
		cfgPath = config.GetCliConfigPath()
	}

	profile := strings.ToLower(viper.GetString(cst.Profile))
	if profile == "" {
		profile = cst.DefaultProfile
	} else if !IsValidProfile(profile) {
		ui.Info(ProfileNameContainsRestrictedCharacters)
		return 1
	} else {
		// If not default profile, that means the user specified --profile [name], so the intent is to add a profile to the config.
		addProfile = true
	}

	viper.Set(cst.Profile, profile)

	if cfgPath != "" {
		if cfgInfo, err := os.Stat(cfgPath); err == nil && cfgInfo != nil {
			cfgExists := false
			switch m := cfgInfo.Mode(); {
			case m.IsDir():
				cfgPath = filepath.Join(cfgPath, ".thy.yml")
				if newCfgInfo, err := os.Stat(cfgPath); err == nil && newCfgInfo != nil {
					ui.Warn(fmt.Sprintf("Found an existing cli-config located at '%s'", cfgPath))
					cfgExists = true
				}
			case m.IsRegular():
				ui.Warn(fmt.Sprintf("Found an existing cli-config located at '%s'", cfgPath))
				cfgExists = true
			}
			if cfgExists {
				viper.SetConfigFile(cfgPath)
				if profile == cst.DefaultProfile {
					if resp, err := ui.Ask("Select an option:\n\t[o] overwrite the config\n\t[a] add a new profile to the config\n\t[n] do nothing\n(default:n)"); err != nil {
						ui.Error(err.Error())
						return 1
					} else if action := getMainAction(resp); action == "n" {
						ui.Info("Exiting.")
						return 0
					} else if action == "a" {
						addProfile = true
					}
				} else {
					err := viper.ReadInConfig(profile)
					// Reading the config for `thy init` and looking for the specified profile must fail, otherwise,
					// the profile already exists and so cannot be added.
					if err == nil {
						msg := fmt.Sprintf("Profile %q already exists in the config.", profile)
						ui.Info(msg)
						return 1
					}
				}

				// If default profile, that means the user did not specify --profile [name], so ask for profile name.
				if addProfile && profile == cst.DefaultProfile {
					var p string
					if p, err = prompt.Ask(ui, "Please enter profile name:"); err != nil {
						return 1
					}

					p = strings.ToLower(p)
					if err := viper.ReadInConfig(p); err == nil {
						msg := fmt.Sprintf("Profile %q already exists in the config.", p)
						ui.Info(msg)
						return 1
					}

					profile = p
					if !IsValidProfile(profile) {
						ui.Info(ProfileNameContainsRestrictedCharacters)
						return 1
					}

					viper.Set(cst.Profile, profile)
				}
			}
		} else if profile != cst.DefaultProfile {
			// If config does not exist, and the user specified a non-default --profile name, then quit and ask to properly init.
			ui.Info("Initial configuration is needed in order to add a custom profile.")
			ui.Info(fmt.Sprintf("Create CLI config file manually or execute command '%s init' to initiate CLI configuration.", cst.CmdRoot))
			return 1
		}
	}

	// tenant
	if tenant, err := prompt.Ask(ui, "Please enter tenant name:"); err != nil {
		return 1
	} else {
		cfg[profile] = jsonish{
			cst.Tenant: tenant,
		}
		viper.Set(cst.Tenant, tenant)
	}

	// domain
	var domain string
	var err error
	var isDevDomain bool

	if devDomain := viper.GetString(cst.Dev); devDomain != "" {
		isDevDomain = true
		domain = devDomain
	} else {
		if domain, err = prompt.Choose(ui, "Please choose domain:",
			prompt.Option{cst.Domain, cst.Domain},
			prompt.Option{cst.DomainEU, cst.DomainEU},
			prompt.Option{cst.DomainAU, cst.DomainAU},
			prompt.Option{cst.DomainCA, cst.DomainCA}); err != nil {
			return 1
		}
		if domain == "" {
			domain = cst.Domain
		}
	}
	AddNode(&cfg, jsonish{cst.DomainKey: domain}, profile)

	// check if tenant has been setup yet
	viper.Set(cst.DomainKey, domain)
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
	var storeType string
	if st, err := prompt.Choose(ui, "Please enter store type:",
		prompt.Option{store.File, "File store"},
		prompt.Option{store.None, "None (no caching)"},
		prompt.Option{store.PassLinux, "Pass (Linux only)"},
		prompt.Option{store.WinCred, "Windows Credential Manager (Windows only)"}); err != nil {
		return 1
	} else {
		storeType = st
		if err := store.ValidateCredentialStore(storeType); err != nil {
			ui.Error(fmt.Sprintf("Failed to get store: %v.", err))
			return 1
		}
		if storeType == store.PassLinux || storeType == store.WinCred {
			isSecureStore = true
		}
		AddNode(&cfg, jsonish{cst.Type: storeType}, profile, cst.Store)
		if storeType == store.File {
			def := filepath.Join(utils.NewEnvProvider().GetHomeDir(), ".thy")
			sp, _ := prompt.AskDefault(ui, "Please enter directory for file store", def)
			if sp != def {
				AddNode(&cfg, jsonish{cst.Path: sp}, profile, cst.Store)
				viper.Set(cst.StorePath, sp)
			}
		}
		viper.Set(cst.StoreType, storeType)
	}

	//cache
	if storeType != store.None {
		if strategy, err := prompt.Choose(ui, "Please enter cache strategy for secrets:",
			prompt.Option{cst.CacheStrategyNever, "Never"},
			prompt.Option{cst.CacheStrategyServerThenCache, "Server then cache"},
			prompt.Option{cst.CacheStrategyCacheThenServer, "Cache then server"},
			prompt.Option{cst.CacheStrategyCacheThenServerThenExpired, "Cache then server, but allow expired cache if server unreachable"}); err != nil {
			return 1
		} else {
			AddNode(&cfg, jsonish{cst.Strategy: strategy}, profile, cst.Cache)
			if strategy != cst.CacheStrategyNever {
				if ageString, err := prompt.Ask(ui, "Please enter cache age (minutes until expiration):"); err != nil {
					return 1
				} else if age, err := strconv.Atoi(ageString); err != nil || age <= 0 {
					vcli.Out().FailS("Error. Unable to parse age. Must be strictly positive int: " + ageString)
				} else {
					AddNode(&cfg, jsonish{cst.Age: age}, profile, cst.Cache)
				}
			}
		}
	}

	//auth
	var authType, user, password, passwordKey, authProvider, encryptionKeyFileName string
	// allow overriding option with flag
	authType = viper.GetString(cst.AuthType)
	authProvider = viper.GetString(cst.AuthProvider)
	if authType == "" {
		if at, err := prompt.Choose(ui, "Please enter auth type:",
			prompt.Option{string(auth.Password), "Password (local user)"},
			prompt.Option{string(auth.ClientCredential), "Client Credential"},
			prompt.Option{string(auth.FederatedThyOne), "Thycotic One (federated)"},
			prompt.Option{string(auth.FederatedAws), "AWS IAM (federated)"},
			prompt.Option{string(auth.FederatedAzure), "Azure (federated)"},
			prompt.Option{string(auth.FederatedGcp), "GCP (federated)"},
			prompt.Option{string(auth.Oidc), "OIDC (federated)"},
			prompt.Option{string(auth.Certificate), "x509 Certificate"}); err != nil {
			return 1
		} else {
			authType = at
			viper.Set(cst.AuthType, authType)
		}
	}
	AddNode(&cfg, jsonish{cst.Type: authType}, profile, cst.NounAuth)
	if storeType != store.None {
		if auth.AuthType(authType) == auth.Password {
			var passwordMessage, userMessage string
			var askFunc = prompt.AskSecure
			tenant := viper.GetString(cst.Tenant)
			if setupRequired {
				userMessage = fmt.Sprintf("Please choose a username for initial local admin for tenant %q:", tenant)
				passwordMessage = "Please choose password:"
				askFunc = prompt.AskSecureConfirm
			} else {
				userMessage = fmt.Sprintf("Please enter username for tenant %q:", tenant)
				passwordMessage = "Please enter password:"
			}

			if user, err = prompt.Ask(ui, userMessage); err != nil {
				return 1
			} else {
				AddNode(&cfg, jsonish{cst.DataUsername: user}, profile, cst.NounAuth)
				viper.Set(cst.Username, user)
			}

			encryptionKeyFileName = auth.GetEncryptionKeyFilename(viper.GetString(cst.Tenant), user)
			if password, err = askFunc(ui, passwordMessage); err != nil {
				return 1
			} else {
				if isSecureStore {
					viper.Set(cst.Password, password)
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
					AddNode(&cfg, jsonish{cst.DataSecurePassword: encrypted}, profile, cst.NounAuth)
					viper.Set(cst.Password, password)
				}
			}
		} else if auth.AuthType(authType) == auth.ClientCredential {
			if id, err := prompt.Ask(ui, "Please enter client id for client auth:"); err != nil {
				return 1
			} else {
				AddNode(&cfg, jsonish{cst.ID: id}, profile, cst.NounAuth, cst.NounClient)
				viper.Set(cst.AuthClientID, id)
			}
			if secret, err := prompt.AskSecure(ui, "Please enter client secret for client auth:"); err != nil {
				return 1
			} else {
				if isSecureStore {
					if err := store.StoreSecureSetting(strings.Join([]string{profile, cst.NounAuth, cst.NounClient, cst.NounSecret}, "."), secret, storeType); err != nil {
						ui.Error(err.Error())
						return 1

					}
				} else {
					AddNode(&cfg, jsonish{cst.NounSecret: secret}, profile, cst.NounAuth, cst.NounClient)
					viper.Set(cst.AuthClientSecret, secret)
				}
			}
		} else if auth.AuthType(authType) == auth.FederatedAws {
			var awsProfile string
			if awsProfile, err = prompt.AskDefault(ui, "Please enter aws profile for federated aws auth", "default"); err != nil {
				return 1
			}
			AddNode(&cfg, jsonish{cst.NounAwsProfile: awsProfile}, profile, cst.NounAuth)
			viper.Set(cst.AwsProfile, awsProfile)
		} else if auth.AuthType(authType) == auth.Oidc || auth.AuthType(authType) == auth.FederatedThyOne {
			if auth.AuthType(authType) == auth.Oidc {
				if authProvider == "" {
					if authProvider, err = prompt.AskDefault(ui, "Please enter auth provider name", cst.DefaultThyOneName); err != nil {
						return 1
					}
				}
			} else {
				authProvider = cst.DefaultThyOneName

				if isDevDomain {
					if authProvider, err = prompt.AskDefault(ui, "Thycotic One authentication provider name", cst.DefaultThyOneName); err != nil {
						return 1

					}
				}
			}

			viper.Set(cst.AuthProvider, authProvider)
			AddNode(&cfg, jsonish{cst.DataProvider: authProvider}, profile, cst.NounAuth)

			var callback string
			if callback = viper.GetString(cst.Callback); callback == "" {
				callback = cst.DefaultCallback
			}

			viper.Set(cst.Callback, callback)
			AddNode(&cfg, jsonish{cst.DataCallback: callback}, profile, cst.NounAuth)
		} else if auth.AuthType(authType) == auth.Certificate {

			clientCert, err := prompt.Ask(ui, "Certificate:")
			if err != nil {
				ui.Error(err.Error())
				return 1
			}

			clientPrivKey, err := prompt.Ask(ui, "Private key:")
			if err != nil {
				ui.Error(err.Error())
				return 1
			}

			viper.Set(cst.AuthCert, clientCert)
			viper.Set(cst.AuthPrivateKey, clientPrivKey)

			AddNode(&cfg, jsonish{cst.NounCert: clientCert}, profile, cst.NounAuth)
			AddNode(&cfg, jsonish{cst.NounPrivateKey: clientPrivKey}, profile, cst.NounAuth)
		}
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
	}

	if err := WriteCliConfig(cfgPath, cfg, addProfile); err != nil {
		ui.Error(err.Error())
		return 1
	}
	if storeType == store.None {
		ui.Output("Config created but no credentials saved, specify them as environment variables or via command line flags.")
	}
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

// WriteCliConfig writes the actual config file given the path, the config data structure and whether an existing config
// must be overwritten or extended with another profile.
func WriteCliConfig(cfgPath string, cfg jsonish, addingProfile bool) error {
	if o, err := yaml.Marshal(cfg); err != nil {
		return err
	} else if !addingProfile {
		// If not adding a profile, then overwrite the config file.
		if err := ioutil.WriteFile(cfgPath, o, 0600); err != nil {
			return err
		}
	} else {
		f, err := os.OpenFile(cfgPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := f.Write(o); err != nil {
			return err
		}
	}
	return nil
}

type jsonish map[string]interface{}

type interfaceMap map[interface{}]interface{}

func MapToJsonish(m interfaceMap, out *jsonish, startKey []string) error {
	for k, v := range m {
		kString, ok := k.(string)
		if !ok {
			return fmt.Errorf("cannot map key %v to string", k)
		}
		switch s := v.(type) {
		case int:
			if s == 0 {
				continue
			}
			AddNode(out, jsonish{kString: s}, startKey...)
		case string:
			if s == "" {
				continue
			}
			AddNode(out, jsonish{kString: s}, startKey...)
		case interfaceMap:
			if startKey == nil {
				startKey = []string{}
			}
			MapToJsonish(s, out, append(startKey, kString))
		default:
			return fmt.Errorf("unsupported value type in cli configuration: %v", reflect.TypeOf(s))
		}
	}
	return nil
}

func RemoveNode(n1 *jsonish, keyPath ...string) {
	n := n1
	for i, k := range keyPath {
		if i < len(keyPath)-1 {
			currNode := *n
			untyped, ok := currNode[k]
			if !ok || untyped == nil {
				currNode[k] = jsonish{}
			}
			nextNode := currNode[k].(jsonish)
			n = &nextNode
		}
	}
	delete(*n, keyPath[len(keyPath)-1])
}

func AddNode(n1 *jsonish, n2 jsonish, keyPath ...string) {
	n := n1
	for _, k := range keyPath {
		currNode := *n
		untyped, ok := currNode[k]
		if !ok || untyped == nil {
			currNode[k] = jsonish{}
		}
		nextNode := currNode[k].(jsonish)
		n = &nextNode
	}

	for k, v := range n2 {
		(*n)[k] = v
	}
}

// getMainAction parses the main intent of the CLI initialization command (add profile to config, overwrite config, or do nothing).
func getMainAction(response string) string {
	resUpper := strings.ToUpper(response)
	if resUpper == "A" || resUpper == "ADD" {
		return "a"
	} else if resUpper == "O" || resUpper == "OVERWRITE" {
		return "o"
	} else if resUpper == "N" || resUpper == "NO" || resUpper == "NOTHING" {
		return "n"
	} else {
		return "n"
	}
}

type heartbeatResponse struct {
	StatusCode int
	Message    string
}

type initializeRequest struct {
	UserName string
	Password string
}
