package openapi

import (
	"fmt"
	"github.com/blinkops/blink-http/consts"
	ops "github.com/blinkops/blink-http/openapi/operations"
	"github.com/blinkops/blink-sdk/plugin"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v2"
	"net/url"
	"sort"
	"strings"
)

type ParsedOpenApi struct {
	requestUrl string
	description string
	actions []plugin.Action
	operations map[string]*ops.OperationDefinition
	mask Mask
}

func GetNamedActionsFromOpenapi(maskFile string, openApiFile string) ([]byte, error) {
	maskData, err := ParseMask(maskFile)
	if err != nil {
		return nil, err
	}

	openApi, err := parseOpenApiFile(openApiFile)
	openApi.mask = maskData
	if err != nil {
		return nil, err
	}
	namedActions := openApi.getNamedActionsGroup()
	namedActionsFile, err := yaml.Marshal(namedActions)
	if err != nil {
		return nil, err
	}
	return namedActionsFile, nil
}

func parseOpenApiFile(OpenApiFile string) (ParsedOpenApi, error) {
	var actions []plugin.Action

	openApi, err := loadOpenApi(OpenApiFile)
	if err != nil || len(openApi.Servers) == 0{
		return ParsedOpenApi{}, err
	}

	// Set default openApi server
	openApiServer := openApi.Servers[0]
	requestUrl := openApiServer.URL

	for urlVariableName, urlVariable := range openApiServer.Variables {
		requestUrl = strings.ReplaceAll(requestUrl, consts.ParamPrefix+urlVariableName+consts.ParamSuffix, urlVariable.Default)
	}

	operationDefinitions , err := ops.DefineOperations(openApi)
	if err != nil {
		return ParsedOpenApi{}, err
	}

	for _, operation := range operationDefinitions {
		actionName := operation.OperationId

		action := plugin.Action{
			Name:        actionName,
			Description: operation.Summary,
			Enabled:     true,
			EntryPoint:  operation.Path,
			Parameters:  map[string]plugin.ActionParameter{},
		}

		for _, pathParam := range operation.AllParams() {
			paramName := pathParam.ParamName
			paramDescription := pathParam.Schema.Description

			if paramDescription == "" {
				paramDescription = pathParam.Spec.Description
			}

			if actionParam := parseActionParam(pathParam.Spec.Schema, pathParam.Required, paramDescription); actionParam != nil {
				action.Parameters[paramName] = *actionParam
			}
		}

		for _, paramBody := range operation.Bodies {
			if paramBody.DefaultBody {

				handleBodyParams(&action, paramBody.Schema.OApiSchema, "", "", paramBody.Required)
				break
			}
		}

		actions = append(actions, action)
	}

	// sort the actions
	// each time we parse the ParsedOpenApi the actions are added to the map in different order.
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Name < actions[j].Name
	})
	return ParsedOpenApi{
		description: openApi.Info.Description,
		requestUrl:  requestUrl,
		actions:     actions,
		operations: operationDefinitions,
	}, nil
}

func loadOpenApi(filePath string) (openApi *openapi3.T, err error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	u, err := url.Parse(filePath)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return loader.LoadFromURI(u)
	}

	return loader.LoadFromFile(filePath)
}

func handleBodyParams(action *plugin.Action, paramSchema *openapi3.Schema, parentPath string, schemaPath string, parentsRequired bool) {
	handleBodyParamOfType(action, paramSchema, parentPath, schemaPath, parentsRequired)

	for propertyName, bodyProperty := range paramSchema.Properties {
		fullParamPath, fullSchemaPath := propertyName, schemaPath

		if bodyProperty.Ref != "" {
			index := strings.LastIndex(bodyProperty.Ref, "/") + 1
			fullSchemaPath += bodyProperty.Ref[index:] + consts.BodyParamDelimiter
			if hasDuplicateSchemas(fullSchemaPath) {
				continue
			}
		}

		// Json params are represented as dot delimited params to allow proper parsing in UI later on
		if parentPath != "" {
			fullParamPath = parentPath + consts.BodyParamDelimiter + fullParamPath
		}

		// Keep recursion until leaf node is found
		if bodyProperty.Value.Properties != nil {
			handleBodyParams(action, bodyProperty.Value, fullParamPath, fullSchemaPath, areParentsRequired(parentsRequired, propertyName, paramSchema))
		} else {
			handleBodyParamOfType(action, bodyProperty.Value, fullParamPath, fullSchemaPath, parentsRequired)
			isParamRequired := false

			for _, requiredParam := range paramSchema.Required {
				if propertyName == requiredParam {
					isParamRequired = parentsRequired
					break
				}
			}

			if actionParam := parseActionParam(bodyProperty, isParamRequired, bodyProperty.Value.Description); actionParam != nil {
				action.Parameters[fullParamPath] = *actionParam
			}
		}
	}
}

func handleBodyParamOfType(action *plugin.Action, paramSchema *openapi3.Schema, parentPath string, schemaPath string, parentsRequired bool) {
	if paramSchema.AllOf != nil || paramSchema.AnyOf != nil || paramSchema.OneOf != nil {

		allSchemas := []openapi3.SchemaRefs{paramSchema.AllOf, paramSchema.AnyOf, paramSchema.OneOf}

		// find properties nested in Allof, Anyof, Oneof
		for _, schemaType := range allSchemas {
			for _, schemaParams := range schemaType {
				handleBodyParams(action, schemaParams.Value, parentPath, schemaPath, parentsRequired)
			}
		}
	}
}

func parseActionParam(paramSchema *openapi3.SchemaRef, isParamRequired bool, paramDescription string) *plugin.ActionParameter {
	var isMulti    bool
	var	paramIndex int64

	paramType := paramSchema.Value.Type
	paramFormat := paramSchema.Value.Format
	paramOptions := getParamOptions(paramSchema.Value.Enum, &paramType)
	paramPlaceholder := getParamPlaceholder(paramSchema.Value.Example, paramType)
	paramDefault := getParamDefault(paramSchema.Value.Default, paramType)
	paramIndex = 999 // parameters will be ordered from lowest to highest in UI. This is the default, meaning - the end of the list.

	// Convert parameters of type object to code:json and parameters of type boolean to bool
	paramType = convertParamType(paramType)

	return &plugin.ActionParameter{
		Type:        paramType,
		Description: paramDescription,
		Placeholder: paramPlaceholder,
		Required:    isParamRequired,
		Default:     paramDefault,
		Options:     paramOptions,
		Index:       paramIndex,
		Format:      paramFormat,
		IsMulti:     isMulti,
	}
}

func areParentsRequired(parentsRequired bool, propertyName string, schema *openapi3.Schema) bool {
	if !parentsRequired {
		return false
	}

	for _, requiredParam := range schema.Required {
		if propertyName == requiredParam {
			return true
		}
	}

	return false
}

func getParamOptions(parsedOptions []interface{}, paramType *string) []string {
	paramOptions := []string{}

	if parsedOptions == nil {
		return nil
	}

	for _, option := range parsedOptions {
		if optionString, ok := option.(string); ok {
			paramOptions = append(paramOptions, optionString)
		}
	}

	if len(paramOptions) > 0 {
		*paramType = consts.TypeDropdown
	}

	return paramOptions
}

func getParamPlaceholder(paramExample interface{}, paramType string) string {
	paramPlaceholder, _ := paramExample.(string)

	if paramType != consts.TypeObject {
		if paramPlaceholder != "" {
			return consts.ParamPlaceholderPrefix + paramPlaceholder
		}
	}

	return paramPlaceholder
}

func getParamDefault(defaultValue interface{}, paramType string) string {
	var paramDefault string

	if paramType != consts.TypeArray {
		if defaultValue == nil {
			paramDefault = ""
		} else {
			paramDefault = fmt.Sprintf("%v", defaultValue)
		}

		return paramDefault
	}

	if defaultList, ok := defaultValue.([]interface{}); ok {
		var defaultStrings []string

		for _, value := range defaultList {
			valueString := fmt.Sprintf("%v", value)
			defaultStrings = append(defaultStrings, valueString)
		}

		paramDefault = strings.Join(defaultStrings, consts.ArrayDelimiter)
	}

	return paramDefault
}

func hasDuplicateSchemas(path string) bool {
	paramsArray := strings.Split(path, consts.BodyParamDelimiter)
	exists := make(map[string]bool)
	for _, param := range paramsArray {
		if exists[param] {
			return true
		} else {
			exists[param] = true
		}
	}
	return false
}

func convertParamType(paramType string) string{
	switch paramType {
	case consts.TypeObject:
		return consts.TypeJson
	case consts.TypeBoolean:
		return consts.TypeBool
	}
	return paramType
}
