package plugins

import (
	"github.com/blinkops/blink-http/plugins/azure"
	"github.com/blinkops/blink-http/plugins/azure-devops"
	"github.com/blinkops/blink-http/plugins/bitbucket"
	"github.com/blinkops/blink-http/plugins/datadog"
	"github.com/blinkops/blink-http/plugins/elasticsearch"
	"github.com/blinkops/blink-http/plugins/gcp"
	"github.com/blinkops/blink-http/plugins/github"
	"github.com/blinkops/blink-http/plugins/gitlab"
	"github.com/blinkops/blink-http/plugins/grafana"
	"github.com/blinkops/blink-http/plugins/jira"
	"github.com/blinkops/blink-http/plugins/okta"
	"github.com/blinkops/blink-http/plugins/opsgenie"
	"github.com/blinkops/blink-http/plugins/pagerduty"
	"github.com/blinkops/blink-http/plugins/pingdom"
	"github.com/blinkops/blink-http/plugins/prometheus"
	"github.com/blinkops/blink-http/plugins/slack"
	"github.com/blinkops/blink-http/plugins/types"
	virus_total "github.com/blinkops/blink-http/plugins/virus-total"
	"github.com/blinkops/blink-http/plugins/wiz"
)

var Plugins = map[string]types.Plugin{
	"azure":         azure.GetNewAzurePlugin(),
	"azure-devops":  azure_devops.GetNewAzureDevopsPlugin(),
	"bitbucket":     bitbucket.GetNewBitbucketPlugin(),
	"datadog":       datadog.GetNewDatadogPlugin(),
	"elasticsearch": elasticsearch.GetNewElasticSearchPlugin(),
	"gcp":           gcp.GetNewGcpPlugin(),
	"github":        github.GetNewGithubPlugin(),
	"gitlab":        gitlab.GetNewGitlabPlugin(),
	"grafana":       grafana.GetNewGrafanaPlugin(),
	"jira":          jira.GetNewJiraPlugin(),
	"okta":          okta.GetNewOktaPlugin(),
	"opsgenie":      opsgenie.GetNewOpsgeniePlugin(),
	"pagerduty":     pagerduty.GetNewPagerdutyPlugin(),
	"pingdom":       pingdom.GetNewPingdomPlugin(),
	"prometheus":    prometheus.GetNewPrometheusPlugin(),
	"slack":         slack.GetNewSlackPlugin(),
	"virus-total":   virus_total.GetNewVirusTotalPlugin(),
	"wiz":           wiz.GetNewWizPlugin(),
}
