package grafana

import (
	"fmt"
	"github.com/blinkops/blink-http/consts"
	http_conn "github.com/blinkops/blink-http/implementation/connections"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

func GetGrafanaConnectionOpts() http_conn.GenericConnectionOpts {
	return http_conn.GenericConnectionOpts{
		HeaderValuePrefixes: http_conn.HeaderValuePrefixes{"AUTHORIZATION": consts.BearerAuthPrefix},
		HeaderAlias:         http_conn.HeaderAlias{"API_KEY": "AUTHORIZATION"},
	}
}

func TestGrafanaConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	requestUrl, ok := connection.Data[consts.RequestUrlKey]
	if !ok {
		return false, []byte("Test connection failed, API Address wasn't provided")
	}
	res, err := http_conn.SendTestConnectionRequest(requestUrl+"/api/org", http.MethodGet, nil, connection)
	if err != nil {
		return false, []byte("Test connection failed. " + err.Error())
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return false, []byte(fmt.Sprintf("Test connection failed. Got status code %v", res.StatusCode))
	}

	return true, nil
}