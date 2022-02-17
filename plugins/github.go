package plugins

import (
	"github.com/blinkops/blink-http/consts"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type GithubPlugin struct{}

func (p GithubPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := HeaderValuePrefixes{"AUTHORIZATION": consts.BearerAuthPrefix}
	aliases := HeaderAlias{"TOKEN": "AUTHORIZATION"}
	return handleGenericConnection(conn, req, prefixes, aliases)
}

func (p GithubPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Github is not yet supported by the http plugin")
}

func getNewGithubPlugin() GithubPlugin {
	return GithubPlugin{}
}
