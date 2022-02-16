package requests

import (
	"errors"
	"fmt"
	"github.com/blinkops/blink-http/plugins"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func ReadBody(responseBody io.ReadCloser) ([]byte, error) {
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			log.Debugf("failed to close responseBody reader, error: %v", err)
		}
	}(responseBody)

	body, err := ioutil.ReadAll(responseBody)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func CreateResponse(response *http.Response, err error, pluginName string) ([]byte, error) {
	if err != nil {
		return nil, err
	}

	if response == nil {
		return nil, errors.New("response has not been provided")
	}

	body, err := ReadBody(response.Body)
	if err != nil {
		return nil, err
	}
	if plugin, ok := plugins.Plugins[pluginName]; ok {
		if pluginWithValidation, ok := plugin.(plugins.PluginWithValidation); ok {
			return pluginWithValidation.ValidateResponse(response.StatusCode, body)
		}
	}
	return ValidateResponse(response.StatusCode, body)
}

func ValidateResponse(statusCode int, body []byte) ([]byte, error) {
	if statusCode < http.StatusOK || statusCode >= http.StatusBadRequest {
		return body, errors.New(fmt.Sprintf("status: %v", statusCode))
	}

	return body, nil
}

func GetHeaders(contentType string, headers string) map[string]string {
	headerMap := ParseStringToMap(headers, ":")
	headerMap["Content-Type"] = contentType

	return headerMap
}

func ParseStringToMap(value string, delimiter string) map[string]string {
	stringMap := make(map[string]string)

	split := strings.Split(value, "\n")
	for _, currentParameter := range split {
		if strings.Contains(currentParameter, delimiter) {
			currentHeaderSplit := strings.Split(currentParameter, delimiter)
			parameterKey, parameterValue := currentHeaderSplit[0], strings.TrimSpace(currentHeaderSplit[1])

			stringMap[parameterKey] = parameterValue
		}
	}
	return stringMap
}