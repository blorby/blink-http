package jira

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/implementation/requests"
	"github.com/blinkops/blink-http/plugins/types"
	"net/http"
	"strings"

	"github.com/blinkops/blink-sdk/plugin"
)

func extractKeys(jsonstr string) string {
	jsonstr = strings.TrimPrefix(jsonstr, "{")
	jsonstr = strings.TrimSuffix(jsonstr, "}")

	return "," + jsonstr
}

func getUserIDByEmail(ctx *plugin.ActionContext, plugin types.Plugin, requestUrl string, userEmail string, timeout int32) (string, error) {
	url := fmt.Sprintf("%s/rest/api/3/user/search?query=%s", requestUrl, userEmail)
	res, err := requests.SendRequest(ctx, plugin, http.MethodGet, url, consts.DefaultTimeout, nil, nil, nil)
	if err != nil {
		return "", err
	}

	var users []struct {
		AccountID string `json:"accountId"`
	}
	err = json.Unmarshal(res.Body, &users)
	if err != nil {
		return "", err
	}

	if len(users) == 0 {
		return "", errors.New("user not found")
	}
	if len(users) > 1 {
		return "", errors.New("the user email input matched more than one user")
	}
	return users[0].AccountID, nil
}

func getIssueTypeIDByName(ctx *plugin.ActionContext, plugin types.Plugin, requestUrl string, issueTypeName string, projectKey string, timeout int32) (string, error) {
	projectID, err := getProjectIDByKey(ctx, plugin, requestUrl, projectKey, timeout)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/rest/api/3/issuetype/project?projectId=%s", requestUrl, projectID)
	res, err := requests.SendRequest(ctx, plugin, http.MethodGet, url, consts.DefaultTimeout, nil, nil, nil)
	if err != nil {
		return "", err
	}

	var issueTypes []struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	}
	err = json.Unmarshal(res.Body, &issueTypes)
	if err != nil {
		return "", err
	}

	for _, issue := range issueTypes {
		if issue.Name == issueTypeName {
			return issue.ID, nil
		}
	}
	return "", errors.New("no issue with the name " + issueTypeName + " was found in this project")
}

func getProjectIDByKey(ctx *plugin.ActionContext, plugin types.Plugin, requestUrl string, projectKey string, timeout int32) (string, error) {
	url := fmt.Sprintf("%s/rest/api/3/project/%s", requestUrl, projectKey)
	res, err := requests.SendRequest(ctx, plugin, http.MethodGet, url, consts.DefaultTimeout, nil, nil, nil)
	if err != nil {
		return "", err
	}

	var project struct {
		ID string `json:"id"`
	}
	err = json.Unmarshal(res.Body, &project)
	if err != nil {
		return "", err
	}

	if project.ID != "" {
		return project.ID, nil
	}
	return "", fmt.Errorf("no project with the key " + projectKey + " found")
}

func getFieldIDByName(ctx *plugin.ActionContext, plugin types.Plugin, requestUrl string, fieldName string, timeout int32) (string, error) {
	fields, err := getFields(ctx, plugin, requestUrl, timeout)
	if err != nil {
		return "", err
	}

	val, ok := fields[fieldName]
	if !ok {
		return "", fmt.Errorf("no such field: %s", fieldName)
	}
	return val, nil
}

func getFields(ctx *plugin.ActionContext, plugin types.Plugin, requestUrl string, timeout int32) (map[string]string, error) {
	url := fmt.Sprintf("%s/rest/api/3/field/search?type=custom", requestUrl)
	res, err := requests.SendRequest(ctx, plugin, http.MethodGet, url, consts.DefaultTimeout, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	obj := Search{}
	err = json.Unmarshal(res.Body, &obj)
	if err != nil {
		return nil, err
	}

	fields := map[string]string{}
	for _, value := range obj.Values {
		fields[value.Name] = value.Id
	}

	return fields, nil
}
