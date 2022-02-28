package prometheus

import (
	"github.com/blinkops/blink-http/plugins/connections"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type PrometheusPlugin struct{}

func (p PrometheusPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	return connections.HandleGenericConnection(conn, req, nil, nil)
}

func (p PrometheusPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Prometheus is not yet supported by the http plugin")
}

func GetNewPrometheusPlugin() PrometheusPlugin {
	return PrometheusPlugin{}
}