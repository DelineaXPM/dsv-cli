package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	cst "thy/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/internal/predictor"
	"github.com/DelineaXPM/dsv-cli/paths"
	"github.com/DelineaXPM/dsv-cli/utils"
	"github.com/DelineaXPM/dsv-cli/vaultcli"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/cli"
	"github.com/spf13/viper"
)

func GetGroupCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounGroup},
		SynopsisText: "Manage groups",
		HelpText: fmt.Sprintf(`Execute an action on a %s from %s

Usage:
   • group %[3]s
   • group --group-name %[3]s
`, cst.NounGroup, cst.ProductName, cst.ExampleGroup),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataGroupName, Usage: fmt.Sprintf("%s of %s (required)", strings.Title(cst.DataName), cst.NounGroup)},
		},
		MinNumberArgs: 1,
		RunFunc: func(vcli vaultcli.CLI, args []string) int {
			groupData := viper.GetString(cst.DataGroupName)
			if groupData == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
				groupData = args[0]
			}
			if groupData == "" {
				return cli.RunResultHelp
			}
			return handleGroupReadCmd(vcli, args)
		},
	})
}

func GetGroupReadCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounGroup},
		SynopsisText: "read <group-name> | --group-name)",
		HelpText: fmt.Sprintf(`Get %[2]s details

Usage:
   • group %[1]s %[4]s
   • group %[1]s --group-name %[4]s
`, cst.Read, cst.NounGroup, cst.ProductName, cst.ExampleGroup),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataGroupName, Usage: fmt.Sprintf("%s of %s (required)", strings.Title(cst.DataName), cst.NounGroup)},
		},
		MinNumberArgs: 1,
		RunFunc: func(vcli vaultcli.CLI, args []string) int {
			groupData := viper.GetString(cst.DataGroupName)
			if groupData == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
				groupData = args[0]
			}
			if groupData == "" {
				return cli.RunResultHelp
			}
			return handleGroupReadCmd(vcli, args)
		},
	})
}

func GetGroupCreateCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounGroup, cst.Create},
		SynopsisText: fmt.Sprintf("%s (<group-name> | --group-name)", cst.Create),
		HelpText: fmt.Sprintf(`Create a %[2]s in %[3]s

Usage:
   • group %[1]s --group-name %[6]s --members user1,user2
   • group %[1]s --data %[4]s
   • group %[1]s --data %[5]s
`, cst.Create, cst.NounGroup, cst.ProductName, cst.ExampleGroupCreate, cst.ExampleDataPath, cst.ExampleGroup),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("Group name and members to add to or delete from the %s. Prefix with '@' to denote filepath (subsumes other arguments)", cst.NounGroup), Predictor: predictor.NewPrefixFilePredictor("*")},
			{Name: cst.DataGroupName, Usage: fmt.Sprintf("%s of %s (required)", strings.Title(cst.DataName), cst.NounGroup)},
			{Name: cst.Members, Usage: "Group members (comma-separated, optional)"},
		},
		RunFunc:    handleGroupCreateCmd,
		WizardFunc: handleGroupCreateWizard,
	})
}

func GetGroupDeleteCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounGroup, cst.Delete},
		SynopsisText: fmt.Sprintf("%s (<group-name> | --group-name)", cst.Delete),
		HelpText: fmt.Sprintf(`Delete a %[2]s from %[3]s

Usage:
   • group %[1]s %[4]s
   • group %[1]s --group-name %[4]s --force
`, cst.Delete, cst.NounGroup, cst.ProductName, cst.ExampleGroup),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataGroupName, Usage: fmt.Sprintf("%s of %s (required)", strings.Title(cst.DataName), cst.NounGroup)},
			{Name: cst.Force, Usage: fmt.Sprintf("Immediately delete %s", cst.NounGroup), ValueType: "bool"},
		},
		MinNumberArgs: 1,
		RunFunc:       handleGroupDeleteCmd,
	})
}

func GetGroupRestoreCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounGroup, cst.Restore},
		SynopsisText: fmt.Sprintf("%s (<group-name> | --group-name)", cst.Restore),
		HelpText: fmt.Sprintf(`Restore a deleted %[2]s in %[3]s

Usage:
   • group %[1]s %[4]s
   • group %[1]s --group-name %[4]s
`, cst.Restore, cst.NounGroup, cst.ProductName, cst.ExampleGroup),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataGroupName, Usage: fmt.Sprintf("%s of %s (required)", strings.Title(cst.DataName), cst.NounGroup)},
		},
		MinNumberArgs: 1,
		RunFunc:       handleGroupRestoreCmd,
	})
}

func GetAddMembersCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounGroup, cst.AddMember},
		SynopsisText: fmt.Sprintf("%s (<data> | (--data|-d))", cst.AddMember),
		HelpText: fmt.Sprintf(`Add Members to a %[2]s in %[3]s

Usage:
   • group %[1]s --group-name %[6]s --members user3,user4
   • group %[1]s --group-name %[6]s --data %[4]s
   • group %[1]s --group-name %[6]s --data %[5]s
`, cst.AddMember, cst.NounGroup, cst.ProductName, cst.ExampleGroupAddMembers, cst.ExampleDataPath, cst.ExampleGroup),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Data, Shorthand: "d", Usage: "Group name and members to add to or delete from the group. Prefix with '@' to denote filepath (subsumes other arguments)", Predictor: predictor.NewPrefixFilePredictor("*")},
			{Name: cst.DataGroupName, Usage: fmt.Sprintf("%s of %s (required)", strings.Title(cst.DataName), cst.NounGroup)},
			{Name: cst.Members, Usage: "Group members (comma-separated, optional)"},
		},
		MinNumberArgs: 2,
		RunFunc:       handleAddMembersCmd,
	})
}

func GetDeleteMembersCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounGroup, cst.DeleteMember},
		SynopsisText: fmt.Sprintf("%s (<data> | (--data|-d))", cst.DeleteMember),
		HelpText: fmt.Sprintf(`Delete members in a %[2]s in %[3]s

Usage:
   • group %[1]s --group-name %[6]s --members member1,member2
   • group %[1]s --group-name %[6]s --data %[4]s
   • group %[1]s --group-name %[6]s --data %[5]s
`, cst.DeleteMember, cst.NounGroup, cst.ProductName, cst.ExampleGroupAddMembers, cst.ExampleDataPath, cst.ExampleGroup),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Data, Shorthand: "d", Usage: fmt.Sprintf("Group name and members to add to or delete from the %s. Prefix with '@' to denote filepath (subsumes other arguments)", cst.NounGroup), Predictor: predictor.NewPrefixFilePredictor("*")},
			{Name: cst.DataGroupName, Usage: fmt.Sprintf("%s of %s (required)", strings.Title(cst.DataName), cst.NounGroup)},
			{Name: cst.Members, Usage: "Group members (comma-separated, optional)"},
		},
		MinNumberArgs: 2,
		RunFunc:       handleDeleteMembersCmd,
	})
}

func GetMemberGroupsCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounUser},
		SynopsisText: "groups (<username> | --username)",
		HelpText: fmt.Sprintf(`Read a %[2]s's %[1]ss from %[3]s

Usage:
   • user %[1]ss --username %[4]s
`, cst.NounGroup, cst.NounUser, cst.ProductName, cst.ExampleUser),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.DataUsername, Usage: fmt.Sprintf("%s of %s (required)", strings.Title(cst.DataUsername), cst.NounUser)},
		},
		MinNumberArgs: 1,
		RunFunc:       handleUsersGroupReadCmd,
	})
}

func GetGroupSearchCmd() (cli.Command, error) {
	return NewCommand(CommandArgs{
		Path:         []string{cst.NounGroup, cst.Search},
		SynopsisText: fmt.Sprintf("%s (<query> | --query)", cst.Search),
		HelpText: fmt.Sprintf(`Search for a %[2]s from %[3]s

Usage:
   • group %[1]s %[4]s
   • group %[1]s --query %[4]s
`, cst.Search, cst.NounGroup, cst.ProductName, cst.ExampleUserSearch),
		FlagsPredictor: []*predictor.Params{
			{Name: cst.Query, Shorthand: "q", Usage: fmt.Sprintf("%s of %ss to fetch (optional)", strings.Title(cst.Query), cst.NounGroup)},
			{Name: cst.Limit, Shorthand: "l", Usage: cst.LimitHelpMessage},
			{Name: cst.Cursor, Usage: cst.CursorHelpMessage},
		},
		RunFunc: handleGroupSearchCmd,
	})
}

func handleGroupReadCmd(vcli vaultcli.CLI, args []string) int {
	groupName := viper.GetString(cst.DataGroupName)
	if groupName == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		groupName = args[0]
	}

	if groupName == "" {
		err := errors.NewS("error: must specify " + cst.DataGroupName)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, apiError := groupRead(vcli, paths.ProcessResource(groupName))
	vcli.Out().WriteResponse(data, apiError)
	return utils.GetExecStatus(apiError)
}

func handleGroupCreateCmd(vcli vaultcli.CLI, args []string) int {
	data := viper.GetString(cst.Data)

	var (
		groupName string
		members   []string
	)
	if data == "" {
		groupName = viper.GetString(cst.DataGroupName)
		if groupName == "" {
			err := errors.NewS("error: must specify " + cst.DataGroupName)
			vcli.Out().WriteResponse(nil, err)
			return utils.GetExecStatus(err)
		}
		if err := vaultcli.ValidateName(groupName); err != nil {
			vcli.Out().FailF("error: %s %q is invalid: %v", cst.DataGroupName, groupName, err)
			return utils.GetExecStatus(err)
		}
		membersString := viper.GetString(cst.Members)
		members = utils.StringToSlice(membersString)
	} else {
		dataModel := struct {
			GroupName   string   `json:"groupName"`
			Members     []string `json:"members"`
			MemberNames []string `json:"memberNames"`
		}{}
		if dataErr := json.Unmarshal([]byte(data), &dataModel); dataErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(dataErr))
			return utils.GetExecStatus(dataErr)
		}
		if dataModel.GroupName == "" {
			err := errors.NewS("error: missing group name (\"groupName\") field in data")
			vcli.Out().WriteResponse(nil, err)
			return utils.GetExecStatus(err)
		}
		groupName = dataModel.GroupName
		if len(dataModel.Members) > 0 {
			members = dataModel.Members
		} else if len(dataModel.MemberNames) > 0 {
			members = dataModel.MemberNames
		}
	}

	resp, apiError := groupCreate(vcli, groupName, members)
	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

func handleGroupDeleteCmd(vcli vaultcli.CLI, args []string) int {
	force := viper.GetBool(cst.Force)
	groupName := viper.GetString(cst.DataGroupName)
	if groupName == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		groupName = args[0]
	}
	if groupName == "" {
		err := errors.NewS("error: must specify " + cst.DataGroupName)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, apiError := groupDelete(vcli, paths.ProcessResource(groupName), force)
	vcli.Out().WriteResponse(data, apiError)
	return utils.GetExecStatus(apiError)
}

func handleGroupRestoreCmd(vcli vaultcli.CLI, args []string) int {
	groupName := viper.GetString(cst.DataGroupName)
	if groupName == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		groupName = args[0]
	}

	if groupName == "" {
		err := errors.NewS("error: must specify " + cst.DataGroupName)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, apiError := groupRestore(vcli, paths.ProcessResource(groupName))
	vcli.Out().WriteResponse(data, apiError)
	return utils.GetExecStatus(apiError)
}

func handleAddMembersCmd(vcli vaultcli.CLI, args []string) int {
	groupName := viper.GetString(cst.DataGroupName)
	if groupName == "" {
		err := errors.NewS("error: flag --group-name is required")
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data := viper.GetString(cst.Data)

	members := []string{}

	if data == "" {
		membersString := viper.GetString(cst.Members)
		if membersString == "" {
			err := errors.NewS("error: must specify members")
			vcli.Out().WriteResponse(nil, err)
			return utils.GetExecStatus(err)
		}
		members = utils.StringToSlice(membersString)
	} else {
		dataModel := struct {
			MemberNames []string `json:"memberNames"`
			Members     []string `json:"members"`
		}{}
		if dataErr := json.Unmarshal([]byte(data), &dataModel); dataErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(dataErr))
			return utils.GetExecStatus(dataErr)
		}
		if len(dataModel.MemberNames) > 0 {
			members = dataModel.MemberNames
		} else if len(dataModel.Members) > 0 {
			members = dataModel.Members
		}
	}

	if len(members) == 0 {
		err := errors.NewS("error: missing list of members to add")
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	resp, apiErr := groupAddMembers(vcli, paths.ProcessResource(groupName), members)
	vcli.Out().WriteResponse(resp, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleUsersGroupReadCmd(vcli vaultcli.CLI, args []string) int {
	username := viper.GetString(cst.DataUsername)
	if username == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		username = args[0]
	}

	if username == "" {
		err := errors.NewS("error: must specify " + cst.DataUsername)
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data, apiErr := userGroupsRead(vcli, username)
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleDeleteMembersCmd(vcli vaultcli.CLI, args []string) int {
	groupName := viper.GetString(cst.DataGroupName)
	if groupName == "" {
		err := errors.NewS("error: flag --group-name is required")
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	data := viper.GetString(cst.Data)

	members := []string{}

	if data == "" {
		membersString := viper.GetString(cst.Members)
		if membersString == "" {
			err := errors.NewS("error: must specify members")
			vcli.Out().WriteResponse(nil, err)
			return utils.GetExecStatus(err)
		}
		members = utils.StringToSlice(membersString)
	} else {
		dataModel := struct {
			MemberNames []string `json:"memberNames"`
			Members     []string `json:"members"`
		}{}
		if dataErr := json.Unmarshal([]byte(data), &dataModel); dataErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(dataErr))
			return utils.GetExecStatus(dataErr)
		}
		if len(dataModel.MemberNames) > 0 {
			members = dataModel.MemberNames
		} else if len(dataModel.Members) > 0 {
			members = dataModel.Members
		}
	}

	if len(members) == 0 {
		err := errors.NewS("error: missing list of members to delete")
		vcli.Out().WriteResponse(nil, err)
		return utils.GetExecStatus(err)
	}

	resp, apiErr := groupDelMembers(vcli, paths.ProcessResource(groupName), members)
	vcli.Out().WriteResponse(resp, apiErr)
	return utils.GetExecStatus(apiErr)
}

func handleGroupSearchCmd(vcli vaultcli.CLI, args []string) int {
	query := viper.GetString(cst.Query)
	limit := viper.GetString(cst.Limit)
	cursor := viper.GetString(cst.Cursor)

	if query == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		query = args[0]
	}

	data, apiErr := groupSearch(vcli, &groupSearchParams{query: query, limit: limit, cursor: cursor})
	vcli.Out().WriteResponse(data, apiErr)
	return utils.GetExecStatus(apiErr)
}

// Wizards:

func handleGroupCreateWizard(vcli vaultcli.CLI) int {
	var groupName string
	groupNamePrompt := &survey.Input{Message: "Group name:"}
	groupNameValidation := func(ans interface{}) error {
		answer := ans.(string)
		if answer == "" {
			return errors.NewS("A group name is required.")
		}
		if err := vaultcli.ValidateName(answer); err != nil {
			return err
		}
		_, apiError := groupRead(vcli, answer)
		if apiError == nil {
			return errors.NewS("A group with this name already exists.")
		}
		return nil
	}
	survErr := survey.AskOne(groupNamePrompt, &groupName, survey.WithValidator(groupNameValidation))
	if survErr != nil {
		vcli.Out().WriteResponse(nil, errors.New(survErr))
		return utils.GetExecStatus(survErr)
	}

	members := []string{}
	for {
		qs := []*survey.Question{
			{
				Name:   "member",
				Prompt: &survey.Input{Message: "Username to add to the group:"},
				Validate: func(ans interface{}) error {
					answer := strings.TrimSpace(ans.(string))

					// Empty answer is allowed and will be ignored.
					// Group can be created without members.
					if answer == "" {
						return nil
					}
					_, apiError := userRead(vcli, answer)
					if apiError == nil {
						return nil
					}
					httpResp := apiError.HttpResponse()
					if httpResp == nil || httpResp.StatusCode != http.StatusNotFound {
						return nil
					}
					return errors.NewS("A user with this username does not exist.")
				},
				Transform: vaultcli.SurveyTrimSpace,
			},
			{Name: "addMore", Prompt: &survey.Confirm{Message: "Add more?", Default: true}},
		}

		answers := struct {
			Member  string
			AddMore bool
		}{}
		survErr := survey.Ask(qs, &answers)
		if survErr != nil {
			vcli.Out().WriteResponse(nil, errors.New(survErr))
			return utils.GetExecStatus(survErr)
		}
		if answers.Member != "" {
			members = append(members, answers.Member)
		}
		if !answers.AddMore {
			break
		}
	}

	resp, apiError := groupCreate(vcli, groupName, members)
	vcli.Out().WriteResponse(resp, apiError)
	return utils.GetExecStatus(apiError)
}

// API callers:

func groupCreate(vcli vaultcli.CLI, name string, members []string) ([]byte, *errors.ApiError) {
	params := map[string]interface{}{
		"groupName": name,
		"members":   members,
	}
	uri := paths.CreateResourceURI(cst.NounGroups, "", "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, &params)
}

func groupAddMembers(vcli vaultcli.CLI, name string, members []string) ([]byte, *errors.ApiError) {
	params := map[string]interface{}{
		"memberNames": members,
	}
	uri := paths.CreateResourceURI(cst.NounGroups, name, "/members", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPost, uri, &params)
}

func groupDelMembers(vcli vaultcli.CLI, name string, members []string) ([]byte, *errors.ApiError) {
	params := map[string]interface{}{
		"memberNames": members,
	}
	uri := paths.CreateResourceURI(cst.NounGroups, name, "/members", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodDelete, uri, &params)
}

func groupRead(vcli vaultcli.CLI, name string) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounGroups, name, "", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}

func userGroupsRead(vcli vaultcli.CLI, username string) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounUsers, username, "/groups", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}

func groupDelete(vcli vaultcli.CLI, name string, force bool) ([]byte, *errors.ApiError) {
	query := map[string]string{"force": strconv.FormatBool(force)}
	uri := paths.CreateResourceURI(cst.NounGroups, name, "", true, query)
	return vcli.HTTPClient().DoRequest(http.MethodDelete, uri, nil)
}

func groupRestore(vcli vaultcli.CLI, name string) ([]byte, *errors.ApiError) {
	uri := paths.CreateResourceURI(cst.NounGroups, name, "/restore", true, nil)
	return vcli.HTTPClient().DoRequest(http.MethodPut, uri, nil)
}

type groupSearchParams struct {
	query  string
	limit  string
	cursor string
}

func groupSearch(vcli vaultcli.CLI, p *groupSearchParams) ([]byte, *errors.ApiError) {
	queryParams := map[string]string{}
	if p.query != "" {
		queryParams[cst.SearchKey] = p.query
	}
	if p.limit != "" {
		queryParams[cst.Limit] = p.limit
	}
	if p.cursor != "" {
		queryParams[cst.Cursor] = p.cursor
	}
	uri := paths.CreateURI(cst.NounGroups, queryParams)
	return vcli.HTTPClient().DoRequest(http.MethodGet, uri, nil)
}
