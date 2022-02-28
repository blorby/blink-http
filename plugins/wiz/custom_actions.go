package wiz

import (
	"encoding/json"
	"errors"
	"github.com/blinkops/blink-http/plugins/types"
	"github.com/blinkops/blink-sdk/plugin"
	"strconv"
)

const (
	projectNameParam                 = "ProjectName"
	limitParam                       = "Limit"
	offsetParam                      = "Offset"
	filterByParam                    = "FilterBy"
	orderByParam                     = "OrderBy"
)

type listCloudConfigurationRulesVariables struct {
	ProjectIds []string `json:"projectId"`
	Limit      int      `json:"first"`
	Offset     string   `json:"after,omitempty"`
	FilterBy   string   `json:"filterBy,omitempty"`
	OrderBy    string   `json:"orderBy,omitempty"`
}

type listControlsVariables struct {
	Limit                   int    `json:"first"`
	Offset                  string `json:"after,omitempty"`
	FilterBy                string `json:"filterBy,omitempty"`
	IssueAnalyticsSelection struct {
		ProjectIds []string `json:"project"`
	} `json:"issue_analytics_selection,omitempty"`
}

type getProjectResponse struct {
	Data struct {
		Projects struct {
			Nodes []struct {
				Id   string `json:"id"`
				Name string `json:"name"`
			} `json:"nodes"`
		} `json:"projects"`
	} `json:"data"`
}

func listCloudConfigurationRules(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}

	projectName, limit, offset, filterBy, orderBy := params[projectNameParam], params[limitParam], params[offsetParam], params[filterByParam], params[orderByParam]
	if projectName == "" {
		return nil, errors.New("project name not provided")
	}
	if limit == "" {
		return nil, errors.New("limit not provided")
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return nil, errors.New("unable to covert limit to integer")
	}

	projectId, err := getProjectIdByName(ctx, request, plugin, projectName)
	if err != nil {
		return nil, err
	}

	variablesStruct := listCloudConfigurationRulesVariables{
		ProjectIds: []string{projectId},
		Limit:      limitInt,
		Offset:     offset,
		FilterBy:   filterBy,
		OrderBy:    orderBy,
	}
	variables, err := json.Marshal(variablesStruct)
	if err != nil {
		return nil, errors.New("failed to marshal variables")
	}

	return execQuery(ctx, request, plugin, listCloudConfigurationRulesQuery, variables)
}

func listControls(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}

	limit, offset, projectName, filterBy := params[limitParam], params[offsetParam], params[projectNameParam], params[filterByParam]
	if limit == "" {
		return nil, errors.New("limit not provided")
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return nil, errors.New("unable to covert limit to integer")
	}

	var projectId string
	if projectName != "" {
		projectId, err = getProjectIdByName(ctx, request, plugin, projectName)
		if err != nil {
			return nil, err
		}
	}

	variablesStruct := listControlsVariables{
		Limit:    limitInt,
		Offset:   offset,
		FilterBy: filterBy,
		IssueAnalyticsSelection: struct {
			ProjectIds []string `json:"project"`
		}{ProjectIds: []string{projectId}},
	}
	variables, err := json.Marshal(variablesStruct)
	if err != nil {
		return nil, errors.New("failed to marshal variables")
	}

	return execQuery(ctx, request, plugin, listControlsQuery, variables)
}
