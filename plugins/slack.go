package plugins

import (
	"github.com/blinkops/blink-http/consts"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type SlackPlugin struct{}

func (p SlackPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := HeaderValuePrefixes{"AUTHORIZATION": consts.BearerAuthPrefix}
	aliases := HeaderAlias{"TOKEN": "AUTHORIZATION"}
	return handleGenericConnection(conn, req, prefixes, aliases)
}

func (p SlackPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Slack is not yet supported by the http plugin")
}

func getNewSlackPlugin() SlackPlugin {
	return SlackPlugin{}
}
