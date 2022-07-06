package cliconfig

import (
	"strings"

	cst "thy/constants"
	terrors "thy/errors"
	"thy/store"

	"github.com/spf13/viper"
)

func getSecureSettingForProfile(key string, profile string) (string, *terrors.ApiError) {
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
	return getSecureSettingForProfile(key, profile)
}
