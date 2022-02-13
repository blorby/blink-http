package implementation

import (
	"bytes"
	"errors"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/blinkops/blink-sdk/plugin"
	log "github.com/sirupsen/logrus"
)

const (
	urlKey         = "url"
	queryKey      = "query"
	contentTypeKey = "contentType"
	headersKey     = "headers"
	cookiesKey     = "cookies"
	bodyKey        = "body"

	usernameKey   = "username"
	passwordKey   = "password"
	tokenKey      = "token"
	apiAddressKey = "API Address"
	basicAuthKey  = "basic-auth"
	bearerAuthKey = "bearer-token"
	apiTokenKey   = "apikey-auth"
)


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

	for name, value := range headers {
		request.Header.Set(name, value)
	}

	if err = handleAuthentication(ctx, request); err != nil {
		return nil, err
	}

	response, err := client.Do(request)

	return createResponse(response, err)
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

func executeGraphQL(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error) {
	providedUrl, ok := request.Parameters[urlKey]
	if !ok {
		return nil, errors.New("no url provided for execution")
	}

	query, ok := request.Parameters[queryKey]
	if !ok {
		query = ""
	}

	body, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return nil, err
	}

	headerMap := map[string]string{"Content-Type": "application/json"}

	return sendRequest(ctx, http.MethodPost, providedUrl, request.Timeout, headerMap, nil, body)
}