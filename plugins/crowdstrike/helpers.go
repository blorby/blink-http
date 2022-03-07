package crowdstrike

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/implementation/requests"
	"github.com/blinkops/blink-http/plugins/types"
	"github.com/blinkops/blink-sdk/plugin"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

const (
	requestUrlParam   = "REQUEST_URL"
	clientIdParam     = "Client ID"
	clientSecretParam = "Client Secret"
)

func getGetInstalledDevicesParams(request *plugin.ExecuteActionRequest) ([]string, bool, error) {
	params, err := request.GetParameters()
	if err != nil {
		return nil, false, err
	}

	deviceSerials, onlyActiveDevices := params[deviceSerialsParam], params[onlyActiveDevicesParam]
	if deviceSerials == "" {
		return nil, false, errors.New("no device serials provided")
	}
	if onlyActiveDevices == "" {
		return nil, false, errors.New("input for 'return only active devices' not provided")
	}

	onlyActiveDevicesBool, err := strconv.ParseBool(onlyActiveDevices)
	if err != nil {
		return nil, false, errors.New("unable to convert 'return only active devices' to boolean")
	}

	deviceSerials = strings.ReplaceAll(deviceSerials, ", ", ",")
	deviceSerialsList := strings.Split(deviceSerials, ",")

	return deviceSerialsList, onlyActiveDevicesBool, nil
}

func getDeviceIdBySerial(ctx *plugin.ActionContext, plugin types.Plugin, requestUrl string, timeout int32, serial string) (string, error) {
	url := fmt.Sprintf("%s/devices/queries/devices/v1?filter=serial_number:'%s'", requestUrl, url.QueryEscape(serial))
	res, err := requests.SendRequest(ctx, plugin, http.MethodGet, url, consts.DefaultTimeout, nil, nil, nil)
	if err != nil {
		return "", err
	}

	var resJson QueryDevicesByFilterResponse
	err = json.Unmarshal(res.Body, &resJson)
	if err != nil {
		return "", errors.New("failed to unmarshal response json")
	}

	if len(resJson.Resources) > 0 {
		return resJson.Resources[0], nil
	}

	return "", nil
}

func isDeviceActive(ctx *plugin.ActionContext, plugin types.Plugin, requestUrl string, timeout int32, deviceId string) (bool, error) {
	url := fmt.Sprintf("%s/devices/entities/devices/v1?ids=%s", requestUrl, url.QueryEscape(deviceId))
	res, err := requests.SendRequest(ctx, plugin, http.MethodGet, url, consts.DefaultTimeout, nil, nil, nil)
	if err != nil {
		return false, err
	}

	var resJson GetDeviceResponse
	err = json.Unmarshal(res.Body, &resJson)
	if err != nil {
		return false, errors.New("failed to unmarshal response json")
	}

	if len(resJson.Resources) > 0 && resJson.Resources[0].Status != "suppressed" {
		return true, nil
	}

	return false, nil
}

func performGetInstalledDevices(ctx *plugin.ActionContext, plugin types.Plugin, request *plugin.ExecuteActionRequest, deviceSerials []string, onlyActiveDevices bool) (chan string, error) {
	requestUrl, err := requests.GetRequestUrl(ctx, PluginName)
	if err != nil {
		return nil, errors.New("no request url provided")
	}

	activeSerialsChan := make(chan string, len(deviceSerials))
	var errorToReturn error

	defer close(activeSerialsChan)

	var wg sync.WaitGroup
	for _, serial := range deviceSerials {
		serialVal := serial
		wg.Add(1)
		go func() {
			defer wg.Done()
			deviceId, err := getDeviceIdBySerial(ctx, plugin, requestUrl, request.Timeout, serialVal)
			if err != nil {
				errorToReturn = err
				return
			}

			// if device is not installed
			if deviceId == "" {
				return
			}

			if onlyActiveDevices {
				// add serial to list only if device is active
				var deviceActive bool
				if deviceActive, err = isDeviceActive(ctx, plugin, requestUrl, request.Timeout, deviceId); err != nil {
					errorToReturn = err
					return
				}

				if deviceActive {
					activeSerialsChan <- serialVal
				}
			} else {
				activeSerialsChan <- serialVal
			}
		}()
	}
	wg.Wait()

	return activeSerialsChan, err
}
