package actions

import (
	"encoding/json"
	"errors"
	"github.com/blinkops/blink-openapi-sdk/consts"
	customact "github.com/blinkops/blink-openapi-sdk/plugin/custom_actions"
	"github.com/blinkops/blink-sdk/plugin"
	"strings"
	"sync"
)

const (
	PluginName       = "bitbucket"
	workspaceIdParam = "Workspace ID"
	groupSlugParam   = "Group Slug"
	usersParam       = "Users"
)

type AddUserToGroupResponse struct {
	User   string `json:"user"`
	Status string `json:"status"`
}

func GetBitbucketCustomActions() customact.CustomActions {
	actions := map[string]customact.ActionHandler{
		"AddUsersToGroup": addUsersToGroup,
	}

	return customact.CustomActions{
		Actions:           actions,
		ActionsFolderPath: "custom_actions/actions",
	}
}

func addUsersToGroup(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) (*plugin.ExecuteActionResponse, error) {
	params, err := request.GetParameters()
	if err != nil {
		return nil, errors.New("invalid parameters provided")
	}
	workspaceId := params[workspaceIdParam]
	groupSlug := params[groupSlugParam]

	if workspaceId == "" {
		return nil, errors.New("no workspace id provided")
	}
	if groupSlug == "" {
		return nil, errors.New("no group slug provided")
	}
	if params[usersParam] == "" {
		return nil, errors.New("no user ids provided")
	}

	users := strings.Split(strings.ReplaceAll(params[usersParam], " ", ""), ",")

	statuses := make(chan AddUserToGroupResponse, len(users))
	errorsChan := make(chan error, len(users))
	var wg sync.WaitGroup
	for _, user := range users {
		userVal := user
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, err := addUserToGroup(ctx, workspaceId, groupSlug, userVal)
			statuses <- status
			if err != nil {
				errorsChan <- err
			}
		}()
	}
	wg.Wait()

	close(statuses)
	close(errorsChan)

	// if there's an error, we'll just return the first one
	if len(errorsChan) > 0 {
		err := <-errorsChan
		if err != nil {
			return nil, err
		}
	}

	var response []AddUserToGroupResponse
	for range users {
		response = append(response, <-statuses)
	}
	marshalledResponse, err := json.Marshal(response)
	if err != nil {
		return nil, errors.New("failed to marshal json")
	}

	return &plugin.ExecuteActionResponse{ErrorCode: consts.OK, Result: marshalledResponse}, nil
}
