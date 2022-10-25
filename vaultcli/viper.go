package vaultcli

import (
	"fmt"
	"strings"
	cst "thy/constants"

	"github.com/spf13/viper"
)

const envVarPrefix = "thy"

func ViperInit() error {
	viper.SetEnvPrefix(envVarPrefix)
	envReplacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(envReplacer)
	viper.AutomaticEnv()

	cfgFile := viper.GetString(cst.Config)

	cf, err := ReadConfigFile(cfgFile)
	if err != nil {
		return err
	}

	profile := viper.GetString(cst.Profile)
	if profile == "" {
		profile = cf.DefaultProfile
	}

	// Set profile name to lower case globally.
	profile = strings.ToLower(profile)
	viper.Set(cst.Profile, profile)

	config, ok := cf.GetProfile(profile)
	if !ok {
		return fmt.Errorf("profile %q not found in configuration file %q", profile, cf.GetPath())
	}

	err = viper.MergeConfigMap(config.data)
	if err != nil {
		return fmt.Errorf("cannot initialize Viper: %w", err)
	}

	return nil
}
