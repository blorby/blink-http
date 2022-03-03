package requests

import (
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/plugins/datadog"
	"github.com/blinkops/blink-http/plugins/github"
	"github.com/blinkops/blink-http/plugins/jira"
	"github.com/blinkops/blink-http/plugins/types"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
)

type HttpTestSuite struct {
	suite.Suite
}

func TestHttpTestSuite(t *testing.T) {
	suite.Run(t, new(HttpTestSuite))
}

func (suite *HttpTestSuite) TestValidateURL() {
	type testCase struct {
		connection   map[string]string
		requestedURL string
		plugin       types.Plugin
	}
	prefixes := []string{"", "www.", "https://", "https://www."}
	host := "host.com"
	// check all possible prefix combinations work
	for _, prefix := range prefixes {
		for _, prefix2 := range prefixes {
			connection := map[string]string{consts.RequestUrlKey: prefix + host}
			requestedUrl := prefix2 + host
			u, err := url.Parse(requestedUrl)
			suite.Nil(err)
			err = validateURL(connection, u, nil)
			suite.Nil(err)
		}
	}
	for _, goodScenario := range []testCase{
		{
			connection:   map[string]string{},
			requestedURL: "https://host.com",
		},
		{
			connection:   map[string]string{consts.ApiAddressKey: ""},
			requestedURL: "https://host.com",
		},
		{
			connection:   map[string]string{"bad key": "www.host.com"},
			requestedURL: "host.com",
		},
		{
			connection:   map[string]string{consts.ApiAddressKey: "https://www.host.com"},
			requestedURL: "subdomain.host.com",
		},
		{
			connection: map[string]string{consts.RequestUrlKey: ""},
			requestedURL: "mydomain.atlassian.net",
			plugin: jira.GetNewJiraPlugin(),
		},
		{
			connection: map[string]string{consts.RequestUrlKey: "www.mydomain.atlassian.net"},
			requestedURL: "mydomain.atlassian.net/path",
			plugin: jira.GetNewJiraPlugin(),
		},
		{
			connection: map[string]string{consts.RequestUrlKey: "mydomain.github.com"},
			requestedURL: "www.mydomain.github.com/path/path2",
			plugin: github.GetNewGithubPlugin(),
		},
		{
			connection: map[string]string{},
			requestedURL: "https://www.api.github.com/path",
			plugin: github.GetNewGithubPlugin(),
		},
		{
			connection: map[string]string{consts.RequestUrlKey: "https://api.datadoghq.com"},
			requestedURL: "api.datadoghq.com",
			plugin: datadog.GetNewDatadogPlugin(),
		},
	} {
		u, err := url.Parse(goodScenario.requestedURL)
		suite.Nil(err)
		err = validateURL(goodScenario.connection, u, goodScenario.plugin)
		suite.Nil(err)
	}
	for _, badScenario := range []testCase{
		{
			connection:   map[string]string{consts.ApiAddressKey: "www.host.com"},
			requestedURL: "bad-address.com",
		},
		{
			connection:   map[string]string{consts.ApiAddressKey: "www.host.com/good-path"},
			requestedURL: "host.com/bad-path",
		},
		{
			connection:   map[string]string{consts.ApiAddressKey: "www.host.com"},
			requestedURL: "fhost.com",
		},
		{
			connection: map[string]string{consts.RequestUrlKey: "www.host.com"},
			requestedURL: "mydomain.atlassian.net",
			plugin: jira.GetNewJiraPlugin(),
		},
		{
			connection: map[string]string{},
			requestedURL: "https://www.host.com",
			plugin: github.GetNewGithubPlugin(),
		},
		{
			connection: map[string]string{consts.RequestUrlKey: "https://api.datadoghq.com"},
			requestedURL: "datadog.com",
			plugin: datadog.GetNewDatadogPlugin(),
		},
	} {
		u, err := url.Parse(badScenario.requestedURL)
		suite.Nil(err)
		err = validateURL(badScenario.connection, u, badScenario.plugin)
		suite.NotNil(err)
	}
}
