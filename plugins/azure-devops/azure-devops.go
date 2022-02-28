package azure_devops

import (
	"encoding/base64"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type AzureDevopsPlugin struct{}

func (p AzureDevopsPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	// OAuth
	if val, ok := conn["Token"]; ok {
		req.Header.Set("AUTHORIZATION", consts.BearerAuthPrefix+val)
		return nil
	}
	// Basic Authentication
	username, userExists := conn["Username"]
	password, passExists := conn["Access Token"]
	if !userExists || !passExists {
		log.Info("failed to set authentication headers")
		return fmt.Errorf("failed to set authentication headers")
	}
	data := []byte(fmt.Sprintf("%s:%s", username, password))
	hashed := base64.StdEncoding.EncodeToString(data)
	req.Header.Set("AUTHORIZATION", consts.BasicAuthPrefix+hashed)
	return nil
}

func (p AzureDevopsPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, AzureDevops is not yet supported by the http plugin")
}

func GetNewAzureDevopsPlugin() AzureDevopsPlugin {
	return AzureDevopsPlugin{}
}