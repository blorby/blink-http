package jira

import (
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/plugins/connections"
	"github.com/blinkops/blink-http/plugins/types"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type JiraPlugin struct{}

func (p JiraPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := connections.HeaderValuePrefixes{"AUTHORIZATION": consts.BearerAuthPrefix}
	aliases := connections.HeaderAlias{"API TOKEN": "PASSWORD", "USER EMAIL": "USERNAME"}
	return connections.HandleGenericConnection(conn, req, prefixes, aliases)
}

func (p JiraPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	requestUrl, ok := connection.Data[consts.RequestUrlKey]
	if !ok {
		return false, []byte(consts.RequestUrlMissing)
	}
	res, err := connections.SendTestConnectionRequest(requestUrl+"/rest/api/3/filter/my", http.MethodGet, nil, connection, p.HandleAuth)
	if err != nil {
		return false, []byte(err.Error())
	}

	if res.StatusCode != http.StatusOK {
		return false, []byte(consts.TestConnectionFailed)
	}
	return true, nil
}

func (p JiraPlugin) GetDefaultRequestUrl() string {
	return ".atlassian.net"
}

func (p JiraPlugin) GetCustomActionHandlers() map[string]types.ActionHandler {
	return map[string]types.ActionHandler{
		"AddComment":       addComment,
		"CreateIssue":      createIssue,
		"RemoveUser":       removeUser,
		"ListCustomFields": listCustomFields,
	}
}

func (p JiraPlugin) GetCustomActionsPath() string {
	return "plugins/jira/actions"
}

func GetNewJiraPlugin() JiraPlugin {
	return JiraPlugin{}
}
