package cliconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	if IsInstallCmd(args) {
		return nil
	}
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

	flagsToValidate := []string{cst.Tenant}
	if err := viper.ReadInConfig(profile); err != nil {
		flagsMissing := false
		for _, f := range flagsToValidate {
			if v := GetFlagBeforeParse(f, args); v == "" {
				flagsMissing = true
			}
		}
		if !flagsMissing {
			return nil
		}
		if eString := err.Error(); strings.Contains(eString, "invalid subkey") {
			return terrors.NewS(fmt.Sprintf("Invalid or non-existent profile in CLI config: %s.", profile))
		} else {
			// Do not return the error to allow users to view help text for commands.
			out := format.NewDefaultOutClient()
			out.FailS(fmt.Sprintf("Create CLI config file manually or execute command '%s init' to initiate CLI configuration - cannot find config.", cst.CmdRoot))
		}
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

	st := store.StoreType(viper.GetString(cst.StoreType))
	s, err := store.GetStore(string(st))
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

func IsInstallCmd(args []string) bool {
	for _, a := range args {
		if a == "--install" || a == "-install" {
			return true
		}
	}
	return false
}

func GetFlagBeforeParse(flag string, args []string) string {
	shortFlag := cst.GetShortFlag(flag)
	shortFlagPattern := ""
	if shortFlag != "" {
		shortFlagPattern = fmt.Sprintf("|-%s", shortFlag)
	}
	flagMatch := fmt.Sprintf("(?:--%s%s)[ =](\\S+)", flag, shortFlagPattern)
	val := ""
	re := regexp.MustCompile(flagMatch)
	match := re.FindStringSubmatch(strings.Join(args, " "))
	if len(match) > 1 {
		val = match[1]
	}
	if val == "" {
		envKey := strings.ToUpper(strings.Replace(strings.Replace(cst.CmdRoot+"_"+flag, ".", "_", -1), "-", "_", -1))
		val = os.Getenv(envKey)
	}
	return val
}
