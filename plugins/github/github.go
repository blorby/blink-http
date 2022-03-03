package github

import (
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/plugins/connections"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type GithubPlugin struct{}

func (p GithubPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := connections.HeaderValuePrefixes{"AUTHORIZATION": consts.BearerAuthPrefix}
	aliases := connections.HeaderAlias{"TOKEN": "AUTHORIZATION"}
	return connections.HandleGenericConnection(conn, req, prefixes, aliases)
}

func (p GithubPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Github is not yet supported by the http plugin")
}

func (p GithubPlugin) GetDefaultRequestUrl() string {
	return "https://api.github.com"
}
func GetNewGithubPlugin() GithubPlugin {
	return GithubPlugin{}
}
