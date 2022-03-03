package pingdom

import (
	"github.com/blinkops/blink-http/implementation/requests"
	"github.com/blinkops/blink-http/plugins/types"
	"github.com/blinkops/blink-sdk/plugin"
	"net/http"
	"net/url"
)

func (p PingdomPlugin) GetCustomActionHandlers() map[string]types.ActionHandler {
	return map[string]types.ActionHandler{
		"CreateWebUptimeCheck": CreateWebUptimeCheck,
	}
}

func (p PingdomPlugin) GetCustomActionsPath() string {
	return "/plugins/pingdom/actions"
}

func CreateWebUptimeCheck(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}

	_url := params["URL"]
	name := params["Name"]

	form := url.Values{
		"name": {name},
		"host": {_url},
		"type": {"http"},
	}

	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	return requests.SendRequest(ctx, plugin, http.MethodPost, "https://api.pingdom.com/api/3.1/checks", request.Timeout, nil, headers, []byte(form.Encode()))
}
