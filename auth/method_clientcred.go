package auth

import (
	cst "thy/constants"
	"thy/errors"
	"thy/store"

	"github.com/spf13/viper"
)

func buildClientcredParams() (*requestBody, error) {
	data := &requestBody{
		GrantType:    authTypeToGrantType[ClientCredential],
		AuthClientID: viper.GetString(cst.AuthClientID),
	}
	if secret, err := store.GetSecureSetting(cst.AuthClientSecret); err != nil || secret == "" {
		if err == nil {
			err = errors.NewS("auth-client-secret setting is empty")
		}
		return nil, err.Grow("Failed to retrieve secure setting: auth-client-secret")
	} else {
		data.AuthClientSecret = secret
	}

	return data, nil
}
