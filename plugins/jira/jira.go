package jira

import (
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/plugins/connections"
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
	return false, []byte("Test connection failed, Jira is not yet supported by the http plugin")
}

func (p JiraPlugin) GetDefaultRequestUrl() string {
	return ".atlassian.net"
}

func GetNewJiraPlugin() JiraPlugin {
	return JiraPlugin{}
}
