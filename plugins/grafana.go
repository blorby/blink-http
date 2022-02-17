package plugins

import (
	"fmt"
	"github.com/blinkops/blink-http/consts"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type GrafanaPlugin struct{}

func (p GrafanaPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := HeaderValuePrefixes{"AUTHORIZATION": consts.BearerAuthPrefix}
	aliases := HeaderAlias{"API_KEY": "AUTHORIZATION"}
	return handleGenericConnection(conn, req, prefixes, aliases)
}

func (p GrafanaPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	requestUrl, ok := connection.Data[consts.RequestUrlKey]
	if !ok {
		return false, []byte("Test connection failed, API Address wasn't provided")
	}
	res, err := sendTestConnectionRequest(requestUrl+"/api/org", http.MethodGet, nil, connection, p.HandleAuth)
	if err != nil {
		return false, []byte("Test connection failed. " + err.Error())
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return false, []byte(fmt.Sprintf("Test connection failed. Got status code %v", res.StatusCode))
	}

	return true, nil
}

func getNewGrafanaPlugin() GrafanaPlugin {
	return GrafanaPlugin{}
}