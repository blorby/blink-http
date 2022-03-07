package bitbucket

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/implementation/requests"
	"github.com/blinkops/blink-http/plugins/types"
	"github.com/blinkops/blink-sdk/plugin"
	"net/http"
	"strings"
)

func addUserToGroup(ctx *plugin.ActionContext, plugin types.Plugin, workspaceId string, groupSlug string, user string) (AddUserToGroupResponse, error) {
	marshalledReqBody, err := json.Marshal("{}")
	if err != nil {
		return AddUserToGroupResponse{}, errors.New("failed to marshal json")
	}
	url := fmt.Sprintf("https://api.bitbucket.org/1.0/groups/%s/%s/members/%s", workspaceId, groupSlug, user)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	res, err := requests.SendRequest(ctx, plugin, http.MethodPut, url, consts.DefaultTimeout, headers, nil, marshalledReqBody)
	if err != nil {
		return AddUserToGroupResponse{}, err
	}

	if res.StatusCode == http.StatusOK {
		return AddUserToGroupResponse{user, "success"}, nil
	}

	return AddUserToGroupResponse{user, strings.ToLower(string(res.Body))}, nil
}
