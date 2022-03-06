package azure_devops

import (
	"encoding/base64"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/plugins/connections"
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
	organization, ok := connection.Data["Organization"]

	if !ok {
		return false, []byte("organization name not provided")
	}

	res, err := connections.SendTestConnectionRequest(fmt.Sprintf("%s/%s/_apis/projects", p.GetDefaultRequestUrl(), organization), http.MethodGet, nil, connection, p.HandleAuth)
	if err != nil {
		return false, []byte(err.Error())
	}

	if res.StatusCode != http.StatusOK {
		return false, []byte("invalid access token")
	}

	return true, nil
}

func (p AzureDevopsPlugin) GetDefaultRequestUrl() string {
	return "https://dev.azure.com"
}

func GetNewAzureDevopsPlugin() AzureDevopsPlugin {
	return AzureDevopsPlugin{}
}
