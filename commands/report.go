package cmd

import (
	"fmt"
	"strings"
	"thy/constants"
	cst "thy/constants"
	"thy/errors"
	"thy/format"
	"thy/paths"
	preds "thy/predictors"
	"thy/requests"
	"thy/utils"

	"github.com/posener/complete"

	"github.com/shurcooL/graphql"

	"github.com/spf13/viper"

	"github.com/thycotic-rd/cli"
)

type report struct {
	graphClient requests.GraphClient
	outClient   format.OutClient
}

func GetSecretReportCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounReport},
		RunFunc:      report{graphClient: requests.NewGraphClient()}.handleSecretReport,
		SynopsisText: "report secret",
		HelpText: fmt.Sprintf(`Read secret report records

Usage:
   • %[1]s --%[2]s /x/y/z --%[3]s testUser --%[4]s sysadmins --%[5]s role1 --%[6]s 5
   • %[1]s --%[2]s 2020-01-01
   `, cst.NounReport, cst.Path, cst.NounUser, cst.NounGroup, cst.NounRole, cst.Limit),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Path):      cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Path, Usage: "Path (optional)"}), false},
			preds.LongFlag(cst.NounUser):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.NounUser, Usage: "User name (optional)"}), false},
			preds.LongFlag(cst.NounGroup): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.NounGroup, Usage: "Group name (optional)"}), false},
			preds.LongFlag(cst.NounRole):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.NounRole, Usage: "Role name (optional)"}), false},
			preds.LongFlag(cst.Limit):     cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Limit, Shorthand: "l", Usage: "Maximum number of results per cursor (optional)"}), false},
			preds.LongFlag(cst.OffSet):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.OffSet, Usage: "Offset for the next secrets (optional)"}), false},
			preds.LongFlag(cst.Cursor):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Cursor, Usage: constants.CursorHelpMessage}), false},
		},
		MinNumberArgs: 0,
	})
}

func GetGroupReportCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounReport},
		RunFunc:      report{graphClient: requests.NewGraphClient()}.handleGroupReport,
		SynopsisText: "report group",
		HelpText: fmt.Sprintf(`Read group report records

Usage:
   • %[1]s --%[2]s testUser --%[3]s 5
   • %[1]s --%[2]s 2020-01-01
   `, cst.NounReport, cst.NounUser, cst.Limit),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.NounUser): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.NounUser, Usage: "User name (optional)"}), false},
			preds.LongFlag(cst.Limit):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Limit, Shorthand: "l", Usage: "Maximum number of results per cursor (optional)"}), false},
			preds.LongFlag(cst.OffSet):   cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.OffSet, Usage: "Offset for the next groups (optional)"}), false},
		},
		MinNumberArgs: 0,
	})
}

func (r report) handleSecretReport(args []string) int {
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

	if r.outClient == nil {
		r.outClient = format.NewDefaultOutClient()
	}

	switch {
	case user != "":
		return r.reportUserSecret(user, path, limit, offset, cursor)
	case group != "":
		return r.reportGroupSecret(group, path, limit, offset)
	case role != "":
		return r.reportRoleSecret(role, path, limit, offset)
		//read sign in user record
	default:
		return r.reportSignInUserSecret(path, limit, offset, cursor)
	}
}

func (r report) handleGroupReport(args []string) int {
	user := viper.GetString(cst.NounUser)
	limit := viper.GetInt(cst.Limit)
	offset := viper.GetInt(cst.OffSet)

	if r.outClient == nil {
		r.outClient = format.NewDefaultOutClient()
	}

	switch {
	case user != "":
		return r.reportUserGroup(user, limit, offset)

	default:
		return r.reportSignInGroup(limit, offset)
	}
}

func (r report) reportRoleSecret(role, path string, limit, offset int) int {

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
		data, err = r.graphClient.DoRequest(uri, &query, map[string]interface{}{
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

		data, err = r.graphClient.DoRequest(uri, &query, map[string]interface{}{
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

			data, err = r.graphClient.DoRequest(uri, &query, variables)

		} else {
			var query struct {
				Role struct {
					GQRole
					Secrets Secrets `graphql:"secrets(filter:{limit:$limit, secretType: ALL, offset:$offset, path:$path})"`
				} `graphql:"role(name: $role)" json:"data"`
			}

			data, err = r.graphClient.DoRequest(uri, &query, variables)
		}
	}

	r.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (r report) reportGroupSecret(group, path string, limit, offset int) int {

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
		data, err = r.graphClient.DoRequest(uri, &query, map[string]interface{}{
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

			data, err = r.graphClient.DoRequest(uri, &query, variables)

		} else {
			var query struct {
				Group struct {
					GQGroup
					Secrets Secrets `graphql:"secrets(filter:{limit:$limit, secretType: ALL, offset:$offset, path:$path})"`
				} `graphql:"group(groupName: $group)" json:"data"`
			}

			data, err = r.graphClient.DoRequest(uri, &query, variables)
		}
	}

	r.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (r report) reportUserSecret(user, path string, limit, offset int, cursor string) int {
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
		data, err = r.graphClient.DoRequest(uri, &query, map[string]interface{}{
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

		data, err = r.graphClient.DoRequest(uri, &query, map[string]interface{}{
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

			data, err = r.graphClient.DoRequest(uri, &query, variables)

		} else {
			var query struct {
				User struct {
					GQUser
					Secrets Secrets `graphql:"secrets(filter:{limit:$limit, secretType: ALL, offset:$offset, path:$path})"`
				} `graphql:"user(name: $user)" json:"data"`
			}

			data, err = r.graphClient.DoRequest(uri, &query, variables)
		}
	}

	r.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (r report) reportSignInUserSecret(path string, limit, offset int, cursor string) int {
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
		data, err = r.graphClient.DoRequest(uri, &query, map[string]interface{}{
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

		data, err = r.graphClient.DoRequest(uri, &query, map[string]interface{}{
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

			data, err = r.graphClient.DoRequest(uri, &query, variables)

		} else {
			var query struct {
				Me struct {
					GQUser
					Secrets Secrets `graphql:"secrets(filter:{limit:$limit, secretType: ALL, offset:$offset, path:$path})"`
				}
			}

			data, err = r.graphClient.DoRequest(uri, &query, variables)
		}
	}

	r.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (r report) reportSignInGroup(limit, offset int) int {
	var query struct {
		Me struct {
			GQUser
			MemberOf MemberOf `graphql:"memberOf(limit:$limit, offset:$offset)"`
		} `json:"data"`
	}

	data, err := r.graphClient.DoRequest(
		paths.CreateURI("report/me", nil), &query, map[string]interface{}{
			"limit":  graphql.Int(limit),
			"offset": graphql.Int(offset),
		})
	r.outClient.WriteResponse(data, nil)
	return utils.GetExecStatus(err)
}

func (r report) reportUserGroup(user string, limit, offset int) int {
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
	data, err := r.graphClient.DoRequest(uri, &query, variables)
	r.outClient.WriteResponse(data, nil)
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
	Groups []struct {
		Name  graphql.String
		Since graphql.String
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
