package openapi

import (
	"encoding/json"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	ops "github.com/blinkops/blink-http/openapi/operations"
	"github.com/blinkops/blink-sdk/plugin"
	"github.com/blinkops/blink-sdk/plugin/connections"
	"github.com/getkin/kin-openapi/openapi3"
	"strings"
	"unicode"
)

type NamedActionsGroup struct {
	Name                  string                            `yaml:"name"`
	Connections           map[string]connections.Connection `yaml:"connection_types"`
	IconUri               string                            `yaml:"icon_uri"`
	IsConnectionOptional  bool                              `yaml:"is_connection_optional"`
	Actions               []NamedAction                     `yaml:"actions"`
}

type NamedAction struct {
	Name                   string                            `yaml:"name"`
	DisplayName            string                            `yaml:"display_name"`
	Description            string                            `yaml:"description"`
	Enabled                bool                              `yaml:"enabled"`
	Parameters             map[string]plugin.ActionParameter `yaml:"parameters"`
	ActionToRun            string                            `yaml:"action_to_run"`
	ActionToRunParamValues map[string]string                 `yaml:"action_to_run_param_values"`
	IsConnectionOptional   string                            `yaml:"is_connection_optional"`
}

func (o ParsedOpenApi) getNamedActionsGroup() NamedActionsGroup {
	var namedActions []NamedAction
	for _, action := range o.actions {
		if len(o.mask.Actions) > 0 && o.mask.GetAction(action.Name) == nil {
			continue
		}
		opDefinition := o.operations[action.Name]
		namedActions = append(namedActions, NamedAction{
			Name:                   action.Name,
			DisplayName:            getDisplayName(action.Name),
			Description:            action.Description,
			Enabled:                true,
			Parameters:             o.getActionParams(action),
			ActionToRun:            o.getActionToRun(opDefinition),
			ActionToRunParamValues: o.getActionParamValues(action, opDefinition),
		})
	}
	group := NamedActionsGroup{
		Name:                  o.mask.Name,
		Connections:           o.mask.Connections,
		IconUri:               o.mask.IconUri,
		IsConnectionOptional:  o.mask.IsConnectionOptional,
		Actions:               namedActions,
	}
	return group
}

func (o ParsedOpenApi) getActionToRun(opDefinition *ops.OperationDefinition) string {
	return fmt.Sprintf("http.%s", strings.ToLower(opDefinition.Method))
}

func (o ParsedOpenApi) getActionParams(action plugin.Action) map[string]plugin.ActionParameter {
	params := map[string]plugin.ActionParameter{}
	for paramName, param := range action.Parameters {
		maskedAct := o.mask.Actions[action.Name]
		if maskedParam, ok := maskedAct.Parameters[paramName]; ok {
			param.DisplayName = maskedParam.Alias

			if maskedParam.Index != 0 {
				param.Index = maskedParam.Index
			}

			if param.Type == "" { // if it's an unrecognized type, let the user put whatever he wants to
				param.Type = consts.TypeJson
			}

			params[paramName] = param
		}
	}
	return params
}

func (o ParsedOpenApi) getActionParamValues(action plugin.Action, opDefinition *ops.OperationDefinition) map[string]string {
	paramValues := map[string]string{}

	paramValues["url"] = o.getActionUrlTemplate(opDefinition)

	body := getBodyTemplate(action, opDefinition.GetDefaultBody())
	if body != "" {
		paramValues["body"] = body
	}

	contentType := opDefinition.GetDefaultBodyType()
	if contentType != "" {
		paramValues["contentType"] = contentType
	}

	return paramValues
}

func (o ParsedOpenApi) getActionUrlTemplate(op *ops.OperationDefinition) string {
	host := fmt.Sprintf("'%s'", o.requestUrl)
	if o.mask.RequestUrl != "" {
		connTemplate := fmt.Sprintf("connection.%s.%s", o.mask.RequestUrl, consts.RequestUrlKey)
		host = fmt.Sprintf("(exists(%s) ? %s : '%s'),", connTemplate, connTemplate, o.requestUrl)
	}

	path := fmt.Sprintf("'%s'", op.Path)
	for _, param := range op.PathParams {
		old := fmt.Sprintf("%s%s%s", consts.ParamPrefix, param.ParamName, consts.ParamSuffix)
		expr := fmt.Sprintf("' + params.%s + '", param.ParamName)
		path = strings.ReplaceAll(path, old, expr)
	}
	path = strings.TrimSuffix(path, " + ''")

	query := ""
	for index, queryParam := range op.QueryParams {
		if index == 0 {
			query += fmt.Sprintf("'?%s=' + params.%s", queryParam.ParamName, queryParam.ParamName)
		} else {
			query += fmt.Sprintf(" + '&%s=' + params.%s", queryParam.ParamName, queryParam.ParamName)
		}
	}
	if query != "" {
		query = fmt.Sprintf("{{%s}}", query)
	}

	url := fmt.Sprintf("{{%s + %s}}%s", host, path, query)
	return url
}

func getBodyTemplate(action plugin.Action, body *ops.RequestBodyDefinition) string {
	if body == nil {
		return ""
	}
	requestBody := map[string]interface{}{}
	for paramName, _ := range action.Parameters {
		mapKeys := strings.Split(paramName, consts.BodyParamDelimiter)
		buildRequestBody(mapKeys, body.Schema.OApiSchema, paramName, requestBody)
	}
	marshaledBody, err := json.Marshal(requestBody)
	if err != nil {
		return ""
	}
	return string(marshaledBody)
}

func getDisplayName(name string) string {
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, ".", " ")
	name = strings.ReplaceAll(name, "[]", "")
	for i := 1; i < len(name); i++ {
		if unicode.IsLower(rune(name[i-1])) && unicode.IsUpper(rune(name[i])) && (i+1 < len(name) || !unicode.IsUpper(rune(name[i+1]))) {
			name = name[:i] + " " + name[i:]
		}
	}
	upperCaseWords := []string{"url", "id", "ids", "ip", "ssl"}
	words := strings.Split(name, " ")
	for i, word := range words {
		if contains(upperCaseWords, word) {
			words[i] = strings.ToUpper(word)
		}
	}
	name = strings.Join(words, " ")
	name = strings.ReplaceAll(name, "IDS", "IDs")
	return strings.Join(strings.Fields(strings.Title(name)), " ")
}

func contains(list []string, word string) bool {
	for _, item := range list {
		if item == word {
			return true
		}
	}
	return false
}

// Build nested json request body from "." delimited parameters
func buildRequestBody(mapKeys []string, propertySchema *openapi3.Schema, paramName string, requestBody map[string]interface{}) {
	key := mapKeys[0]

	// Keep recursion going until leaf node is found
	if len(mapKeys) == 1 && propertySchema != nil{
		subPropertySchema := ops.GetPropertyByName(key, propertySchema)

		if subPropertySchema != nil {
			requestBody[mapKeys[len(mapKeys)-1]] = castBodyParamType(paramName, subPropertySchema.Type)
		}

	} else {
		if _, ok := requestBody[key]; !ok {
			requestBody[key] = map[string]interface{}{}
		}

		subPropertySchema := ops.GetPropertyByName(key, propertySchema)
		buildRequestBody(mapKeys[1:], subPropertySchema, paramName, requestBody[key].(map[string]interface{}))
	}
}

// Cast proper parameter types when building json request body
func castBodyParamType(paramValue string, paramType string) interface{} {
	switch paramType {
	case consts.TypeString:
		return fmt.Sprintf("{{params.%s}}", paramValue)
	case consts.TypeInteger, consts.TypeNumber:
		return fmt.Sprintf("{{int(params.%s)}}", paramValue)
	case consts.TypeBoolean:
		return fmt.Sprintf("{{bool(params.%s)}}", paramValue)
	case consts.TypeArray:
		return []interface{}{fmt.Sprintf("{{arr(params.%s)}}", paramValue)}
	case consts.TypeObject:
		return fmt.Sprintf("{{obj(params.%s)}}", paramValue)
	default:
		return fmt.Sprintf("{{obj(params.%s)}}", paramValue)
	}
}
