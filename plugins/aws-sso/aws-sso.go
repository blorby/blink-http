package aws_sso

import (
	"errors"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/plugins/connections"
	"github.com/blinkops/blink-http/plugins/types"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
	"strings"
)

type AwsSsoPlugin struct{}

func (p AwsSsoPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := connections.HeaderValuePrefixes{"AUTHORIZATION": consts.BearerAuthPrefix}
	aliases := connections.HeaderAlias{"TOKEN": "AUTHORIZATION"}
	return connections.HandleGenericConnection(conn, req, prefixes, aliases)
}

func (p AwsSsoPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	requestUrl, ok := connection.Data[consts.RequestUrlKey]
	if !ok {
		return false, []byte("Test connection failed, API Address wasn't provided")
	}
	if !strings.Contains(requestUrl, "scim.") || !strings.Contains(requestUrl, ".amazonaws.com") {
		return false, []byte("invalid request url")
	}
	res, err := connections.SendTestConnectionRequest(requestUrl+"/Users", http.MethodGet, nil, connection, p.HandleAuth)
	if err != nil {
		return false, []byte("Test connection failed. " + err.Error())
	}
	if res.StatusCode != http.StatusOK {
		return false, []byte(fmt.Sprintf("Test connection failed. Got status code %v", res.StatusCode))
	}

	return true, nil
}

func (p AwsSsoPlugin) GetDefaultRequestUrl() string {
	return ".amazonaws.com"
}

func (p AwsSsoPlugin) GetCustomActionHandlers() map[string] types.ActionHandler {
	return map[string]types.ActionHandler{
		"AddUsersToGroup":       addUsersToGroup,
		"CreateUser":            createUser,
		"DeleteUser":            deleteUser,
		"GetUserByEmail":        getUserByEmail,
		"GetGroupByDisplayName": getGroupByDisplayName,
	}
}

func (p AwsSsoPlugin) GetCustomActionsPath() string {
	return "plugins/aws-sso/actions"
}

func (p AwsSsoPlugin) ValidateResponse(response *types.Response) (*types.Response, error) {
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusBadRequest {
		return response, errors.New(string(response.Body))
	}

	return response, nil
}

func GetNewAwsSsoPlugin() AwsSsoPlugin {
	return AwsSsoPlugin{}
}

