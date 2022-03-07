package crowdstrike

import (
	"encoding/json"
	"errors"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/plugins/types"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type CrowdStrikePlugin struct{}

func (p CrowdStrikePlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	requestUrl := conn[requestUrlParam]

	queryParams := url.Values{
		"client_id":     {conn[clientIdParam]},
		"client_secret": {conn[clientSecretParam]},
	}

	client := &http.Client{}

	request, err := http.NewRequest(http.MethodPost, requestUrl+"/oauth2/token", strings.NewReader(queryParams.Encode()))
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(request)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated {
		return errors.New("invalid credentials")
	}

	defer res.Body.Close()

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
	req.Header.Set("AUTHORIZATION", consts.BearerAuthPrefix + responseBody.AccessToken)
	return nil
}

func (p CrowdStrikePlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	req, err := http.NewRequest(http.MethodGet, "https://api.crowdstrike.com/auth.test", nil)
	if err != nil {
		return false, []byte(err.Error())
	}

	err = p.HandleAuth(req, connection.Data)

	if err != nil {
		return false, []byte(consts.TestConnectionFailed)
	}

	return true, nil
}

func (p CrowdStrikePlugin) GetDefaultRequestUrl() string {
	return ""
}

func (p CrowdStrikePlugin) GetCustomActionHandlers() map[string] types.ActionHandler {
	return map[string]types.ActionHandler{
		"GetInstalledDevices": getInstalledDevices,
		"DeleteDevice":        deleteDevice,
	}
}

func (p CrowdStrikePlugin) GetCustomActionsPath() string {
	return "plugins/crowdstrike/actions"
}

func GetNewCrowdStrikePlugin() CrowdStrikePlugin {
	return CrowdStrikePlugin{}
}

