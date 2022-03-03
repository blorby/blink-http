package aws_sso

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blinkops/blink-http/implementation/requests"
	"github.com/blinkops/blink-http/plugins/types"
	"github.com/blinkops/blink-sdk/plugin"
	"strconv"
)

const (
	PluginName        = "aws-sso"
	emailParam        = "Email"
	groupNameParam    = "Group Name"
	usersParam        = "Users"
	createUsersParam  = "Create users if they don't exist"
	primaryEmailParam = "Primary Email"
	givenNameParam    = "Given Name"
	familyNameParam   = "Family Name"
	activeParam       = "Active"
)

func addUsersToGroup(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	requestUrl, err := requests.GetRequestUrl(ctx, PluginName)
	if err != nil {
		return nil, err
	}

	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}

	// validate params
	users, groupName, createStr := params[usersParam], params[groupNameParam], params[createUsersParam]
	if users == "" {
		return nil, errors.New("group members not provided")
	}
	if groupName == "" {
		return nil, errors.New("group name not provided")
	}

	// convert create to boolean
	create, err := strconv.ParseBool(createStr)
	if err != nil {
		return nil, errors.New("cannot convert create users if they don't exist to boolean")
	}

	// find the group's ID
	group, err := findGroupByDisplayName(ctx, request, plugin, groupName)
	if err != nil {
		return nil, err
	}
	groupId := group.ID

	// perform operation
	var statuses chan AddUserToGroupResponse
	if create {
		statuses, err = createAndAddUsersToGroup(ctx, request, plugin, requestUrl, groupId, users)
	} else {
		statuses, err = performAddUsersToGroup(ctx, request, plugin, requestUrl, groupId, users)
	}

	// if there's an error, we'll just return the first one
	if err != nil {
		return nil, err
	}

	var response []AddUserToGroupResponse
	for len(statuses) > 0 {
		response = append(response, <-statuses)
	}

	marshalledResponse, err := json.Marshal(response)
	if err != nil {
		return nil, errors.New("failed to marshal json")
	}

	return marshalledResponse, nil
}

func createUser(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	requestUrl, err := requests.GetRequestUrl(ctx, PluginName)
	if err != nil {
		return nil, err
	}

	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}

	email, givenName, familyName, active := params[primaryEmailParam], params[givenNameParam], params[familyNameParam], params[activeParam]
	if email == "" {
		return nil, errors.New("email not provided")
	}
	if givenName == "" {
		return nil, errors.New("given name not provided")
	}
	if familyName == "" {
		return nil, errors.New("family name not provided")
	}
	if active == "" {
		return nil, errors.New("active not provided")
	}

	activeBool, err := strconv.ParseBool(active)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot convert active param (%s) to boolean", active))
	}

	userInfo := UserInfo{
		Email: email,
		Name: UserName{
			GivenName:  givenName,
			FamilyName: familyName,
		},
		Active: activeBool,
	}

	user, err := performCreateUser(ctx, request, plugin, requestUrl, userInfo)
	if err != nil {
		return nil, err
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		return nil, errors.New("unable to marshal json")
	}

	return userJson, nil
}

func deleteUser(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	requestUrl, err := requests.GetRequestUrl(ctx, PluginName)
	if err != nil {
		return nil, err
	}

	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}

	email := params[emailParam]
	if email == "" {
		return nil, errors.New("email not provided")
	}

	user, err := findUserByEmail(ctx, request, plugin, requestUrl, email)
	if err != nil {
		return nil, err
	}

	err = performDeleteUser(ctx, request, plugin, requestUrl, user.ID)
	if err != nil {
		return nil, err
	}

	return []byte("user deleted successfully"), nil
}

func getUserByEmail(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	requestUrl, err := requests.GetRequestUrl(ctx, PluginName)
	if err != nil {
		return nil, err
	}

	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}

	email := params[emailParam]
	if email == "" {
		return nil, errors.New("email not provided")
	}

	user, err := findUserByEmail(ctx, request, plugin, requestUrl, email)
	if err != nil {
		return nil, err
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		return nil, errors.New("unable to marshal json")
	}

	return userJson, nil
}

func getGroupByDisplayName(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}

	groupName := params[groupNameParam]
	if groupName == "" {
		return nil, errors.New("group name not provided")
	}

	group, err := findGroupByDisplayName(ctx, request, plugin, groupName)
	if err != nil {
		return nil, err
	}

	groupJson, err := json.Marshal(group)
	if err != nil {
		return nil, errors.New("unable to marshal json")
	}

	return groupJson, nil
}
