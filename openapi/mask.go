package openapi

import (
	"github.com/blinkops/blink-sdk/plugin/connections"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"strings"
)

const (
	FormatDelimiter = "_"
)

var FormatPrefixes = [...]string{"date"}

type (
	Mask struct {
		RequestUrl               string                            `yaml:"request_url_conn_type"` // this is used to generate a template for the request url
		Actions                  map[string]*MaskedAction          `yaml:"actions,omitempty"`
		Name                     string                            `yaml:"name"`
		Connections              map[string]connections.Connection `yaml:"connection_types"`
		IconUri                  string                            `yaml:"icon_uri"`
		IsConnectionOptional     bool                              `yaml:"is_connection_optional"`
		CollectionId             string
		ActionConnectionTypes    []string
		ReverseActionAliasMap    map[string]string
		ReverseParameterAliasMap map[string]map[string]string
	}
	MaskedAction struct {
		Alias       string                            `yaml:"alias,omitempty"`
		DisplayName string                            `yaml:"display_name"`
		Description string                            `yaml:"description,omitempty"`
		Parameters  map[string]*MaskedActionParameter `yaml:"parameters,omitempty"`
	}
	MaskedActionParameter struct {
		Alias       string `yaml:"alias,omitempty"`
		Required    bool   `yaml:"required,omitempty"`
		Type        string `yaml:"type,omitempty"` // password/date - 2017-07-21/date_time - 2017-07-21T17:32:28Z/date_epoch - 1631399887
		Index       int64  `yaml:"index,omitempty"`
		IsMulti     bool   `yaml:"is_multi,omitempty"` // is this a multi-select field
		Default     string `yaml:"default,omitempty"`  // override parameter default value
		Description string `yaml:"description,omitempty"`
	}
)

// ParseMask receives a mask file, parses it and returns a new mask object.
func ParseMask(maskFile string) (mask Mask, err error) {
	if maskFile == "" {
		return
	}

	var rawMaskData []byte
	rawMaskData, err = ioutil.ReadFile(maskFile)
	if err != nil {
		return
	}

	rawMaskData = []byte(strings.ReplaceAll(string(rawMaskData), "\\", ""))
	if err = yaml.Unmarshal(rawMaskData, &mask); err != nil {
		return
	}

	mask.buildActionAliasMap()
	mask.buildParamAliasMap()
	return
}

// GetAction receives an action's name and returns
func (m *Mask) GetAction(actionName string) *MaskedAction {
	originalActionName := m.ReplaceActionAlias(actionName)

	if action, ok := m.Actions[originalActionName]; ok {
		return action
	}

	return nil
}

func (m *Mask) GetParameter(actionName string, paramName string) *MaskedActionParameter {
	originalActionName := m.ReplaceActionAlias(actionName)
	originalParamName := m.replaceActionParameterAlias(actionName, paramName)

	if action, ok := m.Actions[originalActionName]; ok {
		if param, ok := action.Parameters[originalParamName]; ok {
			return param
		}

		return nil
	}

	return nil
}

func (m *Mask) ReplaceActionAlias(actionName string) string {
	if originalName, ok := m.ReverseActionAliasMap[actionName]; ok {
		return originalName
	}
	return actionName
}

func (m *Mask) ReplaceActionParametersAliases(originalActionName string, rawParameters map[string]string) map[string]string {
	requestParameters := map[string]string{}

	for paramName, paramValue := range rawParameters {
		originalName := m.replaceActionParameterAlias(originalActionName, paramName)
		requestParameters[originalName] = paramValue
	}

	return requestParameters
}

func (m *Mask) buildActionAliasMap() {
	reverseActionAliasMap := map[string]string{}

	for originalName, actionData := range m.Actions {
		if actionData.Alias != "" {
			if _, ok := reverseActionAliasMap[actionData.Alias]; ok {
				// error
				log.Fatalln("Alias " + actionData.Alias + " exist multiple times.")
			}
			reverseActionAliasMap[actionData.Alias] = originalName
		}
	}

	m.ReverseActionAliasMap = reverseActionAliasMap
}

func (m *Mask) buildParamAliasMap() {
	reverseParameterAliasMap := map[string]map[string]string{}
	for actionName, actionData := range m.Actions {
		reverseParameterAliasMap[actionName] = map[string]string{}

		for originalName, parameterData := range actionData.Parameters {
			if parameterData.Alias != "" {
				reverseParameterAliasMap[actionName][parameterData.Alias] = originalName
			}
		}
	}
	m.ReverseParameterAliasMap = reverseParameterAliasMap
}

func (m *Mask) replaceActionParameterAlias(actionName string, paramName string) string {
	if actionParams, ok := m.ReverseParameterAliasMap[actionName]; ok {
		if originalName, ok := actionParams[paramName]; ok {
			return originalName
		}
	}
	return paramName
}
