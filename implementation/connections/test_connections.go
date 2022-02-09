package connections

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/implementation/requests"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
	"time"
)



type testConnectionFunc func(*blink_conn.ConnectionInstance) (bool, []byte)

var testConnectionHandlers = map[string]testConnectionFunc{
	"grafana": testGrafanaConnection,
}

func TestConnection(connections map[string]*blink_conn.ConnectionInstance) (bool, []byte) {
	for connName, connInstance := range connections {
		handler, ok := testConnectionHandlers[connName]
		if !ok {
			return false, []byte(fmt.Sprintf("Test connection failed. Connection type %s isn't supported by the http plugin", connName))
		}
		return handler(connInstance)
	}
	return false, []byte(fmt.Sprintf("Test connection failed. No connection to test was provided"))
}

func testGrafanaConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	requestUrl, ok := connection.Data[consts.RequestUrlKey]
	if !ok {
		return false, []byte("Test connection failed, API Address wasn't provided")
	}
	res, err := sendTestConnectionRequest(requestUrl+"/api/org", http.MethodGet, nil, connection)
	if err != nil {
		return false, []byte("Test connection failed. Got " + err.Error())
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return false, []byte(fmt.Sprintf("Test connection failed. Got status code %v", res.StatusCode))
	}

	var grafanaErr struct {
		Message string `json:"message"`
	}

	body, err := requests.ReadBody(res.Body)
	err = json.Unmarshal(body, &grafanaErr)
	if err == nil && grafanaErr.Message != "" {
		return false, []byte(grafanaErr.Message)
	}

	return false, body
}

func sendTestConnectionRequest(url string, method string, data []byte, conn *blink_conn.ConnectionInstance) (*http.Response, error) {
	requestBody := bytes.NewBuffer(data)
	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, err
	}

	if err = HandleAuthSingleConnection(req, conn); err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: time.Second * time.Duration(consts.DefaultTimeout),
	}
	return client.Do(req)
}
