package ermetic

import (
	"encoding/json"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/plugins/connections"
	"github.com/blinkops/blink-http/plugins/types"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type ErmeticPlugin struct{}

func (p ErmeticPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := connections.HeaderValuePrefixes{"AUTHORIZATION": consts.BearerAuthPrefix}
	aliases := connections.HeaderAlias{"TOKEN": "AUTHORIZATION"}
	return connections.HandleGenericConnection(conn, req, prefixes, aliases)
}

func (p ErmeticPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	query := `query {
  Findings(filter: {
    Statuses: Open,
    CreationTimeStart: "2021-07-15T00:00:00",
    CreationTimeEnd: "2021-08-15T00:00:00"
  }) {
    items {
      Title
      Status
      CreationTime
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
    }
  }
}`

	body, err := json.Marshal(map[string]interface{}{"query": query})
	if err != nil {
		return false, []byte("failed to marshal request body")
	}
	res, err := connections.SendTestConnectionRequest(p.GetDefaultRequestUrl(), http.MethodPost, body, connection, p.HandleAuth)
	if res.StatusCode == http.StatusUnauthorized {
		return false, []byte(consts.TestConnectionFailed)
	}

	return true, nil
}

func (p ErmeticPlugin) GetDefaultRequestUrl() string {
	return "https://us.app.ermetic.com/api/graph"
}

func (p ErmeticPlugin) GetCustomActionHandlers() map[string]types.ActionHandler {
	return map[string]types.ActionHandler{
		"ListFindings": listFindings,
		"Query":        query,
		"ListEntities": listEntities,
	}
}

func (p ErmeticPlugin) GetCustomActionsPath() string {
	return "plugins/ermetic/actions"
}

func GetNewErmeticPlugin() ErmeticPlugin {
	return ErmeticPlugin{}
}
