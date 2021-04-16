package cmd

import (
	"fmt"
	cst "thy/constants"
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
   • %[1]s  --%[2]s testUser  --%[3]s role1 --%[4]s 5
   • %[1]s --%[2]s 2020-01-01
   `, cst.NounReport, cst.NounUser, cst.NounRole, cst.Limit),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.NounUser): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.NounUser, Usage: "User name (optional)"}), false},
			preds.LongFlag(cst.NounRole): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.NounRole, Usage: "Role name (optional)"}), false},
			preds.LongFlag(cst.Limit):    cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Limit, Shorthand: "l", Usage: "Maximum number of results per cursor (optional)"}), false},
		},
		MinNumberArgs: 0,
	})
}

func (r report) handleSecretReport(args []string) int {
	user := viper.GetString(cst.NounUser)
	group := viper.GetString(cst.NounGroup)
	path := viper.GetString(cst.Path)
	role := viper.GetString(cst.NounRole)
	limit := viper.GetInt(cst.Limit)

	if r.outClient == nil {
		r.outClient = format.NewDefaultOutClient()
	}

	switch {
	case user != "":
		return r.reportUserSecret(limit, user, path)
	case group != "":
		return r.reportGroupSecret(limit, group, path)
	case role != "":
		return r.reportRoleSecret(limit, role, path)
		//read sign in user record
	default:
		return r.reportSignInUserSecret(limit, path)
	}
}

func (r report) handleGroupReport(args []string) int {
	user := viper.GetString(cst.NounUser)
	role := viper.GetString(cst.NounRole)

	if r.outClient == nil {
		r.outClient = format.NewDefaultOutClient()
	}

	switch {
	case user != "":
		return r.reportUserGroup(user)

	case role != "":
		return r.reportRoleGroup(role)

		//read sign in user record
	default:
		return r.reportSignInGroup()
	}
}

func (r report) reportRoleSecret(limit int, role string, path string) int {
	var query struct {
		Role struct {
			Name           graphql.String
			ExternalID     graphql.String
			Provider       graphql.String
			Created        graphql.String
			LastModified   graphql.String
			CreatedBy      graphql.String
			LastModifiedBy graphql.String
			Version        graphql.String
			Description    graphql.String
			Secrets        []struct {
				Actions        []graphql.String
				ID             graphql.String
				Path           graphql.String
				Created        graphql.String
				LastModified   graphql.String
				LastModifiedBy graphql.String
				Version        graphql.String
			} `graphql:"secrets(limit:$limit, path:$path)"`
		} `graphql:"role(name: $role)" json:"data"`
	}

	variables := map[string]interface{}{
		"limit": graphql.Int(limit),
		"role":  graphql.String(role),
		"path":  graphql.String(path),
	}

	uri := paths.CreateURI("report/query", nil)
	data, err := r.graphClient.DoRequest(uri, &query, variables)
	r.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (r report) reportGroupSecret(limit int, group string, path string) int {
	var query struct {
		Group struct {
			Name           graphql.String
			Created        graphql.String
			LastModified   graphql.String
			CreatedBy      graphql.String
			LastModifiedBy graphql.String
			Version        graphql.String
			Secrets        []struct {
				Actions        []graphql.String
				ID             graphql.String
				Path           graphql.String
				Created        graphql.String
				LastModified   graphql.String
				LastModifiedBy graphql.String
				Version        graphql.String
			} `graphql:"secrets(limit:$limit, path:$path)"`
		} `graphql:"group(groupName: $group)" json:"data"`
	}
	variables := map[string]interface{}{
		"limit": graphql.Int(limit),
		"group": graphql.String(group),
		"path":  graphql.String(path),
	}

	uri := paths.CreateURI("report/query", nil)
	data, err := r.graphClient.DoRequest(uri, &query, variables)
	r.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (r report) reportUserSecret(limit int, user string, path string) int {
	var query struct {
		User struct {
			UserName       graphql.String
			Provider       graphql.String
			Created        graphql.String
			LastModified   graphql.String
			CreatedBy      graphql.String
			LastModifiedBy graphql.String
			Version        graphql.String
			Secrets        []struct {
				Actions        []graphql.String
				ID             graphql.String
				Path           graphql.String
				Created        graphql.String
				LastModified   graphql.String
				LastModifiedBy graphql.String
				Version        graphql.String
			} `graphql:"secrets(limit:$limit, path:$path)"`
		} `graphql:"user(name: $user)" json:"data"`
	}

	variables := map[string]interface{}{
		"limit": graphql.Int(limit),
		"user":  graphql.String(user),
		"path":  graphql.String(path),
	}

	uri := paths.CreateURI("report/query", nil)
	data, err := r.graphClient.DoRequest(uri, &query, variables)
	r.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (r report) reportSignInUserSecret(limit int, path string) int {
	var query struct {
		Me struct {
			UserName       graphql.String
			Provider       graphql.String
			Created        graphql.String
			LastModified   graphql.String
			CreatedBy      graphql.String
			LastModifiedBy graphql.String
			Version        graphql.String
			Secrets        []struct {
				Actions        []graphql.String
				ID             graphql.String
				Path           graphql.String
				Created        graphql.String
				LastModified   graphql.String
				LastModifiedBy graphql.String
				Version        graphql.String
			} `graphql:"secrets(limit:$limit, path:$path)"`
			Home []struct {
				ID             graphql.String
				Path           graphql.String
				Created        graphql.String
				LastModified   graphql.String
				LastModifiedBy graphql.String
				Version        graphql.String
			} `graphql:"home(limit:$limit)"`
		} `json:"data"`
	}

	variables := map[string]interface{}{
		"limit": graphql.Int(limit),
		"path":  graphql.String(path),
	}
	uri := paths.CreateURI("report/me", nil)
	data, err := r.graphClient.DoRequest(uri, &query, variables)
	r.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (r report) reportSignInGroup() int {
	var query struct {
		Me struct {
			UserName       graphql.String
			Provider       graphql.String
			Created        graphql.String
			LastModified   graphql.String
			CreatedBy      graphql.String
			LastModifiedBy graphql.String
			Version        graphql.String
			MemberOf       []struct {
				Name  graphql.String
				Since graphql.String
			} `json:"group"`
		} `json:"data"`
	}

	data, err := r.graphClient.DoRequest(paths.CreateURI("report/me", nil), &query, nil)
	r.outClient.WriteResponse(data, nil)
	return utils.GetExecStatus(err)
}

func (r report) reportRoleGroup(role string) int {
	var query struct {
		Role struct {
			Name           graphql.String
			ExternalID     graphql.String
			Provider       graphql.String
			Created        graphql.String
			LastModified   graphql.String
			CreatedBy      graphql.String
			LastModifiedBy graphql.String
			Version        graphql.String
			Description    graphql.String
			MemberOf       []struct {
				Name  graphql.String
				Since graphql.String
			} `json:"group"`
		} `graphql:"role(name: $role)" json:"data"`
	}

	variables := map[string]interface{}{
		"role": graphql.String(role),
	}

	uri := paths.CreateURI("report/query", nil)
	data, err := r.graphClient.DoRequest(uri, &query, variables)
	r.outClient.WriteResponse(data, nil)
	return utils.GetExecStatus(err)
}

func (r report) reportUserGroup(user string) int {
	var query struct {
		User struct {
			UserName       graphql.String
			Provider       graphql.String
			Created        graphql.String
			LastModified   graphql.String
			CreatedBy      graphql.String
			LastModifiedBy graphql.String
			Version        graphql.String
			MemberOf       []struct {
				Name  graphql.String
				Since graphql.String
			} `json:"group"`
		} `graphql:"user(name: $user)" json:"data"`
	}

	variables := map[string]interface{}{
		"user": graphql.String(user),
	}

	uri := paths.CreateURI("report/query", nil)
	data, err := r.graphClient.DoRequest(uri, &query, variables)
	r.outClient.WriteResponse(data, nil)
	return utils.GetExecStatus(err)
}
