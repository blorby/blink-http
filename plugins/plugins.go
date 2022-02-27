package plugins

import (
	"github.com/blinkops/blink-http/plugins/types"
	"github.com/blinkops/blink-http/plugins/wiz"
)

var Plugins = map[string]types.Plugin{
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
	"wiz":           wiz.GetNewWizPlugin(),
}
