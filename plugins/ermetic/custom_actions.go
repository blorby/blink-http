package ermetic

import (
	"encoding/json"
	"errors"
	"github.com/blinkops/blink-http/plugins/types"
	"github.com/blinkops/blink-sdk/plugin"
)

const (
	requestUrl     = "https://us.app.ermetic.com/api/graph"
	queryParam     = "Query"
	variablesParam = "Variables"
	pageParam      = "Page"
	pageSizeParam  = "Page Size"
	skipParam      = "Skip"
	takeParam      = "Take"
)

func listFindings(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	graphqlVariables, err := getGraphQLVars(request)
	if err != nil {
		return nil, err
	}

	return executeGraphQL(ctx, plugin, listFindingsQuery, graphqlVariables)
}

func listEntities(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	graphqlVariables, err := getGraphQLVars(request)
	if err != nil {
		return nil, err
	}

	return executeGraphQL(ctx, plugin, listEntitiesQuery, graphqlVariables)
}

func query(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin) ([]byte, error) {
	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}

	variables := map[string]interface{}{}
	variablesStr, ok := params[variablesParam]
	if ok {
		err = json.Unmarshal([]byte(variablesStr), &variables)
		if err != nil {
			return nil, errors.New("failed to marshal variables json")
		}
	} else {
		variables = nil
	}

	return executeGraphQL(ctx, plugin, params[queryParam], variables)
}
