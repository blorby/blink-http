package implementation

import (
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
		connection   map[string]interface{}
		requestedURL string
	}
	for _, goodScenario := range []testCase{
		{
			connection:   map[string]interface{}{apiAddressKey: "www.host.com"},
			requestedURL: "host.com",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "www.host.com"},
			requestedURL: "www.host.com",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "www.host.com/path"},
			requestedURL: "https://host.com/path/path2",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "www.host.com"},
			requestedURL: "https://www.host.com",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "host.com"},
			requestedURL: "host.com",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "host.com/path"},
			requestedURL: "www.host.com/path",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "host.com"},
			requestedURL: "https://host.com",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "host.com"},
			requestedURL: "https://www.host.com",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "https://www.host.com/path"},
			requestedURL: "host.com/path/path2",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "https://www.host.com"},
			requestedURL: "www.host.com",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "https://www.host.com"},
			requestedURL: "https://host.com",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "https://www.host.com"},
			requestedURL: "https://www.host.com",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "https://host.com"},
			requestedURL: "host.com",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "https://host.com"},
			requestedURL: "www.host.com",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "https://host.com"},
			requestedURL: "https://host.com",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "https://host.com"},
			requestedURL: "https://www.host.com",
		},
	} {
		u, err := url.Parse(goodScenario.requestedURL)
		suite.Nil(err)
		err = validateURL(goodScenario.connection, u)
		suite.Nil(err)
	}
	for _, badScenario := range []testCase{
		{
			connection:   map[string]interface{}{"bad key": "www.host.com"},
			requestedURL: "host.com",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "www.host.com"},
			requestedURL: "bad-address.com",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "www.host.com/good-path"},
			requestedURL: "host.com/bad-path",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "https://www.host.com"},
			requestedURL: "subdomain.host.com",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: "www.host.com"},
			requestedURL: "fhost.com",
		},
	} {
		u, err := url.Parse(badScenario.requestedURL)
		suite.Nil(err)
		err = validateURL(badScenario.connection, u)
		suite.NotNil(err)
	}
}
