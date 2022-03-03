package azure

import (
	"encoding/json"
	"fmt"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type AzurePlugin struct{}

func (p AzurePlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	// handle oauth
	if token, ok := conn["Token"]; ok {
		req.Header.Set("AUTHORIZATION", "Bearer "+token)
		return nil
	}

	queryParams := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {conn["app_id"]},
		"client_secret": {conn["client_secret"]},
		"resource":      {"https://management.core.windows.net/"},
	}

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", conn["tenant_id"]), strings.NewReader(queryParams.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = res.Body.Close() }()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var responseBody struct {
		AccessToken string `json:"access_token"`
	}

	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return err
	}
	req.Header.Set("AUTHORIZATION", "Bearer "+responseBody.AccessToken)
	return nil

}

func (p AzurePlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Azure is not yet supported by the http plugin")
}

func (p AzurePlugin) GetDefaultRequestUrl() string {
	return "https://management.azure.com"
}

func GetNewAzurePlugin() AzurePlugin {
	return AzurePlugin{}
}
