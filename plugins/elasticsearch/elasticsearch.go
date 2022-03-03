package elasticsearch

import (
	"github.com/blinkops/blink-http/plugins/connections"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type ElasticSearchPlugin struct{}

func (p ElasticSearchPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := connections.HeaderValuePrefixes{"AUTHORIZATION": "ApiKey "}
	aliases := connections.HeaderAlias{"API_KEY": "AUTHORIZATION"}
	return connections.HandleGenericConnection(conn, req, prefixes, aliases)
}

func (p ElasticSearchPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Elasticsearch is not yet supported by the http plugin")
}

func (p ElasticSearchPlugin) GetDefaultRequestUrl() string {
	return ""
}

func GetNewElasticSearchPlugin() ElasticSearchPlugin {
	return ElasticSearchPlugin{}
}
