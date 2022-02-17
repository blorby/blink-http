package plugins

import (
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type ElasticsearchPlugin struct{}

func (p ElasticsearchPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := HeaderValuePrefixes{"AUTHORIZATION": "ApiKey "}
	aliases := HeaderAlias{"API_KEY": "AUTHORIZATION"}
	return handleGenericConnection(conn, req, prefixes, aliases)
}

func (p ElasticsearchPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Elasticsearch is not yet supported by the http plugin")
}

func getNewElasticsearchPlugin() ElasticsearchPlugin {
	return ElasticsearchPlugin{}
}
