package plugins

import (
	"github.com/blinkops/blink-http/consts"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type JiraPlugin struct{}

func (p JiraPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := HeaderValuePrefixes{"AUTHORIZATION": consts.BearerAuthPrefix}
	aliases := HeaderAlias{"API TOKEN": "PASSWORD", "USER EMAIL": "USERNAME"}
	return handleGenericConnection(conn, req, prefixes, aliases)
}

func (p JiraPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Jira is not yet supported by the http plugin")
}

func getNewJiraPlugin() JiraPlugin {
	return JiraPlugin{}
}