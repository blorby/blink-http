package implementation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	conn "github.com/blinkops/blink-http/implementation/connections"
	"github.com/blinkops/blink-http/implementation/requests"
	"github.com/blinkops/blink-sdk/plugin"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

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
	providedUrl, ok := request.Parameters[consts.UrlKey]
	if !ok {
		return nil, errors.New("no url provided for execution")
	}

	contentType, ok := request.Parameters[consts.ContentTypeKey]
	if !ok {
		contentType = "application/x-www-form-urlencoded"
	}

	headers, ok := request.Parameters[consts.HeadersKey]
	if !ok {
		headers = ""
	}

	cookies, ok := request.Parameters[consts.CookiesKey]
	if !ok {
		cookies = ""
	}

	body, ok := request.Parameters[consts.BodyKey]
	if !ok {
		body = ""
	}

	headerMap := requests.GetHeaders(contentType, headers)
	cookieMap := requests.ParseStringToMap(cookies, "=")

	return SendRequest(ctx, method, providedUrl, request.Timeout, headerMap, cookieMap, []byte(body))
}

func executeGraphQL(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error) {
	providedUrl, ok := request.Parameters[consts.UrlKey]
	if !ok {
		return nil, errors.New("no url provided for execution")
	}

	query, ok := request.Parameters[consts.QueryKey]
	if !ok {
		query = ""
	}

	body, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return nil, err
	}

	headerMap := map[string]string{"Content-Type": "application/json"}

	return SendRequest(ctx, http.MethodPost, providedUrl, request.Timeout, headerMap, nil, body)
}

func SendRequest(ctx *plugin.ActionContext, method string, urlString string, timeout int32, headers map[string]string, cookies map[string]string, data []byte) ([]byte, error) {
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

	if err = conn.HandleAuth(ctx, request); err != nil {
		return nil, err
	}

	response, err := client.Do(request)

	return requests.CreateResponse(response, err)
}