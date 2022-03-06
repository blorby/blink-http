package connections

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/blinkops/blink-http/consts"
	"github.com/blinkops/blink-http/implementation/requests"
	"github.com/blinkops/blink-http/plugins/types"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"net/http"
	"strings"
	"time"
)

type (
	HeaderValuePrefixes     map[string]string
	HeaderAlias             map[string]string
)

func HandleGenericConnection(connection map[string]string, request *http.Request, prefixes HeaderValuePrefixes, headerAlias HeaderAlias) error {
	headers := make(map[string]string)
	delete(connection, consts.RequestUrlKey)
	for header, headerValue := range connection {
		header = strings.ToUpper(header)
		// if the header is in our alias map replace it with the value in the map
		// TOKEN -> AUTHORIZATION
		if val, ok := headerAlias[header]; ok {
			header = strings.ToUpper(val)
		}

		// we want to help the user by adding prefixes he might have missed
		// for example:   Bearer <TOKEN>
		if val, ok := prefixes[header]; ok {
			if !strings.HasPrefix(headerValue, val) { // check what prefix the user doesn't have
				// add the prefix
				headerValue = val + headerValue
			}
		}

		// If the user supplied BOTH username and password
		// Username:Password pair should be base64 encoded
		// and sent as "Authorization: base64(user:pass)"
		headers[header] = headerValue
		if username, ok := headers[consts.BasicAuthUsername]; ok {
			if password, ok := headers[consts.BasicAuthPassword]; ok {
				header, headerValue = "Authorization", constructBasicAuthHeader(username, password)
				cleanRedundantHeaders(&request.Header)
			}
		}
		request.Header.Set(header, headerValue)
	}

	return nil
}

func constructBasicAuthHeader(username, password string) string {
	data := []byte(fmt.Sprintf("%s:%s", username, password))
	hashed := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("%s%s", consts.BasicAuthPrefix, hashed)
}

func cleanRedundantHeaders(requestHeaders *http.Header) {
	requestHeaders.Del(consts.BasicAuthUsername)
	requestHeaders.Del(consts.BasicAuthPassword)
}

func SendTestConnectionRequest(url string, method string, data []byte, conn *blink_conn.ConnectionInstance, authHandler types.AuthHandler) (*types.Response, error) {
	requestBody := bytes.NewBuffer(data)
	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, err
	}

	if err = authHandler(req, conn.Data); err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: time.Second * time.Duration(consts.DefaultTimeout),
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := requests.ReadBody(res.Body)
	if err != nil {
		return nil, err
	}

	return &types.Response{
		StatusCode: res.StatusCode,
		Body:       body,
	}, nil
}
