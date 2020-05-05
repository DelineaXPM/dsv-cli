package cmd

import (
	"fmt"
	"strconv"
	"strings"

	cst "thy/constants"
	"thy/errors"
	"thy/format"
	preds "thy/predictors"
	"thy/requests"
	"thy/utils"

	"github.com/posener/complete"

	"github.com/thycotic-rd/cli"
	"github.com/thycotic-rd/viper"
)

type Group struct {
	request   requests.Client
	outClient format.OutClient
}

func GetNoDataOpGroupWrappers() cli.PredictorWrappers {
	return cli.PredictorWrappers{
		preds.LongFlag(cst.DataGroupName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataGroupName, Usage: fmt.Sprintf("%s of %s (required)", strings.Title(cst.DataName), cst.NounGroup)}), false},
	}
}
func GetNoDataOpGroupUserWrappers() cli.PredictorWrappers {
	return cli.PredictorWrappers{
		preds.LongFlag(cst.DataUsername): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataUsername, Usage: fmt.Sprintf("%s of %s (required)", strings.Title(cst.DataUsername), cst.NounUser)}), false},
	}
}
func GetDataOpGroupWrappers(targetEntity string) cli.PredictorWrappers {
	return cli.PredictorWrappers{
		preds.LongFlag(cst.DataGroupName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataGroupName, Usage: fmt.Sprintf("%s of %s (required)", strings.Title(cst.DataName), cst.NounGroup)}), false},
		preds.LongFlag(cst.Data):          cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("Members to add to the %s. Prefix with '@' to denote filepath (required)", targetEntity)}), false},
	}
}

func GetDataOpGroupDeleteWrappers(targetEntity string) cli.PredictorWrappers {
	return cli.PredictorWrappers{
		preds.LongFlag(cst.DataGroupName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataGroupName, Usage: fmt.Sprintf("%s of %s (required)", strings.Title(cst.DataName), cst.NounGroup)}), false},
		preds.LongFlag(cst.Data):          cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("Members to delete from the %s. Prefix with '@' to denote filepath (required)", targetEntity)}), false},
	}
}
func GetOpDataGroupWrappers(targetEntity string) cli.PredictorWrappers {
	return cli.PredictorWrappers{
		preds.LongFlag(cst.Data):          cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("Members to add to the %s. Prefix with '@' to denote filepath (required)", targetEntity)}), false},
		preds.LongFlag(cst.DataGroupName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataGroupName, Usage: fmt.Sprintf("%s of %s (required)", strings.Title(cst.DataName), cst.NounGroup)}), false},
	}
}
func GetGroupCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounGroup},
		RunFunc: func(args []string) int {
			groupData := viper.GetString(cst.DataGroupName)
			if groupData == "" && len(args) > 0 {
				groupData = args[0]
			}
			if groupData == "" {
				return cli.RunResultHelp
			}
			return Group{requests.NewHttpClient(), nil}.handleGroupReadCmd(args)
		},
		SynopsisText: "group (<group-name> | --group-name)",
		HelpText: fmt.Sprintf(`Execute an action on a %s from %s

Usage:
   • group %[3]s
   • group --group-name %[3]s
		`, cst.NounGroup, cst.ProductName, cst.ExampleGroup),
		FlagsPredictor: GetNoDataOpGroupWrappers(),
		MinNumberArgs:  1,
	})
}

func GetGroupReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path: []string{cst.NounGroup},
		RunFunc: func(args []string) int {
			groupData := viper.GetString(cst.DataGroupName)
			if groupData == "" && len(args) > 0 {
				groupData = args[0]
			}
			if groupData == "" {
				return cli.RunResultHelp
			}
			return Group{requests.NewHttpClient(), nil}.handleGroupReadCmd(args)
		},
		SynopsisText: "read <group-name> | --group-name)",
		HelpText: fmt.Sprintf(`Get %[2]s details

Usage:
   • group %[1]s %[4]s
   • group %[1]s --group-name %[4]s
		`, cst.Read, cst.NounGroup, cst.ProductName, cst.ExampleGroup),
		FlagsPredictor: GetNoDataOpGroupWrappers(),
		MinNumberArgs:  1,
	})
}

func GetGroupCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounGroup, cst.Create},
		RunFunc:      Group{requests.NewHttpClient(), nil}.handleCreateCmd,
		SynopsisText: fmt.Sprintf("%s (<group-name> | --group-name)", cst.Create),
		HelpText: fmt.Sprintf(`Create a %[2]s in %[3]s

Usage:
	• group %[1]s --group-name %[6]s
	• group %[1]s --data %[4]s
	• group %[1]s --data %[5]s
				`, cst.Create, cst.NounGroup, cst.ProductName, cst.ExampleGroupCreate, cst.ExampleDataPath, cst.ExampleGroup),
		FlagsPredictor: GetOpDataGroupWrappers(cst.NounGroup),
		MinNumberArgs:  2,
	})
}

func GetGroupDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounGroup, cst.Delete},
		RunFunc:      Group{requests.NewHttpClient(), nil}.handleGroupDeleteCmd,
		SynopsisText: fmt.Sprintf("%s (<group-name> | --group-name)", cst.Delete),
		HelpText: fmt.Sprintf(`Delete a %[2]s from %[3]s

Usage:
   • group %[1]s %[4]s
   • group %[1]s --group-name %[4]s --force
		`, cst.Delete, cst.NounGroup, cst.ProductName, cst.ExampleGroup),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.DataGroupName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataGroupName, Usage: fmt.Sprintf("%s of %s (required)", strings.Title(cst.DataName), cst.NounGroup)}), false},
			preds.LongFlag(cst.Force):         cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Force, Shorthand: "", Usage: fmt.Sprintf("Immediately delete %s", cst.NounGroup), Global: false, ValueType: "bool"}), false},
		},
		MinNumberArgs: 1,
	})
}

func GetGroupRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounGroup, cst.Restore},
		RunFunc:      Group{requests.NewHttpClient(), nil}.handleGroupRestoreCmd,
		SynopsisText: fmt.Sprintf("%s (<group-name> | --group-name)", cst.Restore),
		HelpText: fmt.Sprintf(`Restore a deleted %[2]s in %[3]s

Usage:
   • group %[1]s %[4]s
   • group %[1]s --group-name %[4]s
		`, cst.Restore, cst.NounGroup, cst.ProductName, cst.ExampleGroup),
		FlagsPredictor: GetNoDataOpGroupWrappers(),
		MinNumberArgs:  1,
	})
}

func GetAddMembersCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounGroup, cst.AddMember},
		RunFunc:      Group{requests.NewHttpClient(), nil}.handleAddMembersCmd,
		SynopsisText: fmt.Sprintf("%s (<data> | (--data|-d))", cst.AddMember),
		HelpText: fmt.Sprintf(`Add Members to a %[2]s in %[3]s

Usage:
	• group %[1]s --group-name %[6]s --data %[4]s
	• group %[1]s --group-name %[6]s --data %[5]s
				`, cst.AddMember, cst.NounGroup, cst.ProductName, cst.ExampleGroupAddMembers, cst.ExampleDataPath, cst.ExampleGroup),
		FlagsPredictor: GetDataOpGroupWrappers(cst.NounGroup),
		MinNumberArgs:  2,
	})
}

func GetDeleteMembersCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounGroup, cst.DeleteMember},
		RunFunc:      Group{requests.NewHttpClient(), nil}.handleDeleteMemberCmd,
		SynopsisText: fmt.Sprintf("%s (<data> | (--data|-d))", cst.DeleteMember),
		HelpText: fmt.Sprintf(`Delete members in a %[2]s in %[3]s

Usage:
	• group %[1]s  --group-name %[6]s --data %[4]s
	• group %[1]s  --group-name %[6]s --data %[5]s
				`, cst.DeleteMember, cst.NounGroup, cst.ProductName, cst.ExampleGroupAddMembers, cst.ExampleDataPath, cst.ExampleGroup),
		FlagsPredictor: GetDataOpGroupDeleteWrappers(cst.NounGroup),
		MinNumberArgs:  2,
	})
}

func GetMemberGroupsCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUser},
		RunFunc:      Group{requests.NewHttpClient(), nil}.handleUsersGroupReadCmd,
		SynopsisText: fmt.Sprintf("(<username> | --username)"),
		HelpText: fmt.Sprintf(`Read a %[2]s's %[1]ss from %[3]s

Usage:
   • user %[1]ss --username %[4]s
		`, cst.NounGroup, cst.NounUser, cst.ProductName, cst.ExampleUser),
		FlagsPredictor: GetNoDataOpGroupUserWrappers(),
		MinNumberArgs:  1,
	})
}
func GetGroupSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounGroup, cst.Search},
		RunFunc:      Group{requests.NewHttpClient(), nil}.handleGroupSearchCmd,
		SynopsisText: fmt.Sprintf("%s (<query> | --query)", cst.Search),
		HelpText: fmt.Sprintf(`Search for a %[2]s from %[3]s

		Usage:
		• group %[1]s %[4]s
		• group %[1]s --query %[4]s
				`, cst.Search, cst.NounGroup, cst.ProductName, cst.ExampleUserSearch),
		FlagsPredictor: cli.PredictorWrappers{
			preds.LongFlag(cst.Query):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Query, Shorthand: "q", Usage: fmt.Sprintf("%s of %ss to fetch (required)", strings.Title(cst.Query), cst.NounGroup)}), false},
			preds.LongFlag(cst.Limit):  cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Limit, Shorthand: "l", Usage: fmt.Sprint("Maximum number of results per cursor (optional)")}), false},
			preds.LongFlag(cst.Cursor): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Cursor, Usage: fmt.Sprint("Next cursor for additional results (optional)")}), false},
		},
		MinNumberArgs: 1,
	})
}
func (g Group) handleGroupReadCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	groupData := viper.GetString(cst.DataGroupName)
	if groupData == "" && len(args) > 0 {
		groupData = args[0]
	}
	if groupData == "" {
		err = errors.NewS("error: must specify " + cst.DataGroupName)
	} else {
		uri := utils.CreateResourceURI(cst.NounGroup, groupData, "", true, nil, true)
		data, err = g.request.DoRequest("GET", uri, nil)
	}

	outClient := g.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}

	outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (g Group) handleCreateCmd(args []string) int {
	var err *errors.ApiError
	var resp []byte
	outClient := g.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	uri := utils.CreateResourceURI(cst.NounGroup, "", "", true, nil, true)
	data := viper.GetString(cst.Data)
	if data == "" && len(args) > 1 {
		groupName := viper.GetString(cst.DataGroupName)
		if groupName == "" {
			err := errors.NewS("error: must specify " + cst.DataGroupName)
			outClient.WriteResponse(nil, err)
			return utils.GetExecStatus(err)
		}

		mapData := map[string]string{"groupName": groupName}
		resp, err = g.request.DoRequest("POST", uri, &mapData)
		outClient.WriteResponse(resp, err)
		return utils.GetExecStatus(err)
	}

	resp, err = g.request.DoRequest("POST", uri, []byte(data))
	outClient.WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func (g Group) handleGroupDeleteCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	force := viper.GetBool(cst.Force)
	groupName := viper.GetString(cst.DataGroupName)
	if groupName == "" && len(args) > 0 {
		groupName = args[0]
	}
	if groupName == "" {
		err = errors.NewS("error: must specify " + cst.DataGroupName)
	} else {
		query := map[string]string{"force": strconv.FormatBool(force)}
		uri := utils.CreateResourceURI(cst.NounGroup, groupName, "", true, query, true)
		data, err = g.request.DoRequest("DELETE", uri, nil)
	}

	if g.outClient == nil {
		g.outClient = format.NewDefaultOutClient()
	}

	g.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (g Group) handleGroupRestoreCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	if g.outClient == nil {
		g.outClient = format.NewDefaultOutClient()
	}
	groupName := viper.GetString(cst.DataGroupName)
	if groupName == "" && len(args) > 0 {
		groupName = args[0]
	}

	if groupName == "" {
		err = errors.NewS("error: must specify " + cst.DataGroupName)
	} else {
		uri, err := utils.GetResourceURIFromResourcePath(cst.NounGroup, groupName, "", "", true, nil)
		if err == nil {
			uri += "/restore"
			data, err = g.request.DoRequest("PUT", uri, nil)
		}
	}

	g.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (g Group) handleAddMembersCmd(args []string) int {
	var err *errors.ApiError
	var resp []byte
	groupData := viper.GetString(cst.DataGroupName)
	outClient := g.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	if groupData == "" {
		outClient.WriteResponse(nil, errors.NewS("--group-name required"))
		return utils.GetExecStatus(err)
	}
	uri := utils.CreateResourceURI(cst.NounGroup, groupData, "/members", true, nil, true)

	data := viper.GetString(cst.Data)
	if data == "" && len(args) > 1 {
		data = args[0]
	}
	resp, err = g.request.DoRequest("POST", uri, []byte(data))
	outClient.WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func (g Group) handleUsersGroupReadCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	userData := viper.GetString(cst.DataUsername)
	if userData == "" && len(args) > 0 {
		userData = args[0]
	}
	if userData == "" {
		err = errors.NewS("error: must specify " + cst.DataUsername)
	} else {
		version := viper.GetString(cst.Version)
		if strings.TrimSpace(version) != "" {
			userData = fmt.Sprint(userData, "/", cst.Version, "/", version)
		}
		uri := utils.CreateResourceURI(cst.NounUser, userData, "/groups", true, nil, true)
		data, err = g.request.DoRequest("GET", uri, nil)
	}

	outClient := g.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}

	outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (g Group) handleDeleteMemberCmd(args []string) int {
	var err *errors.ApiError
	var resp []byte
	groupData := viper.GetString(cst.DataGroupName)

	uri := utils.CreateResourceURI(cst.NounGroup, groupData, "/members", true, nil, true)
	data := viper.GetString(cst.Data)
	if data == "" && len(args) > 1 {
		data = args[0]
	}

	resp, err = g.request.DoRequest("DELETE", uri, []byte(data))
	outClient := g.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}

	outClient.WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}

func (g Group) handleGroupSearchCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	query := viper.GetString(cst.Query)
	limit := viper.GetString(cst.Limit)
	cursor := viper.GetString(cst.Cursor)
	if query == "" && len(args) > 0 {
		query = args[0]
	}
	if query == "" {
		err = errors.NewS("error: must specify " + cst.Query)
	} else {
		queryParams := map[string]string{
			cst.SearchKey: query,
			cst.Limit:     limit,
			cst.Cursor:    cursor,
		}
		uri := utils.CreateResourceURI(cst.NounGroup, "", "", false, queryParams, true)
		data, err = g.request.DoRequest("GET", uri, nil)
	}
	outClient := g.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}
