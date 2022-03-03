package pagerduty

import (
	"fmt"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type PagerdutyPlugin struct{}

func (p PagerdutyPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	// Api key
	if val, ok := conn["api_key"]; ok {
		req.Header.Set("AUTHORIZATION", "Token token="+val)
		return nil
	}

	// OAuth
	if val, ok := conn["Token"]; ok {
		req.Header.Set("AUTHORIZATION", "Bearer "+val)
		req.Header.Set("Accept", "application/vnd.pagerduty+json;version=2")
		return nil
	}

	log.Info("failed to set authentication headers")
	return fmt.Errorf("failed to set authentication headers")
}

func (p PagerdutyPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Pagerduty is not yet supported by the http plugin")
}

func (p PagerdutyPlugin) GetDefaultRequestUrl() string {
	return "https://api.pagerduty.com"
}

func GetNewPagerdutyPlugin() PagerdutyPlugin {
	return PagerdutyPlugin{}
}
