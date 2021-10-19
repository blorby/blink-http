package implementation

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/blinkops/blink-sdk/plugin"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	urlKey         = "url"
	timeoutKey     = "timeout"
	contentTypeKey = "contentType"
	headersKey     = "headers"
	cookiesKey     = "cookies"
	bodyKey        = "body"
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

func sendRequest(method string, urlAsString string, timeout string, headers map[string]string, cookies map[string]string, data []byte) ([]byte, error) {
	requestBody := bytes.NewBuffer(data)

	timeoutAsNumber, err := strconv.ParseInt(timeout, 10, 64)
	if err != nil {
		return nil, err
	}

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

	parsedUrl, err := url.Parse(urlAsString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse request url, error: %v", err)
	}
	cookieJar.SetCookies(parsedUrl, cookiesList)

	// Create new http client with predefined options
	client := &http.Client{
		Jar:     cookieJar,
		Timeout: time.Second * time.Duration(timeoutAsNumber),
	}

	request, err := http.NewRequest(method, urlAsString, requestBody)
	if err != nil {
		return nil, err
	}

	for name, value := range headers {
		request.Header.Set(name, value)
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
	return executeCoreHTTPAction(http.MethodGet, ctx, request)
}

func executeHTTPPostAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error) {
	return executeCoreHTTPAction(http.MethodPost, ctx, request)
}

func executeHTTPPutAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error) {
	return executeCoreHTTPAction(http.MethodPut, ctx, request)
}
func executeHTTPDeleteAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error) {
	return executeCoreHTTPAction(http.MethodDelete, ctx, request)
}

func executeCoreHTTPAction(method string, _ *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error) {
	providedUrl, ok := request.Parameters[urlKey]
	if !ok {
		return nil, errors.New("no url provided for execution")
	}

	timeout, ok := request.Parameters[timeoutKey]
	if !ok {
		timeout = "60"
	}

	contentType, ok := request.Parameters[contentTypeKey]
	if !ok {
		return nil, errors.New("no content-type provided for execution")
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

	return sendRequest(method, providedUrl, timeout, headerMap, cookieMap, []byte(body))
}
