package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/DelineaXPM/dsv-cli/auth"
	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/format"
	"github.com/DelineaXPM/dsv-cli/paths"
	"github.com/DelineaXPM/dsv-cli/utils"
	"github.com/DelineaXPM/dsv-cli/vaultcli"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

func GetAuthCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounAuth},
		NoPreAuth:    true,
		SynopsisText: "Get auth token, manage auth cache or change password",
		HelpText: fmt.Sprintf(`Authenticate with %[2]s

Usage:
   • auth --profile staging
   • auth --auth-username %[3]s --auth-password %[4]s
   • auth --auth-type %[5]s --auth-client-id %[6]s --domain %[7]s --auth-client-secret %[8]s
`, cst.NounAuth, cst.ProductName, cst.ExampleUser, "************", cst.ExampleAuthType, cst.ExampleAuthClientID, cst.ExampleDomain, cst.ExampleAuthClientSecret, string(auth.FederatedAws)),
		RunFunc: handleAuth,
	})
}

func GetAuthClearCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounAuth, cst.Clear},
		SynopsisText: fmt.Sprintf("%s %s", cst.NounAuth, cst.Clear),
		HelpText: fmt.Sprintf(`Clear %[1]s %[3]ss from %[2]s

Usage:
   • auth clear
`, cst.NounAuth, cst.ProductName, cst.NounToken),
		NoPreAuth: true,
		RunFuncE:  handleAuthClear,
	})
}

func GetAuthListCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounAuth, cst.List},
		SynopsisText: fmt.Sprintf("%s %s", cst.NounAuth, cst.List),
		HelpText: fmt.Sprintf(`List %[1]s %[3]ss from %[2]s

Usage:
   • auth list
`, cst.NounAuth, cst.ProductName, cst.NounToken),
		NoPreAuth: true,
		RunFuncE:  handleAuthList,
	})
}

func GetAuthChangePasswordCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounAuth, "change-password"},
		SynopsisText: fmt.Sprintf("%s %s", cst.NounAuth, "change-password"),
		HelpText: `Change user password

Usage:
   • auth change-password`,
		RunFunc: handleAuthChangePassword,
	})
}

func handleAuth(vcli vaultcli.CLI, args []string) int {
	var data []byte

	username := vaultcli.GetFlagVal(cst.Username, args)
	if username != "" {
		// We rely on the password auth type being set in order to trigger that flow later
		viper.Set(cst.AuthType, "password")
		password := viper.GetString(cst.Password)
		if password == "" {
			passwordPrompt := &survey.Password{Message: "Please enter password:"}
			survErr := survey.AskOne(passwordPrompt, &password, survey.WithValidator(survey.Required))
			if survErr != nil {
				vcli.Out().WriteResponse(nil, errors.New(survErr))
				return 652
			}
			viper.Set(cst.Password, password)
		}

		// We may have loaded an auth provider from the configuration file, and need to make sure
		// that we rely solely on the args provided by the user here
		authTypeArgIdx := utils.IndexOf(args, "--auth-provider")
		if authTypeArgIdx >= 0 {
			viper.Set(cst.AuthProvider, args[authTypeArgIdx+1])
		} else {
			viper.Set(cst.AuthProvider, "")
		}

		token, apiErr := vcli.Authenticator().GetToken()
		if apiErr == nil {
			data, apiErr = errors.Convert(format.JsonMarshal(token))
			vcli.Out().WriteResponse(data, apiErr)
		} else {
			vcli.Out().WriteResponse(data, apiErr)
			return 516
		}
	} else {
		token, apiErr := vcli.Authenticator().GetToken()
		if apiErr == nil {
			data, apiErr = errors.Convert(format.JsonMarshal(token))
			vcli.Out().WriteResponse(data, apiErr)
		} else {
			vcli.Out().WriteResponse(data, apiErr)
			return 427
		}
	}
	return 0
}

func handleAuthClear(vcli vaultcli.CLI, args []string) error {
	st := viper.GetString(cst.StoreType)
	s, err := vcli.Store(st)
	if err != nil {
		return err
	}
	if err := s.Wipe(cst.TokenRoot); err != nil {
		return err
	}
	log.Print("Successfully cleared local cache")
	return nil
}

func handleAuthList(vcli vaultcli.CLI, args []string) error {
	st := viper.GetString(cst.StoreType)
	s, err := vcli.Store(st)
	if err != nil {
		return err
	}
	keys, err := s.List(cst.TokenRoot)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	for i := range keys {
		keys[i] = strings.ReplaceAll(keys[i], "%2D", "-")
	}
	keysBytes := []byte(strings.Join(keys, "\n"))
	vcli.Out().WriteResponse(keysBytes, nil)
	return nil
}

func handleAuthChangePassword(vcli vaultcli.CLI, args []string) int {
	var currentPassword, newPassword string

	passwordPrompt := &survey.Password{Message: "Please enter your current password:"}
	survErr := survey.AskOne(passwordPrompt, &currentPassword, survey.WithValidator(survey.Required))
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return 1
	}

	passwordPrompt = &survey.Password{Message: "Please enter the new password:"}
	survErr = survey.AskOne(passwordPrompt, &newPassword, survey.WithValidator(survey.Required))
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return 1
	}

	passwordPrompt = &survey.Password{Message: "Please enter the new password (confirm):"}
	passwordValidation := func(ans interface{}) error {
		if ans.(string) != newPassword {
			return errors.NewS("Inputs do not match. Please retry.")
		}
		return nil
	}
	survErr = survey.AskOne(passwordPrompt, &newPassword, survey.WithValidator(passwordValidation))
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return 1
	}

	user := viper.GetString(cst.Username)
	if user == "" {
		if auth.AuthType(viper.GetString(cst.AuthType)) == auth.FederatedAws || auth.AuthType(viper.GetString(cst.AuthType)) == auth.FederatedAzure {
			vcli.Out().FailS("Error: cannot change password for external user - change password with your cloud provider.")
		} else {
			vcli.Out().FailS("Error: cannot get current user from config.")
		}
		return 1
	}
	if provider := viper.GetString(cst.AuthProvider); provider != "" {
		user = provider + ":" + user
	}

	body := map[string]string{cst.CurrentPassword: currentPassword, cst.NewPassword: newPassword}
	template := fmt.Sprintf("%s/%s/%s", cst.NounUsers, user, cst.PasswordKey)
	uri := paths.CreateURI(template, nil)
	resp, apiError := vcli.HTTPClient().DoRequest(http.MethodPost, uri, &body)

	if apiError == nil {
		viper.Set(cst.Key, cst.Password)
		viper.Set(cst.Value, newPassword)
		if n := handleCliConfigUpdateCmd(vcli, nil); n != 0 {
			apiError = errors.NewS("Error while saving the new password to the config.")
			resp = []byte("Please reinitialize with your new password.")
		}
	}

	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}
