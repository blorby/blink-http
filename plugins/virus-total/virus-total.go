package virus_total

import (
	"github.com/blinkops/blink-http/plugins/connections"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type VirusTotalPlugin struct{}

func (p VirusTotalPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	aliases := connections.HeaderAlias{"API KEY": "x-apikey"}
	return connections.HandleGenericConnection(conn, req, nil, aliases)
}

func (p VirusTotalPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, VirusTotal is not yet supported by the http plugin")
}

func (p VirusTotalPlugin) GetDefaultRequestUrl() string {
	return "https://www.virustotal.com"
}

func GetNewVirusTotalPlugin() VirusTotalPlugin {
	return VirusTotalPlugin{}
}
