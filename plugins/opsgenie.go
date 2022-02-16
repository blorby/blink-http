package plugins

import (
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type OpsgeniePlugin struct{}

func (p OpsgeniePlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := HeaderValuePrefixes{"AUTHORIZATION": "GenieKey "}
	aliases := HeaderAlias{"TOKEN": "AUTHORIZATION"}
	return handleGenericConnection(conn, req, prefixes, aliases)
}

func (p OpsgeniePlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Opsgenie is not yet supported by the http plugin")
}

func getNewOpsgeniePlugin() OpsgeniePlugin {
	return OpsgeniePlugin{}
}
