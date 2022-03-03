package grafana

import (
	"fmt"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/plugins/connections"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type GrafanaPlugin struct{}

func (p GrafanaPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := connections.HeaderValuePrefixes{"AUTHORIZATION": consts.BearerAuthPrefix}
	aliases := connections.HeaderAlias{"API_KEY": "AUTHORIZATION"}
	return connections.HandleGenericConnection(conn, req, prefixes, aliases)
}

func (p GrafanaPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	requestUrl, ok := connection.Data[consts.RequestUrlKey]
	if !ok {
		return false, []byte("Test connection failed, API Address wasn't provided")
	}
	res, err := connections.SendTestConnectionRequest(requestUrl+"/api/org", http.MethodGet, nil, connection, p.HandleAuth)
	if err != nil {
		return false, []byte("Test connection failed. " + err.Error())
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return false, []byte(fmt.Sprintf("Test connection failed. Got status code %v", res.StatusCode))
	}

	return true, nil
}

func (p GrafanaPlugin) GetDefaultRequestUrl() string {
	return "http://localhost:3000"
}

func GetNewGrafanaPlugin() GrafanaPlugin {
	return GrafanaPlugin{}
}
