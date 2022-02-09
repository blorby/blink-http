package connections

import (
	"bytes"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/integrations/grafana"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
	"time"
)



type TestConnectionFunc func(*blink_conn.ConnectionInstance) (bool, []byte)

var testConnectionHandlers = map[string]TestConnectionFunc{
	"grafana": grafana.TestGrafanaConnection,
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



func SendTestConnectionRequest(url string, method string, data []byte, conn *blink_conn.ConnectionInstance) (*http.Response, error) {
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
