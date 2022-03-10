package jira

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/implementation/requests"
	"github.com/blinkops/blink-http/plugins/types"
	"github.com/blinkops/blink-sdk/plugin"
	"net/http"
	"sync"
)

const (
	pluginName         = "jira"
	userEmailParam     = "User Email"
	issueIdOrKeyParam  = "Issue ID Or Key"
	contentParam       = "Content"
	summaryParam       = "Summary"
	projectKeyParam    = "Project Key"
	descriptionParam   = "Description"
	issueTypeParam     = "Issue Type"
	assigneeEmailParam = "Assignee Email"
	customFieldsParam  = "Custom Fields"
)

func removeUser(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	requestUrl, err := requests.GetRequestUrl(ctx, pluginName)
	if err != nil {
		return nil, errors.New(consts.RequestUrlMissing)
	}

	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}

	userMail := params[userEmailParam]
	userID, err := getUserIDByEmail(ctx, plugin, requestUrl, userMail, request.Timeout)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/rest/api/3/user?accountId=%s", requestUrl, userID)
	headers := map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
	res, err := requests.SendRequest(ctx, plugin, http.MethodPost, url, consts.DefaultTimeout, headers, nil, nil)

	return res.Body, err
}

func addComment(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	requestUrl, err := requests.GetRequestUrl(ctx, pluginName)
	if err != nil {
		return nil, errors.New(consts.RequestUrlMissing)
	}

	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}
	issueID := params[issueIdOrKeyParam]
	content := params[contentParam]
	reqBody := json.RawMessage(`{
  "body": {
    "type": "doc",
    "version": 1,
    "content": [
      {
        "type": "paragraph",
        "content": [
          {
            "text": "` + content + `",
            "type": "text"
          }
        ]
      }
    ]
  }
}`)
	marshalledReqBody, err := json.Marshal(&reqBody)
	if err != nil {
		return nil, errors.New("failed to marshal request body")
	}

	url := fmt.Sprintf("%s/rest/api/3/issue/%s/comment", requestUrl, issueID)
	headers := map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
	res, err := requests.SendRequest(ctx, plugin, http.MethodPost, url, consts.DefaultTimeout, headers, nil, marshalledReqBody)

	return res.Body, err
}

func listCustomFields(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	requestUrl, err := requests.GetRequestUrl(ctx, pluginName)
	if err != nil {
		return nil, errors.New(consts.RequestUrlMissing)
	}

	val, err := getFields(ctx, plugin, requestUrl, request.Timeout)
	if err != nil {
		return nil, err
	}

	var names []string
	for name := range val {
		names = append(names, name)
	}

	result, err := json.Marshal(names)
	if err != nil {
		return nil, errors.New("failed to marshal custom field names")
	}

	return result, nil
}

func createIssue(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	requestUrl, err := requests.GetRequestUrl(ctx, pluginName)
	if err != nil {
		return nil, errors.New(consts.RequestUrlMissing)
	}

	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}
	summary, project, description, issueType := params[summaryParam], params[projectKeyParam], params[descriptionParam], params[issueTypeParam]
	assignee, hasAssignee := params[assigneeEmailParam]

	var wg sync.WaitGroup
	var assigneeID, issueTypeID string
	if hasAssignee {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assigneeID, err = getUserIDByEmail(ctx, plugin, requestUrl, assignee, request.Timeout)
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		issueTypeID, err = getIssueTypeIDByName(ctx, plugin, requestUrl, issueType, project, request.Timeout)
	}()
	wg.Wait()
	if err != nil {
		return nil, err
	}

	var assigneeField string
	if hasAssignee {
		assigneeField = `,
    "assignee": {
      "id": "` + assigneeID + `"
    }`
	}

	customFields := params[customFieldsParam]

	if customFields != "" {
		userJson := map[string]interface{}{}
		err = json.Unmarshal([]byte(customFields), &userJson)
		if err != nil {
			return nil, errors.New("can't parse custom fields")
		}

		updatedJson := map[string]interface{}{}

		for name, val := range userJson {
			aa, err := getFieldIDByName(ctx, plugin, requestUrl, name, request.Timeout)
			if err != nil {
				return nil, err
			}
			updatedJson[aa] = val
		}
		marshaledJson, err := json.Marshal(updatedJson)
		if err != nil {
			return nil, errors.New("cant marshaledJson custom fields")
		}
		customFields = extractKeys(string(marshaledJson))
	}

	reqBody := json.RawMessage(`{
  "fields": {
    "summary": "` + summary + `",
    "issuetype": {
      "id": "` + issueTypeID + `"
    },
    "project": {
      "key": "` + project + `"
    },
    "description": {
      "type": "doc",
      "version": 1,
      "content": [
        {
          "type": "paragraph",
          "content": [
            {
              "text": "` + description + `",
              "type": "text"
            }
          ]
        }
      ]
    }` + assigneeField +
		customFields + `
  }
}`)
	marshalledReqBody, err := json.Marshal(&reqBody)
	if err != nil {
		return nil, errors.New("failed to marshal request body")
	}

	url := fmt.Sprintf("%s/rest/api/3/issue", requestUrl)
	headers := map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
	res, err := requests.SendRequest(ctx, plugin, http.MethodPost, url, consts.DefaultTimeout, headers, nil, marshalledReqBody)

	return res.Body, err
}
