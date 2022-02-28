package datadog

import (
	"github.com/blinkops/blink-http/plugins/connections"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type DatadogPlugin struct{}

func (p DatadogPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	return connections.HandleGenericConnection(conn, req, nil, nil)
}

func (p DatadogPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Datadog is not yet supported by the http plugin")
}

func GetNewDatadogPlugin() DatadogPlugin {
	return DatadogPlugin{}
}