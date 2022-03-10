package okta

import (
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/plugins/connections"
	"github.com/blinkops/blink-http/plugins/types"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type OktaPlugin struct{}

func (p OktaPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := connections.HeaderValuePrefixes{"AUTHORIZATION": consts.SswsAuthPrefix}
	aliases := connections.HeaderAlias{"API TOKEN": "AUTHORIZATION"}
	return connections.HandleGenericConnection(conn, req, prefixes, aliases)
}

func (p OktaPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	requestUrl, ok := connection.Data[consts.RequestUrlKey]
	if !ok {
		return false, []byte(consts.RequestUrlMissing)
	}

	res, err := connections.SendTestConnectionRequest(requestUrl+"api/v1/apps", http.MethodGet, nil, connection, p.HandleAuth)
	if err != nil {
		return false, []byte(err.Error())
	}

	if res.StatusCode < http.StatusOK || res.StatusCode > http.StatusIMUsed {
		return false, []byte(consts.TestConnectionFailed)
	}

	return true, nil
}

func (p OktaPlugin) GetDefaultRequestUrl() string {
	return ".okta.com"
}

func (p OktaPlugin) GetCustomActionHandlers() map[string]types.ActionHandler {
	return map[string]types.ActionHandler{
		"ListApplicationUsersByName": listApplicationUsersByName,
	}
}

func (p OktaPlugin) GetCustomActionsPath() string {
	return "plugins/okta/actions"
}

func GetNewOktaPlugin() OktaPlugin {
	return OktaPlugin{}
}
