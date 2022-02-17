package plugins

import (
	"encoding/json"
	"github.com/blinkops/blink-http/consts"
	blink_conn "github.com/blinkops/blink-sdk/plugin/connections"
	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type GcpPlugin struct{}

func (p GcpPlugin) HandleAuth(req *http.Request, conn map[string]string) error {
	if token, ok := conn["Token"]; ok {
		req.Header.Set("AUTHORIZATION", consts.BearerAuthPrefix+token)
		return nil
	}

	return handleServiceAccountAuth(conn, req)
}

func (p GcpPlugin) TestConnection(connection *blink_conn.ConnectionInstance) (bool, []byte) {
	return false, []byte("Test connection failed, Gcp is not yet supported by the http plugin")
}

func getNewGcpPlugin() GcpPlugin {
	return GcpPlugin{}
}

func handleServiceAccountAuth(connection map[string]string, request *http.Request) error {
	credentialsString, ok := connection["credentials"]
	if !ok {
		return errors.New("credentials are invalid - could not convert to string")
	}

	signedToken, err := generateJWTToken(credentialsString)
	if err != nil {
		return err
	}

	req, err := constructReq(signedToken)
	if err != nil {
		return err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	defer func() { _ = res.Body.Close() }()
	if err != nil {
		return errors.Errorf("could not execute the http request: %v", err)
	}

	accessToken, err := extractAccessToken(res)
	if err != nil {
		return err
	}
	request.Header.Set("AUTHORIZATION", consts.BearerAuthPrefix+accessToken)

	return nil
}

func generateJWTToken(credentials string) (string, error) {
	var credentialsJson struct {
		PrivateKey  string `json:"private_key"`
		ClientEmail string `json:"client_email"`
	}

	err := json.Unmarshal([]byte(credentials), &credentialsJson)
	if err != nil {
		return "", errors.Errorf("the provided credentials are invalid: missing private key or client information: %v", err)
	}

	mySigningKey := []byte(credentialsJson.PrivateKey)
	rsaSigningKey, err := jwt.ParseRSAPrivateKeyFromPEM(mySigningKey)
	if err != nil {
		return "", errors.Errorf("the provided credentials are invalid: private key is malformed or missing: %v", err)
	}

	type GCPClaims struct {
		Scope string `json:"scope"`
		jwt.StandardClaims
	}
	claims := GCPClaims{
		"https://www.googleapis.com/auth/cloud-platform",
		jwt.StandardClaims{
			Audience:  "https://oauth2.googleapis.com/token",
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Unix() + 3600,
			Issuer:    credentialsJson.ClientEmail,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(rsaSigningKey)
	if err != nil {
		return "", errors.Errorf("the provided credentials are invalid: %v", err)
	}

	return signedToken, nil
}

func constructReq(signedToken string) (*http.Request, error) {
	gcpAuthUrl := "https://oauth2.googleapis.com/token"
	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	data.Set("assertion", signedToken)
	req, err := http.NewRequest("POST", gcpAuthUrl, strings.NewReader(data.Encode())) // URL-encoded payload
	if err != nil {
		return &http.Request{}, errors.Errorf("could not form an http request using the provided credentials: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	return req, nil
}

func extractAccessToken(res *http.Response) (string, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", errors.Errorf("an error occurred reading the response body: %v", err)
	}

	var responseBody struct {
		AccessToken string `json:"access_token"`
	}
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return "", errors.Errorf("could not get the access token from the response: %v", err)
	}
	return responseBody.AccessToken, nil
}