package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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

	"github.com/spf13/viper"
	"github.com/thycotic-rd/cli"
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

func GetOpDataGroupWrappers(targetEntity string) cli.PredictorWrappers {
	return cli.PredictorWrappers{
		preds.LongFlag(cst.Data):          cli.PredictorWrapper{preds.NewPrefixFilePredictor("*"), preds.NewFlagValue(preds.Params{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("Group name and members to add to or delete from the %s. Prefix with '@' to denote filepath (subsumes other arguments)", targetEntity)}), false},
		preds.LongFlag(cst.DataGroupName): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.DataGroupName, Usage: fmt.Sprintf("%s of %s (required)", strings.Title(cst.DataName), cst.NounGroup)}), false},
		preds.LongFlag(cst.Members):       cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Members, Usage: "Group members (comma-separated, optional)"}), false},
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
	• group %[1]s --group-name %[6]s --members user1,user2
	• group %[1]s --data %[4]s
	• group %[1]s --data %[5]s
				`, cst.Create, cst.NounGroup, cst.ProductName, cst.ExampleGroupCreate, cst.ExampleDataPath, cst.ExampleGroup),
		FlagsPredictor: GetOpDataGroupWrappers(cst.NounGroup),
		MinNumberArgs:  0,
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
	• group %[1]s --group-name %[6]s --members user3,user4
	• group %[1]s --group-name %[6]s --data %[4]s
	• group %[1]s --group-name %[6]s --data %[5]s
				`, cst.AddMember, cst.NounGroup, cst.ProductName, cst.ExampleGroupAddMembers, cst.ExampleDataPath, cst.ExampleGroup),
		FlagsPredictor: GetOpDataGroupWrappers(cst.NounGroup),
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
	• group %[1]s --group-name %[6]s --members member1,member2
	• group %[1]s --group-name %[6]s --data %[4]s
	• group %[1]s --group-name %[6]s --data %[5]s
				`, cst.DeleteMember, cst.NounGroup, cst.ProductName, cst.ExampleGroupAddMembers, cst.ExampleDataPath, cst.ExampleGroup),
		FlagsPredictor: GetOpDataGroupWrappers(cst.NounGroup),
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
			preds.LongFlag(cst.Cursor): cli.PredictorWrapper{complete.PredictAnything, preds.NewFlagValue(preds.Params{Name: cst.Cursor, Usage: constants.CursorHelpMessage}), false},
		},
		MinNumberArgs: 1,
	})
}
func (g Group) handleGroupReadCmd(args []string) int {
	var err *errors.ApiError
	var data []byte
	groupName := viper.GetString(cst.DataGroupName)
	if groupName == "" && len(args) > 0 {
		groupName = args[0]
	}
	if groupName == "" {
		err = errors.NewS("error: must specify " + cst.DataGroupName)
	} else {
		uri := paths.CreateResourceURI(cst.NounGroup, paths.ProcessPath(groupName), "", true, nil, true)
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
	if OnlyGlobalArgs(args) {
		return g.handleGroupWorkflow(args)
	}
	var err *errors.ApiError
	var resp []byte
	outClient := g.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	uri := paths.CreateResourceURI(cst.NounGroup, "", "", true, nil, true)
	data := viper.GetString(cst.Data)
	if data == "" {
		groupName := viper.GetString(cst.DataGroupName)
		if groupName == "" {
			err := errors.NewS("error: must specify " + cst.DataGroupName)
			outClient.WriteResponse(nil, err)
			return utils.GetExecStatus(err)
		}
		members := viper.GetString(cst.Members)
		params := make(map[string]interface{})
		params["groupName"] = groupName
		// For backward API compatibility, on group create, members are passed in as "members".
		params["members"] = utils.StringToSlice(members)
		resp, err = g.request.DoRequest("POST", uri, &params)
		outClient.WriteResponse(resp, err)
		return utils.GetExecStatus(err)
	}

	params := make(map[string]interface{})
	if dataErr := json.Unmarshal([]byte(data), &params); dataErr != nil {
		g.outClient.WriteResponse(resp, errors.New(dataErr))
		return utils.GetExecStatus(dataErr)
	}
	if _, ok := params["members"]; !ok {
		params["members"] = params["memberNames"]
		delete(params, "memberNames")
	}
	resp, err = g.request.DoRequest("POST", uri, &params)
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
		uri := paths.CreateResourceURI(cst.NounGroup, paths.ProcessPath(groupName), "", true, query, true)
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
		uri := paths.CreateResourceURI(cst.NounGroup, paths.ProcessPath(groupName), "/restore", true, nil, true)
		data, err = g.request.DoRequest("PUT", uri, nil)
	}

	g.outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (g Group) handleAddMembersCmd(args []string) int {
	var err *errors.ApiError
	var resp []byte
	groupName := viper.GetString(cst.DataGroupName)
	if g.outClient == nil {
		g.outClient = format.NewDefaultOutClient()
	}
	if groupName == "" {
		g.outClient.WriteResponse(nil, errors.NewS("--group-name required"))
		return utils.GetExecStatus(err)
	}
	uri := paths.CreateResourceURI(cst.NounGroup, paths.ProcessPath(groupName), "/members", true, nil, true)
	members := viper.GetString(cst.Members)
	data := viper.GetString(cst.Data)
	if data == "" {
		if members == "" {
			err = errors.NewS("error: must specify group name and members")
			g.outClient.WriteResponse(resp, err)
			return utils.GetExecStatus(err)
		}
		params := make(map[string]interface{})
		params["memberNames"] = utils.StringToSlice(members)
		// For backward API compatibility, on adding members, members are passed in as "memberNames".
		resp, err = g.request.DoRequest("POST", uri, &params)
	} else {
		params := make(map[string]interface{})
		if dataErr := json.Unmarshal([]byte(data), &params); dataErr != nil {
			g.outClient.WriteResponse(resp, errors.New(dataErr))
			return utils.GetExecStatus(dataErr)
		}
		if _, ok := params["memberNames"]; !ok {
			params["memberNames"] = params["members"]
			delete(params, "members")
		}
		resp, err = g.request.DoRequest("POST", uri, &params)
	}
	g.outClient.WriteResponse(resp, err)
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
		uri := paths.CreateResourceURI(cst.NounUser, userData, "/groups", true, nil, true)
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
	groupName := viper.GetString(cst.DataGroupName)

	uri := paths.CreateResourceURI(cst.NounGroup, paths.ProcessPath(groupName), "/members", true, nil, true)
	data := viper.GetString(cst.Data)
	if g.outClient == nil {
		g.outClient = format.NewDefaultOutClient()
	}
	if data != "" {
		params := make(map[string]interface{})
		if dataErr := json.Unmarshal([]byte(data), &params); dataErr != nil {
			g.outClient.WriteResponse(resp, errors.New(dataErr))
			return utils.GetExecStatus(dataErr)
		}
		if _, ok := params["memberNames"]; !ok {
			params["memberNames"] = params["members"]
			delete(params, "members")
		}
		resp, err = g.request.DoRequest("DELETE", uri, &params)
	} else {
		members := viper.GetString(cst.Members)
		if groupName == "" || members == "" {
			err = errors.NewS("error: must specify group name and members")
			g.outClient.WriteResponse(resp, err)
			return utils.GetExecStatus(err)
		}
		params := make(map[string]interface{})
		params["memberNames"] = utils.StringToSlice(members)
		// For backward API compatibility, on deleting members, members are passed in as "memberNames".
		resp, err = g.request.DoRequest("DELETE", uri, &params)
	}
	g.outClient.WriteResponse(resp, err)
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
		uri := paths.CreateResourceURI(cst.NounGroup, "", "", false, queryParams, true)
		data, err = g.request.DoRequest("GET", uri, nil)
	}
	outClient := g.outClient
	if outClient == nil {
		outClient = format.NewDefaultOutClient()
	}
	outClient.WriteResponse(data, err)
	return utils.GetExecStatus(err)
}

func (g Group) handleGroupWorkflow(args []string) int {
	ui := &PasswordUi{
		cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
	}
	if g.outClient == nil {
		g.outClient = format.NewDefaultOutClient()
	}
	params := make(map[string]interface{})
	if resp, err := getStringAndValidate(ui, "Group name:", false, nil, false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		params["groupName"] = resp
	}

	if resp, err := getStringAndValidate(ui, "Members (comma-separated):", true, nil, false, false); err != nil {
		ui.Error(err.Error())
		return 1
	} else {
		// For backward API compatibility, on group create, members are passed in as "members".
		params["members"] = utils.StringToSlice(resp)
	}

	uri := paths.CreateResourceURI(cst.NounGroup, "", "", true, nil, true)
	resp, err := g.request.DoRequest("POST", uri, &params)
	g.outClient.WriteResponse(resp, err)
	return utils.GetExecStatus(err)
}
