package elasticsearch

import (
	"encoding/json"
	"errors"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/plugins/connections"
	"github.com/blinkops/blink-http/plugins/types"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
	"strings"
)

type ElasticSearchPlugin struct{}

func (p ElasticSearchPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	prefixes := connections.HeaderValuePrefixes{"AUTHORIZATION": "ApiKey "}
	aliases := connections.HeaderAlias{"API_KEY": "AUTHORIZATION"}
	return connections.HandleGenericConnection(conn, req, prefixes, aliases)
}

func (p ElasticSearchPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	requestUrl, ok := connection.Data[consts.RequestUrlKey]
	if !ok {
		return false, []byte(consts.RequestUrlMissing)
	}

	res, err := connections.SendTestConnectionRequest(requestUrl, http.MethodGet, nil, connection, p.HandleAuth)
	if err != nil {
		return false, []byte(err.Error())
	}

	if res.StatusCode == http.StatusUnauthorized {
		return false, []byte(consts.TestConnectionFailed)
	}

	res, err = p.ValidateResponse(res)
	if err != nil {
		return false, []byte(err.Error())
	}

	return true, nil
}

func (p ElasticSearchPlugin) GetDefaultRequestUrl() string {
	return ""
}

func (p ElasticSearchPlugin) ValidateResponse(response *types.Response) (*types.Response, error) {
	type ElasticSearchError struct {
		Error *struct {
			Reason       string `json:"reason"`
		} `json:"error,omitempty"`
	}

	data := ElasticSearchError{}

	err := json.Unmarshal(response.Body, &data)
	if err != nil {
		if strings.Contains(string(response.Body), "html") {
			return nil, errors.New("invalid URL")
		}
		return nil, errors.New("invalid JSON from server")
	}

	if data.Error != nil {
		return nil, errors.New(string(response.Body))
	}

	return nil, nil
}

func GetNewElasticSearchPlugin() ElasticSearchPlugin {
	return ElasticSearchPlugin{}
}
