package cmd

import (
	"fmt"
	"strings"

	cst "thy/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/internal/predictor"
	"github.com/DelineaXPM/dsv-cli/paths"
	"github.com/DelineaXPM/dsv-cli/utils"
	"github.com/DelineaXPM/dsv-cli/vaultcli"

	"github.com/mitchellh/cli"
	"github.com/shurcooL/graphql"
	"github.com/spf13/viper"
)

func GetReportCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounReport},
		SynopsisText: "Show report records for secrets and groups",
		NoConfigRead: true,
		NoPreAuth:    true,
	})
}

func GetSecretReportCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounReport, cst.NounSecret},
		SynopsisText: "report secret",
		HelpText: fmt.Sprintf(`Read secret report records

Usage:
   • %[1]s %[7]s --%[2]s /x/y/z --%[3]s testUser --%[4]s sysadmins --%[5]s role1 --%[6]s 5
   • %[1]s %[7]s --%[2]s /x/y/z
   `, cst.NounReport, cst.Path, cst.NounUser, cst.NounGroup, cst.NounRole, cst.Limit, cst.NounSecret),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Path, Usage: "Path (optional)"},
			{Name: cst.NounUser, Usage: "User name (optional)"},
			{Name: cst.NounGroup, Usage: "Group name (optional)"},
			{Name: cst.NounRole, Usage: "Role name (optional)"},
			{Name: cst.Limit, Shorthand: "l", Usage: cst.LimitHelpMessage},
			{Name: cst.OffSet, Usage: "Offset for the next secrets (optional)"},
			{Name: cst.Cursor, Usage: cst.CursorHelpMessage},
		},
		RunFunc: handleSecretReport,
	})
}

func GetGroupReportCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounReport, cst.NounGroup},
		SynopsisText: "report group",
		HelpText: fmt.Sprintf(`Read group report records

Usage:
   • %[1]s %[4]s --%[2]s testUser --%[3]s 5
   • %[1]s %[4]s --%[2]s testUser
   `, cst.NounReport, cst.NounUser, cst.Limit, cst.NounGroup),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.NounUser, Usage: "User name (optional)"},
			{Name: cst.Limit, Shorthand: "l", Usage: cst.LimitHelpMessage},
			{Name: cst.OffSet, Usage: "Offset for the next groups (optional)"},
		},
		RunFunc: handleGroupReport,
	})
}

func handleSecretReport(vcli vaultcli.CLI, args []string) int {
	user := viper.GetString(cst.NounUser)
	group := viper.GetString(cst.NounGroup)
	role := viper.GetString(cst.NounRole)
	limit := viper.GetInt(cst.Limit)
	offset := viper.GetInt(cst.OffSet)
	cursor := viper.GetString(cst.Cursor)

	path := viper.GetString(cst.Path)
	if len(args) > 0 && !strings.HasPrefix(args[0], "--") {
		path = args[0]
	}

	switch {
	case user != "":
		return reportUserSecret(vcli, user, path, limit, offset, cursor)
	case group != "":
		return reportGroupSecret(vcli, group, path, limit, offset)
	case role != "":
		return reportRoleSecret(vcli, role, path, limit, offset)
		//read sign in user record
	default:
		return reportSignInUserSecret(vcli, path, limit, offset, cursor)
	}
}

func handleGroupReport(vcli vaultcli.CLI, args []string) int {
	user := viper.GetString(cst.NounUser)
	limit := viper.GetInt(cst.Limit)
	offset := viper.GetInt(cst.OffSet)

	switch {
	case user != "":
		return reportUserGroup(vcli, user, limit, offset)

	default:
		return reportSignInGroup(vcli, limit, offset)
	}
}

func reportRoleSecret(vcli vaultcli.CLI, role, path string, limit, offset int) int {

	var data []byte
	var err *errors.ApiError

	uri := paths.CreateURI("report/query", nil)
	switch path {
	case "<.*>", ".*":
		var query struct {
			Role struct {
				GQRole
				Secrets Secrets `graphql:"secrets(filter:{limit:$limit, secretType: SECRET, offset:$offset})"`
				Home    Home    `graphql:"home(limit:$limit)"`
			} `graphql:"role(name: $role)" json:"data"`
		}
		data, err = vcli.GraphQLClient().DoRequest(uri, &query, map[string]interface{}{
			"limit":  graphql.Int(limit),
			"role":   graphql.String(role),
			"offset": graphql.Int(offset),
		})

	case "home:<.*>", "home:.*":
		var query struct {
			Role struct {
				GQRole
				Home Home `graphql:"home(limit:$limit, cursor:$cursor)"`
			} `graphql:"role(name: $role)" json:"data"`
		}

		data, err = vcli.GraphQLClient().DoRequest(uri, &query, map[string]interface{}{
			"limit":  graphql.Int(limit),
			"role":   graphql.String(role),
			"cursor": graphql.String(""),
		})

	default:
		variables := map[string]interface{}{
			"limit":  graphql.Int(limit),
			"path":   graphql.String(path),
			"role":   graphql.String(role),
			"offset": graphql.Int(offset),
		}
		if strings.HasSuffix(path, "<.*>") || strings.HasSuffix(path, ".*") {
			var query struct {
				Role struct {
					GQRole
					Secrets Secrets `graphql:"secrets(filter:{limit:$limit, secretType: SECRET, offset:$offset, path:$path})"`
				} `graphql:"role(name: $role)" json:"data"`
			}

			data, err = vcli.GraphQLClient().DoRequest(uri, &query, variables)

		} else {
			var query struct {
				Role struct {
					GQRole
					Secrets Secrets `graphql:"secrets(filter:{limit:$limit, secretType: ALL, offset:$offset, path:$path})"`
				} `graphql:"role(name: $role)" json:"data"`
			}

			data, err = vcli.GraphQLClient().DoRequest(uri, &query, variables)
		}
	}

	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func reportGroupSecret(vcli vaultcli.CLI, group, path string, limit, offset int) int {

	var data []byte
	var err *errors.ApiError

	uri := paths.CreateURI("report/query", nil)
	switch path {
	case "<.*>", ".*":
		var query struct {
			Group struct {
				GQGroup
				Secrets Secrets `graphql:"secrets(filter:{limit:$limit, secretType: SECRET, offset:$offset})"`
			} `graphql:"group(groupName: $group)" json:"data"`
		}
		data, err = vcli.GraphQLClient().DoRequest(uri, &query, map[string]interface{}{
			"limit":  graphql.Int(limit),
			"group":  graphql.String(group),
			"offset": graphql.Int(offset),
		})

	default:
		variables := map[string]interface{}{
			"limit":  graphql.Int(limit),
			"path":   graphql.String(path),
			"group":  graphql.String(group),
			"offset": graphql.Int(offset),
		}
		if strings.HasSuffix(path, "<.*>") || strings.HasSuffix(path, ".*") {
			var query struct {
				Group struct {
					GQGroup
					Secrets Secrets `graphql:"secrets(filter:{limit:$limit, secretType: SECRET, offset:$offset, path:$path})"`
				} `graphql:"group(groupName: $group)" json:"data"`
			}

			data, err = vcli.GraphQLClient().DoRequest(uri, &query, variables)

		} else {
			var query struct {
				Group struct {
					GQGroup
					Secrets Secrets `graphql:"secrets(filter:{limit:$limit, secretType: ALL, offset:$offset, path:$path})"`
				} `graphql:"group(groupName: $group)" json:"data"`
			}

			data, err = vcli.GraphQLClient().DoRequest(uri, &query, variables)
		}
	}

	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func reportUserSecret(vcli vaultcli.CLI, user, path string, limit, offset int, cursor string) int {
	var data []byte
	var err *errors.ApiError

	uri := paths.CreateURI("report/query", nil)
	switch path {
	case "<.*>", ".*":
		var query struct {
			User struct {
				GQUser
				Secrets Secrets `graphql:"secrets(filter:{limit:$limit, secretType: SECRET, offset:$offset})"`
				Home    Home    `graphql:"home(limit:$limit), cursor:$cursor)"`
			} `graphql:"user(name: $user)" json:"data"`
		}
		data, err = vcli.GraphQLClient().DoRequest(uri, &query, map[string]interface{}{
			"limit":  graphql.Int(limit),
			"user":   graphql.String(user),
			"offset": graphql.Int(offset),
			"cursor": graphql.String(cursor),
		})

	case "home:<.*>", "home:.*":
		var query struct {
			User struct {
				GQUser
				Home Home `graphql:"home(limit:$limit, cursor:$cursor)"`
			} `graphql:"user(name: $user)" json:"data"`
		}

		data, err = vcli.GraphQLClient().DoRequest(uri, &query, map[string]interface{}{
			"limit":  graphql.Int(limit),
			"user":   graphql.String(user),
			"cursor": graphql.String(cursor),
		})

	default:
		variables := map[string]interface{}{
			"limit":  graphql.Int(limit),
			"path":   graphql.String(path),
			"user":   graphql.String(user),
			"offset": graphql.Int(offset),
		}
		if strings.HasSuffix(path, "<.*>") || strings.HasSuffix(path, ".*") {
			var query struct {
				User struct {
					GQUser
					Secrets Secrets `graphql:"secrets(filter:{limit:$limit, secretType: SECRET, offset:$offset, path:$path})"`
				} `graphql:"user(name: $user)" json:"data"`
			}

			data, err = vcli.GraphQLClient().DoRequest(uri, &query, variables)

		} else {
			var query struct {
				User struct {
					GQUser
					Secrets Secrets `graphql:"secrets(filter:{limit:$limit, secretType: ALL, offset:$offset, path:$path})"`
				} `graphql:"user(name: $user)" json:"data"`
			}

			data, err = vcli.GraphQLClient().DoRequest(uri, &query, variables)
		}
	}

	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func reportSignInUserSecret(vcli vaultcli.CLI, path string, limit, offset int, cursor string) int {
	var data []byte
	var err *errors.ApiError

	uri := paths.CreateURI("report/me", nil)
	switch path {
	case "<.*>", ".*":
		var query struct {
			Me struct {
				GQUser
				Secrets Secrets `graphql:"secrets(filter:{limit:$limit, secretType: SECRET, offset:$offset})"`
				Home    Home    `graphql:"home(limit:$limit, cursor:$cursor)"`
			} `json:"data"`
		}
		data, err = vcli.GraphQLClient().DoRequest(uri, &query, map[string]interface{}{
			"limit":  graphql.Int(limit),
			"offset": graphql.Int(offset),
			"cursor": graphql.String(cursor),
		})

	case "home:<.*>", "home:.*":
		var query struct {
			Me struct {
				GQUser
				Home Home `graphql:"home(limit:$limit, cursor:$cursor)"`
			} `json:"data"`
		}

		data, err = vcli.GraphQLClient().DoRequest(uri, &query, map[string]interface{}{
			"limit":  graphql.Int(limit),
			"cursor": graphql.String(cursor),
		})

	default:
		variables := map[string]interface{}{
			"limit":  graphql.Int(limit),
			"path":   graphql.String(path),
			"offset": graphql.Int(offset),
		}
		if strings.HasSuffix(path, "<.*>") || strings.HasSuffix(path, ".*") {
			var query struct {
				Me struct {
					GQUser
					Secrets Secrets `graphql:"secrets(filter:{limit:$limit, secretType: SECRET, offset:$offset, path:$path})"`
				}
			}

			data, err = vcli.GraphQLClient().DoRequest(uri, &query, variables)

		} else {
			var query struct {
				Me struct {
					GQUser
					Secrets Secrets `graphql:"secrets(filter:{limit:$limit, secretType: ALL, offset:$offset, path:$path})"`
				}
			}

			data, err = vcli.GraphQLClient().DoRequest(uri, &query, variables)
		}
	}

	vcli.Out().WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func reportSignInGroup(vcli vaultcli.CLI, limit, offset int) int {
	var query struct {
		Me struct {
			GQUser
			MemberOf MemberOf `graphql:"memberOf(limit:$limit, offset:$offset)"`
		} `json:"data"`
	}

	data, err := vcli.GraphQLClient().DoRequest(
		paths.CreateURI("report/me", nil), &query, map[string]interface{}{
			"limit":  graphql.Int(limit),
			"offset": graphql.Int(offset),
		})
	vcli.Out().WriteResponse(data, nil)
	return utils.GetExecStatus(err)
}

func reportUserGroup(vcli vaultcli.CLI, user string, limit, offset int) int {
	var query struct {
		User struct {
			GQUser
			MemberOf MemberOf `graphql:"memberOf(limit:$limit, offset:$offset)"`
		} `graphql:"user(name: $user)" json:"data"`
	}

	variables := map[string]interface{}{
		"user":   graphql.String(user),
		"limit":  graphql.Int(limit),
		"offset": graphql.Int(offset),
	}

	uri := paths.CreateURI("report/query", nil)
	data, err := vcli.GraphQLClient().DoRequest(uri, &query, variables)
	vcli.Out().WriteResponse(data, nil)
	return utils.GetExecStatus(err)
}

type Secrets struct {
	Secrets []struct {
		Actions        []graphql.String
		ID             graphql.String
		Path           graphql.String
		Created        graphql.String
		LastModified   graphql.String
		LastModifiedBy graphql.String
		Version        graphql.String
	}
	Pagination struct {
		Offset   graphql.Int
		Limit    graphql.Int
		PageSize graphql.Int
	}
}

type MemberOf struct {
	Memberships []struct {
		Group struct {
			ID   graphql.String
			Name graphql.String
		}
		Since     graphql.String
		CreatedBy graphql.String
	}
	Pagination struct {
		Offset     graphql.Int
		Limit      graphql.Int
		PageSize   graphql.Int
		TotalItems graphql.Int
	}
}

type Home struct {
	Secrets []struct {
		ID             graphql.String
		Path           graphql.String
		Created        graphql.String
		LastModified   graphql.String
		LastModifiedBy graphql.String
		Version        graphql.String
	}
	Pagination struct {
		Cursor graphql.String
		Limit  graphql.Int
	}
}

type GQUser struct {
	UserName       graphql.String
	Provider       graphql.String
	Created        graphql.String
	LastModified   graphql.String
	CreatedBy      graphql.String
	LastModifiedBy graphql.String
	Version        graphql.String
}

type GQGroup struct {
	Name           graphql.String
	Created        graphql.String
	LastModified   graphql.String
	CreatedBy      graphql.String
	LastModifiedBy graphql.String
	Version        graphql.String
}

type GQRole struct {
	Name           graphql.String
	ExternalID     graphql.String
	Provider       graphql.String
	Created        graphql.String
	LastModified   graphql.String
	CreatedBy      graphql.String
	LastModifiedBy graphql.String
	Version        graphql.String
	Description    graphql.String
}
