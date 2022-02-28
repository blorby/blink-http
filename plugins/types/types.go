package types

import (
	"github.com/blinkops/blink-sdk/plugin"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

type ActionHandler func(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin Plugin) ([]byte, error)
type AuthHandler func(req *http.Request, conn map[string]string) error

type Plugin interface {
	TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte)
	HandleAuth(req *http.Request, conn map[string]string) error
}

type CustomPlugin interface {
	Plugin
	GetCustomActionHandlers() map[string]ActionHandler
	GetCustomActionsPath() string
}

type PluginWithValidation interface {
	Plugin
	ValidateResponse(statusCode int, body []byte) ([]byte, error)
}
