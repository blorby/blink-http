package implementation

import (
	"github.com/blinkops/blink-http/consts"
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
	}
	prefixes := []string{"", "www.", "https://", "https://www."}
	host := "host.com"
	// check all possible prefix combinations work
	for _, prefix := range prefixes {
		for _, prefix2 := range prefixes {
			connection := map[string]string{consts.ApiAddressKey: prefix + host}
			requestedUrl := prefix2 + host
			u, err := url.Parse(requestedUrl)
			suite.Nil(err)
			err = validateURL(connection, u)
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
	} {
		u, err := url.Parse(goodScenario.requestedURL)
		suite.Nil(err)
		err = validateURL(goodScenario.connection, u)
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
			connection:   map[string]string{consts.ApiAddressKey: "https://www.host.com"},
			requestedURL: "subdomain.host.com",
		},
		{
			connection:   map[string]string{consts.ApiAddressKey: "www.host.com"},
			requestedURL: "fhost.com",
		},
	} {
		u, err := url.Parse(badScenario.requestedURL)
		suite.Nil(err)
		err = validateURL(badScenario.connection, u)
		suite.NotNil(err)
	}
}
