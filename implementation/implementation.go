package implementation

import (
	"errors"
	"path"

	"github.com/blinkops/blink-sdk/plugin"
	"github.com/blinkops/blink-sdk/plugin/actions"
	"github.com/blinkops/blink-sdk/plugin/config"
	"github.com/blinkops/blink-sdk/plugin/connections"
	description2 "github.com/blinkops/blink-sdk/plugin/description"
	log "github.com/sirupsen/logrus"
)

type ActionHandler func(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error)

type HttpPlugin struct {
	description      plugin.Description
	actions          []plugin.Action
	supportedActions map[string]ActionHandler
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

func (p *HttpPlugin) TestCredentials(_ map[string]*connections.ConnectionInstance) (*plugin.CredentialsValidationResponse, error) {
	return &plugin.CredentialsValidationResponse{
		AreCredentialsValid:   true,
		RawValidationResponse: []byte("credentials validation is not supported on this plugin :("),
	}, nil
}

func NewHTTPPlugin(rootPluginDirectory string, supportedActions map[string]ActionHandler) (*HttpPlugin, error) {
	pluginConfig := config.GetConfig()

	description, err := description2.LoadPluginDescriptionFromDisk(path.Join(rootPluginDirectory, pluginConfig.Plugin.PluginDescriptionFilePath))
	if err != nil {
		return nil, err
	}

	loadedConnections, err := connections.LoadConnectionsFromDisk(path.Join(rootPluginDirectory, pluginConfig.Plugin.PluginDescriptionFilePath))
	if err != nil {
		return nil, err
	}

	log.Infof("Loaded %d connections from disk", len(loadedConnections))
	description.Connections = loadedConnections

	actionsFromDisk, err := actions.LoadActionsFromDisk(path.Join(rootPluginDirectory, pluginConfig.Plugin.ActionsFolderPath))
	if err != nil {
		return nil, err
	}


	return &HttpPlugin{
		description:      *description,
		actions:          actionsFromDisk,
		supportedActions: supportedActions,
	}, nil
}
