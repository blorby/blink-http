package crowdstrike

import (
	"encoding/json"
	"errors"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/implementation/requests"
	"github.com/blinkops/blink-http/plugins/types"
	"github.com/blinkops/blink-sdk/plugin"
	"net/http"
)

const (
	PluginName             = "crowdstrike"
	hostAgentIdParam       = "Host Agent ID"
	deviceSerialsParam     = "Device Serials"
	onlyActiveDevicesParam = "Return only active devices"
)

type QueryDevicesByFilterResponse struct {
	Resources []string `json:"resources"`
}

type GetDeviceResponse struct {
	Resources []struct {
		Status string `json:"detection_suppression_status,omitempty"`
	} `json:"resources"`
}

func getInstalledDevices(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	deviceSerials, onlyActiveDevices, err := getGetInstalledDevicesParams(request)

	activeSerialsChan, err := performGetInstalledDevices(ctx, plugin, request, deviceSerials, onlyActiveDevices)
	if err != nil {
		return nil, err
	}

	var activeSerials []string
	for len(activeSerialsChan) > 0 {
		activeSerials = append(activeSerials, <-activeSerialsChan)
	}

	activeSerialsStr, err := json.Marshal(activeSerials)
	if err != nil {
		return nil, errors.New("failed to marshal active serials json")
	}

	return activeSerialsStr, nil
}

func deleteDevice(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	requestUrl, err := requests.GetRequestUrl(ctx, PluginName)
	if err != nil {
		return nil, errors.New(consts.RequestUrlMissing)
	}

	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}

	id := params[hostAgentIdParam]

	reqBody := json.RawMessage(`{
  "ids": [
    "` + id + `"
  ]
}`)
	marshalledReqBody, err := json.Marshal(&reqBody)
	if err != nil {
		return nil, errors.New("failed to marshal json")
	}

	url := requestUrl + "/devices/entities/devices-actions/v2?action_name=hide_host"
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	res, err := requests.SendRequest(ctx, plugin, http.MethodPost, url, consts.DefaultTimeout, headers, nil, marshalledReqBody)

	return res.Body, err
}
