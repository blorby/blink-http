package github

import (
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/plugins/connections"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type GithubPlugin struct{}

func (p GithubPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := connections.HeaderValuePrefixes{"AUTHORIZATION": consts.BearerAuthPrefix}
	aliases := connections.HeaderAlias{"TOKEN": "AUTHORIZATION"}
	return connections.HandleGenericConnection(conn, req, prefixes, aliases)
}

func (p GithubPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	requestUrl, ok := connection.Data[consts.RequestUrlKey]
	if !ok {
		requestUrl = p.GetDefaultRequestUrl()
	}

	res, err := connections.SendTestConnectionRequest(requestUrl+"/user", http.MethodGet, nil, connection, p.HandleAuth)
	if err != nil {
		return false, []byte(err.Error())
	}

	if res.StatusCode != http.StatusOK {
		return false, []byte(consts.TestConnectionFailed)
	}

	return true, nil
}

func (p GithubPlugin) GetDefaultRequestUrl() string {
	return "https://api.github.com"
}
func GetNewGithubPlugin() GithubPlugin {
	return GithubPlugin{}
}
