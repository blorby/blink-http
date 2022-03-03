package wiz

import (
	"encoding/json"
	"errors"
	"github.com/blinkops/blink-http/plugins/types"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type WizPlugin struct{}

func (p WizPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	queryParams := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {conn["Client ID"]},
		"client_secret": {conn["Client Secret"]},
		"audience":      {"beyond-api"},
	}

	client := &http.Client{}

	request, err := http.NewRequest(http.MethodPost, "https://auth.wiz.io/oauth/token", strings.NewReader(queryParams.Encode()))
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(request)
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

	if responseBody.AccessToken == "" {
		return errors.New("invalid credentials")
	}

	req.Header.Set("AUTHORIZATION", "Bearer "+responseBody.AccessToken)
	return nil
}

func (p WizPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {

	req, _ := http.NewRequest(http.MethodGet, "https://api.wiz.com/auth.test", nil)

	err := p.HandleAuth(req, connection.Data)

	if err != nil {
		return false, []byte(err.Error())
	}

	return true, nil
}

func (p WizPlugin) GetDefaultRequestUrl() string {
	return ""
}

func (p WizPlugin) GetCustomActionHandlers() map[string] types.ActionHandler {
	return map[string] types.ActionHandler{
		"ListCloudConfigurationRules": listCloudConfigurationRules,
		"ListControls":                listControls,
	}
}

func (p WizPlugin) GetCustomActionsPath() string {
	return "plugins/wiz/actions"
}

func GetNewWizPlugin() WizPlugin {
	return WizPlugin{}
}
