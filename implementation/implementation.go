package implementation

import (
	"errors"
	"fmt"
	"github.com/blinkops/blink-http/plugins"
	"github.com/blinkops/blink-sdk/plugin"
	"github.com/blinkops/blink-sdk/plugin/actions"
	"github.com/blinkops/blink-sdk/plugin/config"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	description2 "github.com/blinkops/blink-sdk/plugin/description"
	log "github.com/sirupsen/logrus"
	"path"
)

type HttpPlugin struct {
	description      plugin.Description
	actions          []plugin.Action
	supportedActions map[string]plugins.ActionHandler
}

func (p *HttpPlugin) Describe() plugin.Description {
	log.Debug("Handling Describe request!")
	return p.description
}

func (p *HttpPlugin) GetActions() []plugin.Action {
	log.Debug("Handling GetActions request!")
	return p.actions
}

func (p *HttpPlugin) ExecuteAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) (*plugin.ExecuteActionResponse, error) {
	log.Debugf("Executing action: %v\n Context: %v", *request, ctx.GetAllContextEntries())

	actionHandler, ok := p.supportedActions[request.Name]
	if !ok {
		return nil, errors.New("action is not supported: " + request.Name)
	}

	resultBytes, err := actionHandler(ctx, request)
	if err != nil {
		if resultBytes == nil {
			log.Error("Failed executing action, err: ", err)
			return nil, err
		}

		return &plugin.ExecuteActionResponse{
			ErrorCode: 1,
			Result:    resultBytes,
		}, nil

	}

	if len(resultBytes) > 0 && resultBytes[len(resultBytes)-1] == '\n' {
		resultBytes = resultBytes[:len(resultBytes)-1]
	}

	return &plugin.ExecuteActionResponse{
		ErrorCode: 0,
		Result:    resultBytes,
	}, nil
}

func (p *HttpPlugin) TestCredentials(connections map[string]*blink_conn.ConnectionInstance) (*plugin.CredentialsValidationResponse, error) {
	for connName, connInstance := range connections {
		integration, ok := plugins.Plugins[connName]
		if !ok {
			return &plugin.CredentialsValidationResponse {
				AreCredentialsValid:   false,
				RawValidationResponse: []byte(fmt.Sprintf("Test connection failed. Connection type %s isn't supported by the http plugin", connName)),
			}, nil
		}

		isValid, response := integration.TestConnection(connInstance)
		return &plugin.CredentialsValidationResponse{
			AreCredentialsValid:   isValid,
			RawValidationResponse: response,
		}, nil

	}
	return &plugin.CredentialsValidationResponse{
		AreCredentialsValid:   false,
		RawValidationResponse: []byte(fmt.Sprintf("Test connection failed. No connection to test was provided")),
	}, nil

}

func NewHTTPPlugin(rootPluginDirectory string) (*HttpPlugin, error) {
	pluginConfig := config.GetConfig()

	description, err := description2.LoadPluginDescriptionFromDisk(path.Join(rootPluginDirectory, pluginConfig.Plugin.PluginDescriptionFilePath))
	if err != nil {
		return nil, err
	}

	loadedConnections, err := blink_conn.LoadConnectionsFromDisk(path.Join(rootPluginDirectory, pluginConfig.Plugin.PluginDescriptionFilePath))
	if err != nil {
		return nil, err
	}

	log.Infof("Loaded %d connections from disk", len(loadedConnections))
	description.Connections = loadedConnections

	actionsFromDisk, err := actions.LoadActionsFromDisk(path.Join(rootPluginDirectory, pluginConfig.Plugin.ActionsFolderPath))
	if err != nil {
		return nil, err
	}

	supportedActions := map[string]plugins.ActionHandler{
		"get":     executeHTTPGetAction,
		"post":    executeHTTPPostAction,
		"put":     executeHTTPPutAction,
		"delete":  executeHTTPDeleteAction,
		"patch":   executeHTTPPatchAction,
		"graphQL": executeGraphQL,
	}

	for _, integration := range plugins.Plugins {
		customPlugin, ok := integration.(plugins.CustomPlugin)
		if !ok {
			continue
		}
		err = addActionsFromPlugin(actionsFromDisk, rootPluginDirectory, customPlugin.GetCustomActionsPath())
		if err != nil {
			return nil, err
		}

		err = addSupportedActions(supportedActions, customPlugin.GetCustomActionHandlers())
		if err != nil {
			return nil, err
		}

	}

	return &HttpPlugin{
		description:      *description,
		actions:          actionsFromDisk,
		supportedActions: supportedActions,
	}, nil
}

func addActionsFromPlugin(currentActions []plugin.Action,rootPluginDirectory string, actionsPath string) error {
	newActionsFromDisk, err := actions.LoadActionsFromDisk(path.Join(rootPluginDirectory, actionsPath))
	if err != nil {
		return err
	}
	currentActions = append(currentActions, newActionsFromDisk...)
	return nil
}

func addSupportedActions(actions map[string]plugins.ActionHandler, newActions map[string]plugins.ActionHandler) error {
	for name, handler := range newActions {
		if _, ok := actions[name]; ok {
			return errors.New(fmt.Sprintf("failed to init plugin, found duplicate action: %s", name))
		}
		actions[name] = handler
	}
	return nil
}
