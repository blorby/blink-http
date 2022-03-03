package pingdom

import (
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/implementation/requests"
	conns "github.com/blinkops/blink-http/plugins/connections"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type PingdomPlugin struct{}

func (p PingdomPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := conns.HeaderValuePrefixes{"AUTHORIZATION": consts.BearerAuthPrefix}
	aliases := conns.HeaderAlias{"TOKEN": "AUTHORIZATION"}
	return conns.HandleGenericConnection(conn, req, prefixes, aliases)
}

func (p PingdomPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	res, err := conns.SendTestConnectionRequest("https://api.pingdom.com/api/3.1/credits", http.MethodGet, nil, connection, p.HandleAuth)
	if err != nil {
		return false, []byte(err.Error())
	}

	body, err := requests.ReadBody(res.Body)
	if err != nil {
		return false, []byte(err.Error())
	}

	_, err = requests.ValidateResponse(res.StatusCode, body)
	if err != nil {
		return false, []byte(err.Error())
	}

	return true, body
}

func (p PingdomPlugin) GetDefaultRequestUrl() string {
	return "https://api.pingdom.com/api/3.1"
}

func GetNewPingdomPlugin() PingdomPlugin {
	return PingdomPlugin{}
}
