package cliconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cst "thy/constants"
	terrors "thy/errors"
	"thy/format"
	"thy/store"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// InitCliConfig reads in CLI config file and environment variables.
func InitCliConfig(cfgFile string, profile string, args []string) *terrors.ApiError {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			return terrors.New(err).Grow("failed to initialize cli config")
		}

		// Search config in home directory with name ".thy.yml".
		viper.AddConfigPath(home)
		viper.SetConfigType(cst.CliConfigType)
		viper.SetConfigName(cst.CliConfigName)
	}

	viper.SetEnvPrefix(cst.EnvVarPrefix)
	envReplacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(envReplacer)
	viper.AutomaticEnv()

	if profile == "" {
		profile = viper.GetString(cst.Profile)
		if profile == "" {
			profile = cst.DefaultProfile
		}
	}

	if err := viper.ReadInConfig(profile); err != nil {
		if v := GetFlagBeforeParse(cst.Tenant, args); v != "" {
			return nil
		}

		if eString := err.Error(); strings.Contains(eString, "invalid subkey") {
			return terrors.NewS(fmt.Sprintf("Invalid or non-existent profile in CLI config: %s.", profile))
		}

		// Do not return the error to allow users to view help text for commands.
		out := format.NewDefaultOutClient()
		out.FailS(fmt.Sprintf("Create CLI config file manually or execute command '%s init' to initiate CLI configuration - cannot find config.", cst.CmdRoot))
	}

	return nil
}

func GetCliConfigPath() string {
	cfgPath := viper.GetString(cst.Config)
	if cfgPath == "" {
		home, _ := homedir.Dir()
		if home != "" {
			cfgPath = filepath.Join(home, fmt.Sprintf("%s.%s", cst.CliConfigName, cst.CliConfigType))
		}
	}
	return cfgPath
}

func GetSecureSettingForProfile(key string, profile string) (string, *terrors.ApiError) {
	if key == "" {
		return "", terrors.NewS("key cannot be empty")
	}
	if val := viper.GetString(key); val != "" {
		return val, nil
	}

	if profile == "" {
		return "", terrors.NewS("profile cannot be empty")
	}
	keyProfile := profile + "-" + key

	storeType := viper.GetString(cst.StoreType)
	s, err := store.GetStore(storeType)
	if err != nil {
		return "", terrors.New(err).Grow("failed to fetch store")
	}
	keyFull := cst.CliConfigRoot + "-" + strings.Replace(keyProfile, ".", "-", -1)
	var res string
	err = s.Get(keyFull, &res)
	return res, err
}

func GetSecureSetting(key string) (string, *terrors.ApiError) {
	if key == "" {
		return "", terrors.NewS("key cannot be empty")
	}
	profile := viper.GetString(cst.Profile)
	if profile == "" {
		profile = cst.DefaultProfile
	}
	return GetSecureSettingForProfile(key, profile)
}

func GetFlagBeforeParse(flag string, args []string) string {
	flagNames := []string{`--` + flag} // Long form of the flag. For example: `--config`
	shortFlag := cst.GetShortFlag(flag)
	if shortFlag != "" {
		flagNames = append(flagNames, `-`+shortFlag) // Add short form of the flag. For example: `-c`
	}

	for i, arg := range args {
		for _, flagName := range flagNames {

			if arg == flagName && len(args)-1 >= i+1 {
				return args[i+1]
			}

			if strings.HasPrefix(arg, flagName+"=") {
				return arg[len(flagName)+1:]
			}
		}
	}

	envKey := cst.CmdRoot + "_" + flag
	envKey = strings.Replace(envKey, ".", "_", -1)
	envKey = strings.Replace(envKey, "-", "_", -1)
	envKey = strings.ToUpper(envKey)
	return os.Getenv(envKey)
}
