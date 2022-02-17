package plugins

import (
	"github.com/blinkops/blink-http/consts"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type PingdomPlugin struct{}

func (p PingdomPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := HeaderValuePrefixes{"AUTHORIZATION": consts.BearerAuthPrefix}
	aliases := HeaderAlias{"TOKEN": "AUTHORIZATION"}
	return handleGenericConnection(conn, req, prefixes, aliases)
}

func (p PingdomPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Pingdom is not yet supported by the http plugin")
}

func getNewPingdomPlugin() PingdomPlugin {
	return PingdomPlugin{}
}
