package plugins

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/plugins/types"
	"github.com/blinkops/blink-sdk/plugin"
	"github.com/blinkops/blink-sdk/plugin/connections"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

func SendRequest(ctx *plugin.ActionContext, method string, urlString string, timeout int32, headers map[string]string, cookies map[string]string, data []byte, plugin types.Plugin) ([]byte, error) {
	requestBody := bytes.NewBuffer(data)

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar, error: %v", err)
	}

	var cookiesList []*http.Cookie
	for name, value := range cookies {
		cookiesList = append(cookiesList, &http.Cookie{
			Name:  name,
			Value: value,
		})
	}

	parsedUrl, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse request url, error: %v", err)
	}
	cookieJar.SetCookies(parsedUrl, cookiesList)

	// Create new http client with predefined options
	client := &http.Client{
		Jar:     cookieJar,
		Timeout: time.Second * time.Duration(timeout),
	}

	request, err := http.NewRequest(method, urlString, requestBody)
	if err != nil {
		return nil, err
	}

	for name, value := range headers {
		request.Header.Set(name, value)
	}

	for connName, connInstance := range ctx.GetAllConnections() {
		if err = handleAuth(connName, connInstance, request, plugin); err != nil {
			return nil, err
		}
	}

	response, err := client.Do(request)

	return CreateResponse(response, err, plugin)
}

func handleAuth(connName string, connInstance *connections.ConnectionInstance, req *http.Request, plugin types.Plugin) error {
	if plugin != nil {
		if err := plugin.HandleAuth(req, connInstance.Data); err != nil {
			return err
		}
	}
	switch connName {
	case consts.BasicAuthKey:
		if err := handleBasicAuth(connInstance.Data, req); err != nil {
			return err
		}
	case consts.BearerAuthKey:
		if err := handleBearerToken(connInstance.Data, req); err != nil {
			return err
		}
	case consts.ApiTokenKey:
		if err := handleApiKeyAuth(connInstance.Data, req); err != nil {
			return err
		}
	default:
		return errors.New("invalid connection type")
	}
	return nil
}

func handleBasicAuth(connection map[string]string, req *http.Request) error {
	if err := validateURL(connection, req.URL); err != nil {
		return err
	}
	uname, ok := connection[consts.UsernameKey]
	if !ok {
		return errors.New("basic-auth connection does not contain a username attribute or the username is not a string")
	}
	password, ok := connection[consts.PasswordKey]
	if !ok {
		return errors.New("basic-auth connection does not contain a password attribute or the password is not a string")
	}
	req.Header.Add("Authorization", "Basic "+basicAuth(uname, password))
	return nil
}

func handleBearerToken(connection map[string]string, req *http.Request) error {
	if err := validateURL(connection, req.URL); err != nil {
		return err
	}
	token, ok := connection[consts.TokenKey]
	if !ok {
		return errors.New("bearer-token connection does not contain a token attribute or the token is not a string")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	return nil
}

func handleApiKeyAuth(connection map[string]string, req *http.Request) error {
	if err := validateURL(connection, req.URL); err != nil {
		return err
	}
	for i := 1; i <= 3; i++ {
		h, ok1 := connection["Header "+fmt.Sprint(i)]
		v, ok2 := connection["Value "+fmt.Sprint(i)]
		if ok1 && ok2 {
			req.Header.Add(h, v)
		}
	}
	return nil
}

func validateURL(connection map[string]string, requestedURL *url.URL) error {
	apiAddressString, ok := connection[consts.ApiAddressKey]
	// if there's no api address defined, any url will be allowed
	if !ok {
		return nil
	}
	apiAddressString = strings.Replace(apiAddressString, "www.", "", 1)
	apiAddress, err := url.Parse(apiAddressString)
	if err != nil {
		return err
	}
	requestedURL, err = url.Parse(strings.Replace(requestedURL.String(), "www.", "", 1))
	if err != nil {
		return err
	}
	if !strings.HasPrefix(requestedURL.Host+requestedURL.Path, apiAddress.Host+apiAddress.Path) {
		return errors.New("the requested url's host/path does not match the host/path defined in the connection. this is not allowed in order to prevent sending credentials to unwanted hosts/paths. the allowed host/path is " + apiAddressString)
	}
	return nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

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

func CreateResponse(response *http.Response, err error, plugin types.Plugin) ([]byte, error) {
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
	if plugin != nil {
		if pluginWithValidation, ok := plugin.(types.PluginWithValidation); ok {
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
