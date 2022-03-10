package okta

import (
	"errors"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/implementation/requests"
	"github.com/blinkops/blink-http/plugins/types"
	"net/http"
	"net/url"

	"github.com/blinkops/blink-sdk/plugin"
)

const (
	pluginName = "okta"
	applicationNameParam = "Application Name"
	queryParam = "Query"
	limitParam = "Limit"
)

func listApplicationUsersByName(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	requestUrl, err := requests.GetRequestUrl(ctx, pluginName)
	if err != nil {
		return nil, errors.New(consts.RequestUrlMissing)
	}

	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}

	applicationName := params[applicationNameParam]
	if applicationName == "" {
		return nil, errors.New("application name not provided")
	}

	applicationID, err := getApplicationIDByName(ctx, plugin, requestUrl, applicationName, request.Timeout)
	if err != nil {
		return nil, err
	}

	query, limit := params[queryParam], params[limitParam]

	finalUrl := fmt.Sprintf("%s/api/v1/apps/%s/users?q=%s&limit=%s", requestUrl, applicationID, url.QueryEscape(query), url.QueryEscape(limit))
	res, err := requests.SendRequest(ctx, plugin, http.MethodGet, finalUrl, consts.DefaultTimeout, nil, nil, nil)

	return res.Body, err
}
