package implementation

import (
	"encoding/json"
	"errors"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/implementation/requests"
	"github.com/blinkops/blink-http/plugins/types"
	"github.com/blinkops/blink-sdk/plugin"
	"net/http"
)

func executeHTTPGetAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	return executeCoreHTTPAction(ctx, http.MethodGet, request, plugin)
}

func executeHTTPPostAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	return executeCoreHTTPAction(ctx, http.MethodPost, request, plugin)
}

func executeHTTPPutAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	return executeCoreHTTPAction(ctx, http.MethodPut, request, plugin)
}

func executeHTTPDeleteAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	return executeCoreHTTPAction(ctx, http.MethodDelete, request, plugin)
}

func executeHTTPPatchAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	return executeCoreHTTPAction(ctx, http.MethodPatch, request, plugin)
}

func executeCoreHTTPAction(ctx *plugin.ActionContext, method string, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
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

	return requests.SendRequest(ctx, plugin, method, providedUrl, request.Timeout, headerMap, cookieMap, []byte(body))
}

func executeGraphQL(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	providedUrl, ok := request.Parameters[consts.UrlKey]
	if !ok {
		return nil, errors.New("no url provided for execution")
	}

	query, ok := request.Parameters[consts.QueryKey]
	if !ok {
		query = ""
	}

	variables, ok := request.Parameters[consts.VariablesKey]
	if !ok {
		variables = ""
	}

	body, err := json.Marshal(map[string]string{"query": query, "variables": variables})
	if err != nil {
		return nil, err
	}

	headerMap := map[string]string{"Content-Type": "application/json"}

	return requests.SendRequest(ctx, plugin, http.MethodPost, providedUrl, request.Timeout, headerMap, nil, body)
}
