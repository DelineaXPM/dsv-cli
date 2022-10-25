package main

import (
	"io"
	"log"
	"os"
	"runtime/debug"

	cmd "thy/commands"
	cst "thy/constants"
	"thy/format"
	"thy/utils"
	"thy/version"

	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

func main() {
	defer func() {
		if v := recover(); v != nil {
			out := format.NewDefaultOutClient()
			if viper.GetString(cst.Verbose) == "" {
				out.FailS("An unknown error occurred. Use the verbose flag (-v) to see more information.")
			} else {
				err, ok := v.(error)
				if ok {
					out.Fail(err)
				}
				out.FailS(string(debug.Stack()))
			}
		}
	}()

	exitStatus, err := runCLI(os.Args)
	if err != nil {
		format.NewDefaultOutClient().Fail(err)
		os.Exit(utils.GetExecStatus(err))
	}
	os.Exit(exitStatus)
}

func runCLI(args []string) (exitStatus int, err error) {
	c := cli.NewCLI(cst.CmdRoot, version.Version)
	c.Args = args[1:]
	c.Name = cst.CmdRoot
	c.Commands = map[string]cli.CommandFactory{
		"secret":                        cmd.GetSecretCmd,
		"secret read":                   cmd.GetSecretReadCmd,
		"secret describe":               cmd.GetSecretDescribeCmd,
		"secret search":                 cmd.GetSecretSearchCmd,
		"secret delete":                 cmd.GetSecretDeleteCmd,
		"secret restore":                cmd.GetSecretRestoreCmd,
		"secret create":                 cmd.GetSecretCreateCmd,
		"secret update":                 cmd.GetSecretUpdateCmd,
		"secret rollback":               cmd.GetSecretRollbackCmd,
		"secret edit":                   cmd.GetSecretEditCmd,
		"secret bustcache":              cmd.GetSecretBustCacheCmd,
		"policy":                        cmd.GetPolicyCmd,
		"policy read":                   cmd.GetPolicyReadCmd,
		"policy search":                 cmd.GetPolicySearchCmd,
		"policy delete":                 cmd.GetPolicyDeleteCmd,
		"policy restore":                cmd.GetPolicyRestoreCmd,
		"policy create":                 cmd.GetPolicyCreateCmd,
		"policy edit":                   cmd.GetPolicyEditCmd,
		"policy update":                 cmd.GetPolicyUpdateCmd,
		"policy rollback":               cmd.GetPolicyRollbackCmd,
		"auth":                          cmd.GetAuthCmd,
		"auth clear":                    cmd.GetAuthClearCmd,
		"auth list":                     cmd.GetAuthListCmd,
		"auth change-password":          cmd.GetAuthChangePasswordCmd,
		"user":                          cmd.GetUserCmd,
		"user read":                     cmd.GetUserReadCmd,
		"user search":                   cmd.GetUserSearchCmd,
		"user delete":                   cmd.GetUserDeleteCmd,
		"user restore":                  cmd.GetUserRestoreCmd,
		"user create":                   cmd.GetUserCreateCmd,
		"user update":                   cmd.GetUserUpdateCmd,
		"whoami":                        cmd.GetWhoAmICmd,
		"eval":                          cmd.GetEvaluateFlagCmd,
		"cli-config":                    cmd.GetCliConfigCmd,
		"init":                          cmd.GetCliConfigInitCmd,
		"cli-config init":               cmd.GetCliConfigInitCmd,
		"cli-config update":             cmd.GetCliConfigUpdateCmd,
		"cli-config clear":              cmd.GetCliConfigClearCmd,
		"cli-config read":               cmd.GetCliConfigReadCmd,
		"cli-config edit":               cmd.GetCliConfigEditCmd,
		"cli-config use-profile":        cmd.GetCliConfigUseProfileCmd,
		"config":                        cmd.GetConfigCmd,
		"config read":                   cmd.GetConfigReadCmd,
		"config update":                 cmd.GetConfigUpdateCmd,
		"config edit":                   cmd.GetConfigEditCmd,
		"config auth-provider":          cmd.GetAuthProviderCmd,
		"config auth-provider read":     cmd.GetAuthProviderReadCmd,
		"config auth-provider search":   cmd.GetAuthProviderSearchCmd,
		"config auth-provider delete":   cmd.GetAuthProviderDeleteCmd,
		"config auth-provider restore":  cmd.GetAuthProviderRestoreCmd,
		"config auth-provider create":   cmd.GetAuthProviderCreateCmd,
		"config auth-provider update":   cmd.GetAuthProviderUpdateCmd,
		"config auth-provider edit":     cmd.GetAuthProviderEditCmd,
		"config auth-provider rollback": cmd.GetAuthProviderRollbackCmd,
		"role":                          cmd.GetRoleCmd,
		"role read":                     cmd.GetRoleReadCmd,
		"role search":                   cmd.GetRoleSearchCmd,
		"role delete":                   cmd.GetRoleDeleteCmd,
		"role restore":                  cmd.GetRoleRestoreCmd,
		"role create":                   cmd.GetRoleCreateCmd,
		"role update":                   cmd.GetRoleUpdateCmd,
		"client":                        cmd.GetClientCmd,
		"client read":                   cmd.GetClientReadCmd,
		"client delete":                 cmd.GetClientDeleteCmd,
		"client restore":                cmd.GetClientRestoreCmd,
		"client create":                 cmd.GetClientCreateCmd,
		"client search":                 cmd.GetClientSearchCmd,
		"usage":                         cmd.GetUsageCmd,
		"audit":                         cmd.GetAuditSearchCmd,
		"group":                         cmd.GetGroupCmd,
		"group read":                    cmd.GetGroupReadCmd,
		"group create":                  cmd.GetGroupCreateCmd,
		"group delete":                  cmd.GetGroupDeleteCmd,
		"group restore":                 cmd.GetGroupRestoreCmd,
		"group search":                  cmd.GetGroupSearchCmd,
		"group add-members":             cmd.GetAddMembersCmd,
		"user groups":                   cmd.GetMemberGroupsCmd,
		"group delete-members":          cmd.GetDeleteMembersCmd,
		"pki":                           cmd.GetPkiCmd,
		"pki register":                  cmd.GetPkiRegisterCmd,
		"pki sign":                      cmd.GetPkiSignCmd,
		"pki leaf":                      cmd.GetPkiLeafCmd,
		"pki generate-root":             cmd.GetPkiGenerateRootCmd,
		"pki ssh-cert":                  cmd.GetPkiSSHCertCmd,
		"siem":                          cmd.GetSiemCmd,
		"siem create":                   cmd.GetSiemCreateCmd,
		"siem update":                   cmd.GetSiemUpdateCmd,
		"siem read":                     cmd.GetSiemReadCmd,
		"siem search":                   cmd.GetSiemSearchCmd,
		"siem delete":                   cmd.GetSiemDeleteCmd,
		"home":                          cmd.GetHomeCmd,
		"home read":                     cmd.GetHomeReadCmd,
		"home create":                   cmd.GetHomeCreateCmd,
		"home update":                   cmd.GetHomeUpdateCmd,
		"home delete":                   cmd.GetHomeDeleteCmd,
		"home search":                   cmd.GetHomeSearchCmd,
		"home describe":                 cmd.GetHomeDescribeCmd,
		"home edit":                     cmd.GetHomeEditCmd,
		"home rollback":                 cmd.GetHomeRollbackCmd,
		"home restore":                  cmd.GetHomeRestoreCmd,
		"pool":                          cmd.GetPoolCmd,
		"pool create":                   cmd.GetPoolCreateCmd,
		"pool read":                     cmd.GetPoolReadCmd,
		"pool list":                     cmd.GetPoolListCmd,
		"pool delete":                   cmd.GetPoolDeleteCmd,
		"engine":                        cmd.GetEngineCmd,
		"engine read":                   cmd.GetEngineReadCmd,
		"engine list":                   cmd.GetEngineListCmd,
		"engine create":                 cmd.GetEngineCreateCmd,
		"engine delete":                 cmd.GetEngineDeleteCmd,
		"engine ping":                   cmd.GetEnginePingCmd,
		"crypto":                        cmd.GetCryptoCmd,
		"crypto auto":                   cmd.GetCryptoCmd,
		"crypto key-create":             cmd.GetAutoKeyCreateCmd,
		"crypto key-read":               cmd.GetAutoKeyReadMetadataCmd,
		"crypto key-delete":             cmd.GetAutoKeyDeleteCmd,
		"crypto key-restore":            cmd.GetAutoKeyRestoreCmd,
		"crypto encrypt":                cmd.GetEncryptCmd,
		"crypto decrypt":                cmd.GetDecryptCmd,
		"crypto rotate":                 cmd.GetEncryptionRotateCmd,
		"crypto auto key-create":        cmd.GetAutoKeyCreateCmd,
		"crypto auto key-read":          cmd.GetAutoKeyReadMetadataCmd,
		"crypto auto key-delete":        cmd.GetAutoKeyDeleteCmd,
		"crypto auto key-restore":       cmd.GetAutoKeyRestoreCmd,
		"crypto auto encrypt":           cmd.GetEncryptCmd,
		"crypto auto decrypt":           cmd.GetDecryptCmd,
		"crypto auto rotate":            cmd.GetEncryptionRotateCmd,
		"crypto manual":                 cmd.GetCryptoManualCmd,
		"crypto manual key-upload":      cmd.GetManualKeyUploadCmd,
		"crypto manual key-update":      cmd.GetManualKeyUpdateCmd,
		"crypto manual key-read":        cmd.GetManualKeyReadCmd,
		"crypto manual key-delete":      cmd.GetManualKeyDeleteCmd,
		"crypto manual key-restore":     cmd.GetManualKeyRestoreCmd,
		"crypto manual encrypt":         cmd.GetManualKeyEncryptCmd,
		"crypto manual decrypt":         cmd.GetManualKeyDecryptCmd,
		"report":                        cmd.GetReportCmd,
		"report secret":                 cmd.GetSecretReportCmd,
		"report group":                  cmd.GetGroupReportCmd,
		"breakglass":                    cmd.GetBreakGlassCmd,
		"breakglass status":             cmd.GetBreakGlassGetStatusCmd,
		"breakglass generate":           cmd.GetBreakGlassGenerateCmd,
		"breakglass apply":              cmd.GetBreakGlassApplyCmd,
	}

	c.Autocomplete = true
	c.AutocompleteInstall = "install"
	c.AutocompleteUninstall = "uninstall"
	log.SetOutput(io.Discard)

	return c.Run()
}
