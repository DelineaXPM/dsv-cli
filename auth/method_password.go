package auth

import (
	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/internal/store"

	"github.com/spf13/viper"
)

const authSP = "auth.securePassword"

func buildPasswordParams() (*requestBody, error) {
	data := &requestBody{
		GrantType: authTypeToGrantType[Password],
		Username:  viper.GetString(cst.Username),
		Password:  viper.GetString(cst.Password),
		Provider:  viper.GetString(cst.AuthProvider),
	}

	// If plaintext Password exists, that means Viper retrieves it from memory. Use this Password to authenticate.
	// If it is an empty string, look for SecurePassword, which Viper gets only from config. Get the corresponding
	// key file and use it to decrypt SecurePassword.
	if data.Password == "" {
		passSetting := authSP
		storeType := viper.GetString(cst.StoreType)
		if storeType == store.WinCred || storeType == store.PassLinux {
			passSetting = cst.Password
		}
		if pass, err := store.GetSecureSetting(passSetting); err == nil && pass != "" {
			if passSetting == authSP {
				keyPath := GetEncryptionKeyFilename(viper.GetString(cst.Tenant), data.Username)
				key, err := store.ReadFileInDefaultPath(keyPath)
				if err != nil || key == "" {
					return nil, KeyfileNotFoundError
				}
				decrypted, decryptionErr := Decrypt(pass, key)
				if decryptionErr != nil {
					return nil, errors.NewS("Failed to decrypt the password with key.")
				}
				data.Password = decrypted
			} else {
				data.Password = pass
			}
		}
	}
	return data, nil
}
