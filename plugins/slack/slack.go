package slack

import (
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/plugins/connections"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type SlackPlugin struct{}

func (p SlackPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := connections.HeaderValuePrefixes{"AUTHORIZATION": consts.BearerAuthPrefix}
	aliases := connections.HeaderAlias{"TOKEN": "AUTHORIZATION"}
	return connections.HandleGenericConnection(conn, req, prefixes, aliases)
}

func (p SlackPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Slack is not yet supported by the http plugin")
}

func (p SlackPlugin) GetDefaultRequestUrl() string {
	return "https://slack.com/api"
}
func GetNewSlackPlugin() SlackPlugin {
	return SlackPlugin{}
}
