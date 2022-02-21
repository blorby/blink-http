package plugins

import (
	"github.com/blinkops/blink-sdk/plugin"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
)

var Plugins = map[string]Plugin{
	"azure":         getNewAzurePlugin(),
	"azure-devops":  getNewAzureDevopsPlugin(),
	"bitbucket":     getNewBitbucketPlugin(),
	"datadog":       getNewDatadogPlugin(),
	"elasticsearch": getNewElasticsearchPlugin(),
	"gcp":           getNewGcpPlugin(),
	"github":        getNewGithubPlugin(),
	"gitlab":        getNewGitlabPlugin(),
	"grafana":       getNewGrafanaPlugin(),
	"jira":          getNewJiraPlugin(),
	"okta":          getNewOktaPlugin(),
	"opsgenie":      getNewOpsgeniePlugin(),
	"pagerduty":     getNewPagerdutyPlugin(),
	"pingdom":       getNewPingdomPlugin(),
	"prometheus":    getNewPrometheusPlugin(),
	"slack":         getNewSlackPlugin(),
	"virus-total":   getNewVirusTotalPlugin(),
	"wiz":           getNewWizPlugin(),
}

type ActionHandler func(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error)
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
