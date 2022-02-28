package plugins

import (
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/implementation"
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
	res, err := sendTestConnectionRequest("https://api.pingdom.com/api/3.1/credits", http.MethodGet, nil, connection, p.HandleAuth)
	if err != nil {
		return false, []byte(err.Error())
	}

	body, err := implementation.ReadBody(res.Body)
	if err != nil {
		return false, []byte(err.Error())
	}

	_, err = implementation.ValidateResponse(res.StatusCode, body)
	if err != nil {
		return false, []byte(err.Error())
	}

	return true, body
}

func getNewPingdomPlugin() PingdomPlugin {
	return PingdomPlugin{}
}
