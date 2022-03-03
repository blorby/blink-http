package gitlab

import (
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/plugins/connections"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type GitlabPlugin struct{}

func (p GitlabPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := connections.HeaderValuePrefixes{"AUTHORIZATION": consts.BearerAuthPrefix}
	aliases := connections.HeaderAlias{"TOKEN": "AUTHORIZATION"}
	return connections.HandleGenericConnection(conn, req, prefixes, aliases)
}

func (p GitlabPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Gitlab is not yet supported by the http plugin")
}

func (p GitlabPlugin) GetDefaultRequestUrl() string {
	return "https://gitlab.com/api/v4"
}
func GetNewGitlabPlugin() GitlabPlugin {
	return GitlabPlugin{}
}
