package opsgenie

import (
	"github.com/blinkops/blink-http/plugins/connections"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type OpsgeniePlugin struct{}

func (p OpsgeniePlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := connections.HeaderValuePrefixes{"AUTHORIZATION": "GenieKey "}
	aliases := connections.HeaderAlias{"TOKEN": "AUTHORIZATION"}
	return connections.HandleGenericConnection(conn, req, prefixes, aliases)
}

func (p OpsgeniePlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Opsgenie is not yet supported by the http plugin")
}

func (p OpsgeniePlugin) GetDefaultRequestUrl() string {
	return "https://api.opsgenie.com"
}

func GetNewOpsgeniePlugin() OpsgeniePlugin {
	return OpsgeniePlugin{}
}
