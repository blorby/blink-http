package pingdom

import (
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/plugins/connections"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type PingdomPlugin struct{}

func (p PingdomPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := connections.HeaderValuePrefixes{"AUTHORIZATION": consts.BearerAuthPrefix}
	aliases := connections.HeaderAlias{"TOKEN": "AUTHORIZATION"}
	return connections.HandleGenericConnection(conn, req, prefixes, aliases)
}

func (p PingdomPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Pingdom is not yet supported by the http plugin")
}

func (p PingdomPlugin) GetDefaultRequestUrl() string {
	return "https://api.pingdom.com/api/3.1"
}

func GetNewPingdomPlugin() PingdomPlugin {
	return PingdomPlugin{}
}
