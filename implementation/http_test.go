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
	prefixes := []string{"", "www.", "https://", "https://www."}
	host := "host.com"
	// check all possible prefix combinations work
	for _, prefix := range prefixes {
		for _, prefix2 := range prefixes {
			connection := map[string]interface{}{apiAddressKey: prefix + host}
			requestedUrl := prefix2 + host
			u, err := url.Parse(requestedUrl)
			suite.Nil(err)
			err = validateURL(connection, u)
			suite.Nil(err)
		}
	}
	for _, goodScenario := range []testCase{
		{
			connection:   map[string]interface{}{},
			requestedURL: "https://host.com",
		},
		{
			connection:   map[string]interface{}{apiAddressKey: ""},
			requestedURL: "https://host.com",
		},
		{
			connection:   map[string]interface{}{"bad key": "www.host.com"},
			requestedURL: "host.com",
		},
	} {
		u, err := url.Parse(goodScenario.requestedURL)
		suite.Nil(err)
		err = validateURL(goodScenario.connection, u)
		suite.Nil(err)
	}
	for _, badScenario := range []testCase{
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
