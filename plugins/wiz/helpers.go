package wiz

import (
	"encoding/json"
	"errors"
	"github.com/blinkops/blink-http/implementation/requests"
	"github.com/blinkops/blink-http/plugins/types"
	"github.com/blinkops/blink-sdk/plugin"
	"net/http"
)

func execQuery(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin, query string, variables []byte) ([]byte, error) {
	requestUrl, err := getRequestUrl(ctx)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(map[string]string{"query": query, "variables": string(variables)})
	if err != nil {
		return nil, err
	}

	headerMap := map[string]string{"Content-Type": "application/json"}

	return requests.SendRequest(ctx, plugin, http.MethodPost, requestUrl, request.Timeout, headerMap, nil, body)
}

func getRequestUrl(ctx *plugin.ActionContext) (string, error) {
	connection, err := ctx.GetCredentials("wiz")
	if err != nil {
		return "", err
	}

	requestUrl, ok := connection[requestUrlParam]
	if !ok {
		return "", errors.New("no request url provided")
	}

	return requestUrl, nil
}

func getProjectIdByName(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin, projectName string) (string, error) {
	variables := `{"first":1, "search":"`+ projectName +`"}`

	resp, err := execQuery(ctx, request, plugin, listProjectsQuery, []byte(variables))
	if err != nil {
		return "", err
	}

	var respJson getProjectResponse
	err = json.Unmarshal(resp, &respJson)
	if err != nil {
		return "", errors.New("failed to unmarshal project json")
	}

	if len(respJson.Data.Projects.Nodes) != 1 || respJson.Data.Projects.Nodes[0].Name != projectName {
		return "", errors.New("project not found")
	}

	return respJson.Data.Projects.Nodes[0].Id, nil
}
