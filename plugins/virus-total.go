package plugins

import (
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type VirusTotalPlugin struct{}

func (p VirusTotalPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	aliases := HeaderAlias{"API KEY": "x-apikey"}
	return handleGenericConnection(conn, req, nil, aliases)
}

func (p VirusTotalPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, VirusTotal is not yet supported by the http plugin")
}

func getNewVirusTotalPlugin() VirusTotalPlugin {
	return VirusTotalPlugin{}
}