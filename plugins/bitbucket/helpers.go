package actions

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blinkops/blink-openapi-sdk/consts"
	openapi "github.com/blinkops/blink-openapi-sdk/plugin"
	"github.com/blinkops/blink-sdk/plugin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func SetCustomAuthHeaders(connection map[string]interface{}, request *http.Request) error {
	// OAuth
	if val, ok := connection["Token"]; ok {
		if valString, ok := val.(string); ok {
			request.Header.Set("AUTHORIZATION", consts.BearerAuth+valString)
			return nil
		}
	}
	// Api key
	username, userExists := connection["USERNAME"]
	password, passExists := connection["PASSWORD"]
	if !userExists || !passExists {
		log.Info("failed to set authentication headers")
		return fmt.Errorf("failed to set authentication headers")
	}
	if userString, ok := username.(string); ok {
		if passString, ok := password.(string); ok {
			data := []byte(fmt.Sprintf("%s:%s", userString, passString))
			hashed := base64.StdEncoding.EncodeToString(data)
			request.Header.Set("AUTHORIZATION", consts.BasicAuth+hashed)
			return nil
		}
	}
	log.Info("failed to set authentication headers")
	return fmt.Errorf("failed to set authentication headers")
}

func execRequest(ctx *plugin.ActionContext, request *http.Request) *plugin.ExecuteActionResponse {
	res := &plugin.ExecuteActionResponse{ErrorCode: consts.OK}
	response, err := openapi.ExecuteRequest(ctx, request, PluginName, nil, nil, 30, SetCustomAuthHeaders)
	if err != nil {
		res.Result = []byte(err.Error())
		res.ErrorCode = consts.Error
		return res
	}
	res.Result = response.Body
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return res
	}
	res.ErrorCode = consts.Error
	return res
}

func addUserToGroup(ctx *plugin.ActionContext, workspaceId string, groupSlug string, user string) (AddUserToGroupResponse, error) {
	marshalledReqBody, err := json.Marshal("{}")
	if err != nil {
		return AddUserToGroupResponse{}, errors.New("failed to marshal json")
	}
	url := fmt.Sprintf("https://api.bitbucket.org/1.0/groups/%s/%s/members/%s", workspaceId, groupSlug, user)
	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(marshalledReqBody))
	req.Header.Set("Content-Type", "application/json")

	response, err := openapi.ExecuteRequest(ctx, req, PluginName, nil, nil, 30, SetCustomAuthHeaders)
	if err != nil {
		return AddUserToGroupResponse{}, err
	}

	if response.StatusCode == http.StatusOK {
		return AddUserToGroupResponse{user, "success"}, nil
	}

	return AddUserToGroupResponse{user, strings.ToLower(string(response.Body))}, nil
}
