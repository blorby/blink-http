package implementation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/implementation/requests"
	"github.com/blinkops/blink-http/plugins"
	"github.com/blinkops/blink-sdk/plugin"
	"github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
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

	var pluginName string
	for connName, connInstance := range ctx.GetAllConnections() {
		if err = handleAuth(connName, connInstance, request); err != nil {
			return nil, err
		}
		pluginName = connName
	}

	response, err := client.Do(request)

	return requests.CreateResponse(response, err, pluginName)
}

func handleAuth(connName string, connInstance *connections.ConnectionInstance, req *http.Request) error {
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
		if integration, ok := plugins.Plugins[connName]; ok {
			if err := integration.HandleAuth(req, connInstance.Data); err != nil {
				return err
			}
		}
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
