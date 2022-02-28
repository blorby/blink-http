package okta

import (
	"github.com/blinkops/blink-http/plugins/connections"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type OktaPlugin struct{}

func (p OktaPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	return connections.HandleGenericConnection(conn, req, nil, nil)
}

func (p OktaPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Okta is not yet supported by the http plugin")
}

func GetNewOktaPlugin() OktaPlugin {
	return OktaPlugin{}
}
