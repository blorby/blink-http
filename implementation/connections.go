package implementation

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/blinkops/blink-openapi-sdk/consts"
	"github.com/blinkops/blink-sdk/plugin"
	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type (
	genericConnectionOpts struct {
		headerValuePrefixes
		headerAlias
	}
	headerValuePrefixes     map[string]string
	headerAlias             map[string]string
	uniqueConnectionHandler func(connection map[string]interface{}, request *http.Request) error
)

var uniqueConnectionHandlersMap = map[string]uniqueConnectionHandler{
	"gcp":          handleGCPConnection,
	"azure":        handleAzureConnection,
	"azure-devops": handleAzureDevopsConnection,
	"bitbucket":    handleBitbucketConnection,
	"pagerduty":    handlePagerdutyConnection,
	basicAuthKey:   handleBasicAuth,
	bearerAuthKey:  handleBearerToken,
	apiTokenKey:    handleApiKeyAuth,
}

var genericConnectionsOptsMap = map[string]genericConnectionOpts{
	"github": {
		headerValuePrefixes{"AUTHORIZATION": consts.BearerAuth},
		headerAlias{"TOKEN": "AUTHORIZATION"},
	},
	"gitlab": {
		headerValuePrefixes{"AUTHORIZATION": consts.BearerAuth},
		headerAlias{"TOKEN": "AUTHORIZATION"},
	},
	"grafana": {
		headerValuePrefixes{"AUTHORIZATION": consts.BearerAuth},
		headerAlias{"API_KEY": "AUTHORIZATION"},
	},
	"jira": {
		headerValuePrefixes{"AUTHORIZATION": consts.BasicAuth},
		headerAlias{"API TOKEN": "PASSWORD", "USER EMAIL": "USERNAME"},
	},
	"opsgenie": {
		headerValuePrefixes{"AUTHORIZATION": "GenieKey "},
		headerAlias{"TOKEN": "AUTHORIZATION"},
	},
	"pingdom": {
		headerValuePrefixes{"AUTHORIZATION": consts.BearerAuth},
		headerAlias{"TOKEN": "AUTHORIZATION"},
	},
	"slack": {
		headerValuePrefixes{"AUTHORIZATION": consts.BearerAuth},
		headerAlias{"TOKEN": "AUTHORIZATION"},
	},
	"virus-total": {
		nil,
		headerAlias{"API KEY": "x-apikey"},
	},
	"elasticsearch": {
		headerValuePrefixes{"AUTHORIZATION": "ApiKey "},
		headerAlias{"API_KEY": "AUTHORIZATION"},
	},
	"prometheus": {},
	"datadog":    {},
	"okta":       {},
}

func handleAuthentication(ctx *plugin.ActionContext, req *http.Request) error {
	for connName, connInstance := range ctx.GetAllConnections() {
		secret, err := connInstance.ResolveCredentials()
		if err != nil {
			return errors.Errorf("failed to resolve credentials for %s: %v", connName, err)
		}
		if handler, ok := uniqueConnectionHandlersMap[connName]; ok {
			err = handler(secret, req)
		} else if connectionOpts, ok := genericConnectionsOptsMap[connName]; ok {
			err = handleGenericConnection(secret, req, connectionOpts.headerValuePrefixes, connectionOpts.headerAlias)
		}
		if err != nil {
			return errors.Errorf("failed to set auth headers for connection %s: %v", connName, err)
		}
	}
	return nil
}

func handleBasicAuth(connection map[string]interface{}, req *http.Request) error {
	if err := validateURL(connection, req.URL); err != nil {
		return err
	}
	uname, ok := connection[usernameKey].(string)
	if !ok {
		return errors.New("basic-auth connection does not contain a username attribute or the username is not a string")
	}
	password, ok := connection[passwordKey].(string)
	if !ok {
		return errors.New("basic-auth connection does not contain a password attribute or the password is not a string")
	}
	req.Header.Add("Authorization", "Basic "+basicAuth(uname, password))
	return nil
}

func handleBearerToken(connection map[string]interface{}, req *http.Request) error {
	if err := validateURL(connection, req.URL); err != nil {
		return err
	}
	token, ok := connection[tokenKey].(string)
	if !ok {
		return errors.New("bearer-token connection does not contain a token attribute or the token is not a string")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	return nil
}

func handleApiKeyAuth(connection map[string]interface{}, req *http.Request) error {
	if err := validateURL(connection, req.URL); err != nil {
		return err
	}
	for i := 1; i <= 3; i++ {
		h, ok1 := connection["Header "+fmt.Sprint(i)].(string)
		v, ok2 := connection["Value "+fmt.Sprint(i)].(string)
		if ok1 && ok2 {
			req.Header.Add(h, v)
		}
	}
	return nil
}

func validateURL(connection map[string]interface{}, requestedURL *url.URL) error {
	apiAddressInter, ok := connection[apiAddressKey]
	// if there's no api address defined, any url will be allowed
	if !ok || apiAddressInter == nil {
		return nil
	}
	apiAddressString, ok := apiAddressInter.(string)
	if !ok {
		return errors.New("the api address is not convertible to a string")
	}
	apiAddressString = strings.Replace(apiAddressString, "www.", "", 1)
	apiAddress, err := url.Parse(apiAddressString)
	if err != nil {
		return err
	}
	requestedURL, err = url.Parse(strings.Replace(requestedURL.String(), "www.", "", 1))
	if err != nil {
		return err
	}
	if !strings.HasPrefix(requestedURL.Host+requestedURL.Path, apiAddress.Host+apiAddress.Path) {
		return errors.New("the requested url's host/path does not match the host/path defined in the connection. this is not allowed in order to prevent sending credentials to unwanted hosts/paths. the allowed host/path is " + apiAddressString)
	}
	return nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func handleBitbucketConnection(connection map[string]interface{}, request *http.Request) error {
	// OAuth
	if val, ok := connection["Token"]; ok {
		if valString, ok := val.(string); ok {
			request.Header.Set("AUTHORIZATION", consts.BearerAuth+valString)
			return nil
		}
	}
	// Api key
	username, userExists := connection["USERNAME"]
	password, passExists := connection["PASSWORD"]
	if !userExists || !passExists {
		log.Info("failed to set authentication headers")
		return fmt.Errorf("failed to set authentication headers")
	}
	if userString, ok := username.(string); ok {
		if passString, ok := password.(string); ok {
			data := []byte(fmt.Sprintf("%s:%s", userString, passString))
			hashed := base64.StdEncoding.EncodeToString(data)
			request.Header.Set("AUTHORIZATION", consts.BasicAuth+hashed)
			return nil
		}
	}
	log.Info("failed to set authentication headers")
	return fmt.Errorf("failed to set authentication headers")
}

func handlePagerdutyConnection(connection map[string]interface{}, request *http.Request) error {
	// Api key
	if val, ok := connection["api_key"]; ok {
		if valString, ok := val.(string); ok {
			request.Header.Set("AUTHORIZATION", "Token token="+valString)
			return nil
		}
	}
	// OAuth
	if val, ok := connection["Token"]; ok {
		if valString, ok := val.(string); ok {
			request.Header.Set("AUTHORIZATION", "Bearer "+valString)
			request.Header.Set("Accept", "application/vnd.pagerduty+json;version=2")
			return nil
		}
	}
	log.Info("failed to set authentication headers")
	return fmt.Errorf("failed to set authentication headers")
}

func handleAzureDevopsConnection(connection map[string]interface{}, request *http.Request) error {
	// OAuth
	if val, ok := connection["Token"]; ok {
		if valString, ok := val.(string); ok {
			request.Header.Set("AUTHORIZATION", consts.BearerAuth+valString)
			return nil
		}
	}
	// Basic Authentication
	username, userExists := connection["Username"]
	password, passExists := connection["Access Token"]
	if !userExists || !passExists {
		log.Info("failed to set authentication headers")
		return fmt.Errorf("failed to set authentication headers")
	}
	if userString, ok := username.(string); ok {
		if passString, ok := password.(string); ok {
			data := []byte(fmt.Sprintf("%s:%s", userString, passString))
			hashed := base64.StdEncoding.EncodeToString(data)
			request.Header.Set("AUTHORIZATION", consts.BasicAuth+hashed)
			return nil
		}
	}
	log.Info("failed to set authentication headers")
	return fmt.Errorf("failed to set authentication headers")
}

func handleGCPConnection(connection map[string]interface{}, request *http.Request) error {
	if token, ok := connection["Token"].(string); ok {
		request.Header.Set("AUTHORIZATION", consts.BearerAuth+token)
		return nil
	}

	return handleServiceAccountAuth(connection, request)
}

func handleServiceAccountAuth(connection map[string]interface{}, request *http.Request) error {
	credentialsString, ok := connection["credentials"].(string)
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
	defer res.Body.Close()
	if err != nil {
		return errors.Errorf("could not execute the http request: %v", err)
	}

	accessToken, err := extractAccessToken(res)
	if err != nil {
		return err
	}
	request.Header.Set("AUTHORIZATION", consts.BearerAuth+accessToken)

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

func handleAzureConnection(connection map[string]interface{}, request *http.Request) error {
	// handle oauth
	if token, ok := connection["Token"]; ok {
		if tokenString, ok := token.(string); ok {
			request.Header.Set("AUTHORIZATION", "Bearer "+tokenString)
		} else {
			return fmt.Errorf("unable to convert OAuth token to string")
		}
		return nil
	}

	queryParams := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {connection["app_id"].(string)},
		"client_secret": {connection["client_secret"].(string)},
		"resource":      {"https://management.core.windows.net/"},
	}

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", connection["tenant_id"]), strings.NewReader(queryParams.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var responseBody struct {
		AccessToken string `json:"access_token"`
	}

	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return err
	}
	request.Header.Set("AUTHORIZATION", "Bearer "+responseBody.AccessToken)
	return nil
}

func handleGenericConnection(connection map[string]interface{}, request *http.Request, prefixes headerValuePrefixes, headerAlias headerAlias) error {
	headers := make(map[string]string)
	for header, headerValue := range connection {
		if headerValueString, ok := headerValue.(string); ok {
			header = strings.ToUpper(header)
			// if the header is in our alias map replace it with the value in the map
			// TOKEN -> AUTHORIZATION
			if val, ok := headerAlias[header]; ok {
				header = strings.ToUpper(val)
			}

			// we want to help the user by adding prefixes he might have missed
			// for example:   Bearer <TOKEN>
			if val, ok := prefixes[header]; ok {
				if !strings.HasPrefix(headerValueString, val) { // check what prefix the user doesn't have
					// add the prefix
					headerValueString = val + headerValueString
				}
			}

			// If the user supplied BOTH username and password
			// Username:Password pair should be base64 encoded
			// and sent as "Authorization: base64(user:pass)"
			headers[header] = headerValueString
			if username, ok := headers[consts.BasicAuthUsername]; ok {
				if password, ok := headers[consts.BasicAuthPassword]; ok {
					header, headerValueString = "Authorization", constructBasicAuthHeader(username, password)
					cleanRedundantHeaders(&request.Header)
				}
			}
			request.Header.Set(header, headerValueString)
		}
	}
	return nil
}

func constructBasicAuthHeader(username, password string) string {
	data := []byte(fmt.Sprintf("%s:%s", username, password))
	hashed := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("%s%s", consts.BasicAuth, hashed)
}

func cleanRedundantHeaders(requestHeaders *http.Header) {
	requestHeaders.Del(consts.BasicAuthUsername)
	requestHeaders.Del(consts.BasicAuthPassword)
}
