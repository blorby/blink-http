package bitbucket

import (
	"encoding/base64"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/plugins/connections"
	"github.com/blinkops/blink-http/plugins/types"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type BitbucketPlugin struct{}

func (p BitbucketPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	// OAuth
	if val, ok := conn["Token"]; ok {
		req.Header.Set("AUTHORIZATION", consts.BearerAuthPrefix+val)
	}
	// Api key
	username, userExists := conn["USERNAME"]
	password, passExists := conn["PASSWORD"]
	if !userExists || !passExists {
		log.Info("failed to set authentication headers")
		return fmt.Errorf("failed to set authentication headers")
	}
	data := []byte(fmt.Sprintf("%s:%s", username, password))
	hashed := base64.StdEncoding.EncodeToString(data)
	req.Header.Set("AUTHORIZATION", consts.BasicAuthPrefix+hashed)
	return nil
}

func (p BitbucketPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	requestUrl, ok := connection.Data[consts.RequestUrlKey]
	if !ok {
		requestUrl = p.GetDefaultRequestUrl()
	}
	res, err := connections.SendTestConnectionRequest(fmt.Sprintf("%s/user", requestUrl), http.MethodGet, nil, connection, p.HandleAuth)
	if err != nil {
		return false, []byte(err.Error())
	}

	if res.StatusCode != http.StatusOK {
		return false, []byte(consts.TestConnectionFailed)
	}

	return true, nil
}

func (p BitbucketPlugin) GetDefaultRequestUrl() string {
	return "https://api.bitbucket.org/2.0"
}

func (p BitbucketPlugin) GetCustomActionHandlers() map[string] types.ActionHandler {
	return map[string]types.ActionHandler{
		"AddUsersToGroup": addUsersToGroup,
	}
}

func (p BitbucketPlugin) GetCustomActionsPath() string {
	return "plugins/bitbucket/actions"
}

func GetNewBitbucketPlugin() BitbucketPlugin {
	return BitbucketPlugin{}
}
