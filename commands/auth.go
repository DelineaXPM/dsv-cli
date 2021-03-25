package cmd

import (
	"fmt"
	"os"
	"strings"

	"thy/auth"
	cst "thy/constants"
	"thy/errors"
	"thy/format"
	"thy/paths"

	"thy/requests"
	"thy/store"
	"thy/utils"

	"github.com/howeyc/gopass"

	"github.com/apex/log"
	"github.com/spf13/viper"
	"github.com/thycotic-rd/cli"
)

type AuthCommand struct {
	outClient format.OutClient
	token     func() auth.Authenticator
	getStore  func(stString string) (store.Store, *errors.ApiError)
	request   requests.Client
}

func GetAuthCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounAuth},
		RunFunc: AuthCommand{
			nil,
			auth.NewAuthenticatorDefault,
			store.GetStore, nil}.handleAuth,
		NoPreAuth:    true,
		SynopsisText: fmt.Sprintf("%s", cst.NounAuth),
		HelpText: fmt.Sprintf(`Authenticate with %[2]s

Usage:
   • auth --profile staging
   • auth --auth-username %[3]s --auth-password %[4]s
   • auth --auth-type %[7]s --auth-client-id=%[5]s --auth-client-secret=%[6]s 
		`, cst.NounAuth, cst.ProductName, cst.ExampleUser, cst.ExamplePassword, cst.ExampleAuthClientID, cst.ExampleAuthClientSecret, string(auth.FederatedAws)),
	})
}

func GetAuthClearCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounAuth, cst.Clear},
		RunFunc: AuthCommand{
			nil,
			auth.NewAuthenticatorDefault,
			store.GetStore, nil}.handleAuthClear,
		SynopsisText: fmt.Sprintf("%s %s", cst.NounAuth, cst.Clear),
		HelpText: fmt.Sprintf(`Clear %[1]s %[3]ss from %[2]s

Usage:
   • auth clear
		`, cst.NounAuth, cst.ProductName, cst.NounToken),
		NoPreAuth: true,
	})
}

func GetAuthListCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounAuth, cst.List},
		RunFunc: AuthCommand{
			nil,
			auth.NewAuthenticatorDefault,
			store.GetStore, nil}.handleAuthList,
		SynopsisText: fmt.Sprintf("%s %s", cst.NounAuth, cst.List),
		HelpText: fmt.Sprintf(`List %[1]s %[3]ss from %[2]s

Usage:
   • auth list
		`, cst.NounAuth, cst.ProductName, cst.NounToken),
		NoPreAuth: true,
	})
}

func GetAuthChangePasswordCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounAuth, cst.ChangePassword},
		RunFunc: AuthCommand{
			nil,
			auth.NewAuthenticatorDefault,
			store.GetStore,
			requests.NewHttpClient()}.handleAuthChangePassword,
		SynopsisText: fmt.Sprintf("%s %s", cst.NounAuth, cst.ChangePassword),
		HelpText: `Change user password

Usage:
   • auth change-password`,
	})
}

func (ac AuthCommand) handleAuth(args []string) int {
	var data []byte
	token, err := ac.token().GetToken()
	if err == nil {
		data, err = errors.Convert(format.JsonMarshal(token))
	}
	outClient := ac.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}

	outClient.WriteResponse(data, err)
	return 0
}

func (ac AuthCommand) handleAuthClear(args []string) int {
	var err *errors.ApiError
	var s store.Store
	st := viper.GetString(cst.StoreType)
	if s, err = ac.getStore(st); err == nil {
		err = s.Wipe(cst.TokenRoot)
	}
	if err == nil {
		log.Info("Successfully cleared local cache")
	}

	outClient := ac.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}

	outClient.WriteResponse(nil, err)
	return 0
}

func (ac AuthCommand) handleAuthList(args []string) int {
	var err *errors.ApiError
	st := viper.GetString(cst.StoreType)
	var keysBytes []byte
	if s, e := ac.getStore(st); e == nil {
		if keys, e := s.List(cst.TokenRoot); e != nil {
			err = e
		} else if len(keys) > 0 {
			for i := range keys {
				keys[i] = strings.Replace(keys[i], "%2D", "-", -1)
			}
			keysBytes = []byte(strings.Join(keys, "\n"))
		}
	} else {
		err = e
	}
	outClient := ac.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse(keysBytes, err)
	return 0
}

// PasswordUi embeds a BasicUi and overrides the AskSecret method to allow for password masking.
type PasswordUi struct {
	cli.BasicUi
}

// AskSecret prompts for password and masks it as the user types.
func (ui PasswordUi) AskSecret(query string) (string, error) {
	var password []byte
	var err error
	ui.Output(query)
	password, err = gopass.GetPasswdMasked()
	if err != nil {
		if err != gopass.ErrInterrupted {
			ui.Error(err.Error())
		}
	}
	return string(password), err
}

func (ac AuthCommand) handleAuthChangePassword(args []string) int {
	ui := &PasswordUi{
		cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
	}
	var currentPassword, newPassword string
	var err error
	if currentPassword, err = getStringAndValidate(ui, "Please enter your current password:", false, nil, true, false); err != nil {
		return 1
	}
	if newPassword, err = getStringAndValidate(ui, "Please enter the new password:", false, nil, true, true); err != nil {
		return 1
	}
	if ac.outClient == nil {
		ac.outClient = format.NewDefaultOutClient()
	}

	user := viper.GetString(cst.Username)
	if user == "" {
		if auth.AuthType(viper.GetString(cst.AuthType)) == auth.FederatedAws || auth.AuthType(viper.GetString(cst.AuthType)) == auth.FederatedAzure {
			ac.outClient.FailS("Error: cannot change password for external user - change password with your cloud provider.")
		} else {
			ac.outClient.FailS("Error: cannot get current user from config.")
		}
		return 1
	}
	if provider := viper.GetString(cst.AuthProvider); provider != "" {
		user = provider + ":" + user
	}
	resp, apiError := ac.doChangePassword(user, currentPassword, newPassword)
	if apiError == nil {
		viper.Set(cst.Key, cst.Password)
		viper.Set(cst.Value, newPassword)
		if n := handleCliConfigUpdateCmd(nil); n != 0 {
			apiError = errors.NewS("Error while saving the new password to the config.")
			resp = []byte("Please reinitialize with your new password.")
		}
	}

	ac.outClient.WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func (ac AuthCommand) doChangePassword(user, current, new string) ([]byte, *errors.ApiError) {
	body := map[string]string{cst.CurrentPassword: current, cst.NewPassword: new}
	template := fmt.Sprintf("%ss/%s/%s", cst.NounUser, user, cst.PasswordKey)
	uri := paths.CreateURI(template, nil)
	return ac.request.DoRequest("POST", uri, &body)
}
