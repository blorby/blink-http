package implementation

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/blinkops/blink-sdk/plugin"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const (
	urlKey         = "URL"
	contentTypeKey = "Content type"
	headersKey     = "Headers"
	cookiesKey     = "Cookies"
	bodyKey        = "Body"
	usernameKey    = "username"
	passwordKey    = "password"
	tokenKey       = "token"
	apiAddressKey  = "API Address"
	basicAuthKey   = "basic-auth"
	bearerAuthKey  = "bearer-token"
	apiTokenKey    = "apikey-auth"
)

func handleBasicAuth(ctx *plugin.ActionContext, req *http.Request) error {
	credentials, _ := ctx.GetCredentials(basicAuthKey)
	if credentials == nil {
		return nil
	}
	if err := validateURL(credentials, req.URL); err != nil {
		return err
	}
	uname, ok := credentials[usernameKey].(string)
	if !ok {
		return errors.New("basic-auth connection does not contain a username attribute or the username is not a string")
	}
	password, ok := credentials[passwordKey].(string)
	if !ok {
		return errors.New("basic-auth connection does not contain a password attribute or the password is not a string")
	}
	req.Header.Add("Authorization", "Basic "+basicAuth(uname, password))
	return nil
}

func handleBearerToken(ctx *plugin.ActionContext, req *http.Request) error {
	credentials, _ := ctx.GetCredentials(bearerAuthKey)
	if credentials == nil {
		return nil
	}
	if err := validateURL(credentials, req.URL); err != nil {
		return err
	}
	token, ok := credentials[tokenKey].(string)
	if !ok {
		return errors.New("bearer-token connection does not contain a token attribute or the token is not a string")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	return nil
}

func handleApiKeyAuth(ctx *plugin.ActionContext, req *http.Request) error {
	credentials, _ := ctx.GetCredentials(apiTokenKey)
	if credentials == nil {
		return nil
	}
	if err := validateURL(credentials, req.URL); err != nil {
		return err
	}
	h1, ok1 := credentials["Header 1"].(string)
	v1, ok2 := credentials["Value 1"].(string)
	if ok1 && ok2 {
		req.Header.Add(h1, v1)
	}
	h2, ok1 := credentials["Header 2"].(string)
	v2, ok2 := credentials["Value 2"].(string)
	if ok1 && ok2 {
		req.Header.Add(h2, v2)
	}
	h3, ok1 := credentials["Header 3"].(string)
	v3, ok2 := credentials["Value 3"].(string)
	if ok1 && ok2 {
		req.Header.Add(h3, v3)
	}
	return nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func readBody(responseBody io.ReadCloser) ([]byte, error) {
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

func createResponse(response *http.Response, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}

	if response == nil {
		return nil, errors.New("response has not been provided")
	}

	body, err := readBody(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusBadRequest {
		return body, errors.New(fmt.Sprintf("status: %v", response.StatusCode))
	}

	return body, nil
}

func sendRequest(ctx *plugin.ActionContext, method string, urlString string, timeout int32, headers map[string]string, cookies map[string]string, data []byte) ([]byte, error) {
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

	if err = fixURL(request.URL); err != nil {
		return nil, err
	}

	for name, value := range headers {
		request.Header.Set(name, value)
	}

	if err = handleBasicAuth(ctx, request); err != nil {
		return nil, err
	}
	if err = handleBearerToken(ctx, request); err != nil {
		return nil, err
	}
	if err = handleApiKeyAuth(ctx, request); err != nil {
		return nil, err
	}
	response, err := client.Do(request)

	return createResponse(response, err)
}

func validateURL(connection map[string]interface{}, requestedURL *url.URL) error {
	apiAddressString, ok := connection[apiAddressKey].(string)
	if !ok {
		return errors.New("connection does not contain a api address attribute or the api address is not convertable to string")
	}
	apiAddressString = strings.Replace(apiAddressString, "www.", "", 1)
	apiAddress, err := url.Parse(apiAddressString)
	if err != nil {
		return err
	}
	err = fixURL(apiAddress)
	if err != nil {
		return err
	}
	requestedURL, err = url.Parse(strings.Replace(requestedURL.String(), "www.", "", 1))
	if err != nil {
		return err
	}
	if !strings.HasPrefix(requestedURL.Host+requestedURL.Path, apiAddress.Host+apiAddress.Path) {
		return errors.New("the requested url's host/path does not match the host/path defined in the connection. this is not allowed in order to prevent sending credentials to unwanted hosts/paths")
	}
	return nil
}

// fixURL adds https if the user forgot to include the scheme and trims unnecessary trailing slashes
func fixURL(u *url.URL) error {
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	val, err := url.Parse(strings.TrimSuffix(u.String(), "/"))
	u = val

	return err
}

func getHeaders(contentType string, headers string) map[string]string {
	headerMap := parseStringToMap(headers, ":")
	headerMap["Content-Type"] = contentType

	return headerMap
}

func parseStringToMap(value string, delimiter string) map[string]string {
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

func executeHTTPGetAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error) {
	return executeCoreHTTPAction(ctx, http.MethodGet, request)
}

func executeHTTPPostAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error) {
	return executeCoreHTTPAction(ctx, http.MethodPost, request)
}

func executeHTTPPutAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error) {
	return executeCoreHTTPAction(ctx, http.MethodPut, request)
}

func executeHTTPDeleteAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error) {
	return executeCoreHTTPAction(ctx, http.MethodDelete, request)
}

func executeHTTPPatchAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error) {
	return executeCoreHTTPAction(ctx, http.MethodPatch, request)
}

func executeCoreHTTPAction(ctx *plugin.ActionContext, method string, request *plugin.ExecuteActionRequest) ([]byte, error) {
	providedUrl, ok := request.Parameters[urlKey]
	if !ok {
		return nil, errors.New("no url provided for execution")
	}

	contentType, ok := request.Parameters[contentTypeKey]
	if !ok {
		contentType = "application/x-www-form-urlencoded"
	}

	headers, ok := request.Parameters[headersKey]
	if !ok {
		headers = ""
	}

	cookies, ok := request.Parameters[cookiesKey]
	if !ok {
		cookies = ""
	}

	body, ok := request.Parameters[bodyKey]
	if !ok {
		body = ""
	}

	headerMap := getHeaders(contentType, headers)
	cookieMap := parseStringToMap(cookies, "=")

	return sendRequest(ctx, method, providedUrl, request.Timeout, headerMap, cookieMap, []byte(body))
}
