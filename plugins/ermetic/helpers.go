package ermetic

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/implementation/requests"
	"github.com/blinkops/blink-http/plugins/types"
	"github.com/blinkops/blink-sdk/plugin"
	"net/http"
	"strconv"
	"strings"
)

func executeGraphQL(ctx *plugin.ActionContext, plugin types.Plugin, query string, variables map[string]interface{}) ([]byte, error) {
	body, err := json.Marshal(map[string]interface{}{"query": query, "variables": variables})
	if err != nil {
		return nil, errors.New("failed to marshal json")
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	res, err := requests.SendRequest(ctx, plugin, http.MethodPost, requestUrl, consts.DefaultTimeout, headers, nil, body)
	if err != nil {
		return nil, err
	}

	return res.Body, err

}

func getGraphQLVars(request *plugin.ExecuteActionRequest) (map[string]interface{}, error) {
	params, err := request.GetParameters()
	if err != nil {
		return nil, err
	}

	pageNumber := params[pageParam]
	delete(params, pageParam)

	pageSize := params[pageSizeParam]
	delete(params, pageSizeParam)

	graphqlVariables := map[string]interface{}{}
	for k, v := range params {
		// normalize the names (graphQL variables cant have a space)
		k = strings.ReplaceAll(k, " ", "")
		v = strings.ReplaceAll(v, " ", "")

		// if we get back a list from the ui, convert it to []string.
		if strings.ContainsAny(v, ",") {
			b := strings.Split(v, ",")
			graphqlVariables[k] = b
		} else {
			graphqlVariables[k] = v
		}
	}

	s, err := strconv.Atoi(pageNumber)
	if err != nil {
		return nil, fmt.Errorf("%s is not a number", pageNumber)
	}

	sz, err := strconv.Atoi(pageSize)
	if err != nil {
		return nil, fmt.Errorf("%s is not a number", pageSize)
	}

	graphqlVariables[skipParam] = calcPageOffset(sz, s)
	graphqlVariables[takeParam] = sz

	// if params are missing fill them with null.
	for k, v := range graphqlVariables {
		if v.(string) == "" {
			graphqlVariables[k] = nil
		}
	}

	return graphqlVariables, nil
}

// calcPageOffset given a page size and a page number the function calculates
// the $take and the $skip variables.
// for page size 100 and page number 1
// take 100 and skip is 0
// for 200 take is 100 and skip is 100
func calcPageOffset(pageSize, pageNumber int) int {
	return (pageSize * pageNumber) - pageSize
}
