package okta

import (
	"encoding/json"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/implementation/requests"
	"github.com/blinkops/blink-http/plugins/types"
	"net/http"
	"net/url"

	"github.com/blinkops/blink-sdk/plugin"
)

func getApplicationIDByName(ctx *plugin.ActionContext, plugin types.Plugin, requestUrl string, applicationName string, timeout int32) (string, error) {
	finalUrl := fmt.Sprintf("%s/api/v1/apps?q=%s", requestUrl, url.QueryEscape(applicationName))
	res, err := requests.SendRequest(ctx, plugin, http.MethodGet, finalUrl, consts.DefaultTimeout, nil, nil, nil)
	if err != nil {
		return "", err
	}

	var apps []struct {
		ApplicationID string `json:"id"`
	}
	if err = json.Unmarshal(res.Body, &apps); err != nil {
		return "", err
	}
	if len(apps) == 0 {
		return "", fmt.Errorf("the application name input did not match any application")
	}
	if len(apps) > 1 {
		return "", fmt.Errorf("the application name input matched more than one application")
	}

	return apps[0].ApplicationID, nil
}
