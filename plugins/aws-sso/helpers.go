package aws_sso

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blinkops/blink-http/implementation/requests"
	"github.com/blinkops/blink-http/plugins/types"
	"github.com/blinkops/blink-sdk/plugin"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

func findUserByEmail(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin, requestUrl string, email string) (*AwsUser, error){
	filter := fmt.Sprintf("userName eq \"%s\"", email)

	resp, err := requests.SendRequest(ctx, plugin, http.MethodGet, requestUrl+"/Users?filter="+url.QueryEscape(filter), request.Timeout, nil, nil, nil)
	if err != nil {
		if resp.Body != nil {
			return nil, errors.New(string(resp.Body))
		}
		return nil, err
	}

	var res UserFilterResults
	err = json.Unmarshal(resp.Body, &res)
	if err != nil {
		return nil, err
	}

	if res.TotalResults != 1 {
		return nil, errors.New("user not found")
	}

	return &res.Resources[0], nil
}

func addUserToGroup(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin, requestUrl string, groupId string, userId string) (string, error) {
	body := &GroupMemberChange{
		Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		Operations: []GroupMemberChangeOperation{
			{
				Operation: "add",
				Path:      "members",
				Members: []GroupMemberChangeMember{
					{Value: userId},
				},
			},
		},
	}

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		return "", errors.New("failed to marshal add user to group request")
	}

	resp, err := requests.SendRequest(ctx, plugin, http.MethodPatch, fmt.Sprintf(requestUrl+"/Groups/%s", groupId), request.Timeout, nil, nil, marshalledBody)
	if err != nil {
		if resp != nil {
			if resp.StatusCode == http.StatusBadRequest {
				return "user not found", nil
			}
			if resp.StatusCode == http.StatusUnauthorized {
				return "unauthorized", nil
			}
		}
		return "", err
	}

	if resp.StatusCode == http.StatusNoContent {
		return "success", nil
	}

	return "failed to add user to group", nil
}

func findGroupByDisplayName(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin, displayName string) (*Group, error) {
	requestUrl, err := requests.GetRequestUrl(ctx, PluginName)
	if err != nil {
		return nil, errors.New("no request url provided")
	}

	filter := fmt.Sprintf("displayName eq \"%s\"", displayName)

	resp, err := requests.SendRequest(ctx, plugin, http.MethodGet, requestUrl+"/Groups?filter="+url.QueryEscape(filter), request.Timeout, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	var r GroupFilterResults
	err = json.Unmarshal(resp.Body, &r)
	if err != nil {
		return nil, err
	}

	if r.TotalResults != 1 {
		return nil, errors.New("group not found")
	}

	return &r.Resources[0], nil
}

func performCreateUser(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin, requestUrl string, userInfo UserInfo) (*AwsUser, error) {
	user := AwsUser{
		Schemas:  []string{"urn:ietf:params:scim:schemas:core:2.0:AwsUser"},
		Username: userInfo.Email,
		Name: struct {
			FamilyName string `json:"familyName"`
			GivenName  string `json:"givenName"`
		}{
			FamilyName: userInfo.Name.FamilyName,
			GivenName:  userInfo.Name.GivenName,
		},
		DisplayName: fmt.Sprintf("%s %s", userInfo.Name.GivenName, userInfo.Name.FamilyName),
		Active:      userInfo.Active,
		Emails: []UserEmail{
			{
				Value:   userInfo.Email,
				Type:    "work",
				Primary: true,
			},
		},
	}

	marshalledUser, err := json.Marshal(user)
	if err != nil {
		return nil, errors.New("failed to marshal json")
	}

	resp, err := requests.SendRequest(ctx, plugin, http.MethodPost, requestUrl+"/Users", request.Timeout, nil, nil, marshalledUser)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp.Body, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func performDeleteUser(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin, requestUrl string, id string) error {
	resp, err := requests.SendRequest(ctx, plugin, http.MethodDelete, fmt.Sprintf("%s/Users/%s", requestUrl, id), request.Timeout, nil, nil, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return errors.New("failed to delete user")
	}

	return nil
}

func performAddUsersToGroup(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin, requestUrl string, groupId string, users string) (chan AddUserToGroupResponse, error){
	users = strings.ReplaceAll(users, ", ", ",")
	emails := strings.Split(users, ",")

	statuses := make(chan AddUserToGroupResponse, len(emails))
	var errorToReturn error

	var wg sync.WaitGroup
	for _, email := range emails {
		emailVal := email
		wg.Add(1)
		go func() {
			defer wg.Done()
			user, err := findUserByEmail(ctx, request, plugin, requestUrl, emailVal)

			if err != nil {
				status := AddUserToGroupResponse{User: emailVal, Status: err.Error()}
				statuses <- status
				return
			}

			statusStr, err := addUserToGroup(ctx, request, plugin, requestUrl, groupId, user.ID)
			status := AddUserToGroupResponse{User: emailVal, Status: statusStr}
			statuses <- status
			if err != nil {
				errorToReturn = err
			}
		}()
	}
	wg.Wait()

	close(statuses)

	return statuses, errorToReturn
}

func performCreateAndAddUsersToGroup(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin, requestUrl string, groupId string, usersArr []UserInfo) (chan AddUserToGroupResponse, error) {
	statuses := make(chan AddUserToGroupResponse, len(usersArr))
	var errorToReturn error

	defer close(statuses)

	var wg sync.WaitGroup
	for _, user := range usersArr {
		userVal := user
		wg.Add(1)
		go func() {
			defer wg.Done()
			created := false
			awsUser, err := findUserByEmail(ctx, request, plugin, requestUrl, userVal.Email)
			if err != nil {
				if err.Error() != "user not found" {
					status := AddUserToGroupResponse{User: userVal.Email, Status: err.Error()}
					statuses <- status
					return
				}

				userVal.Active = true
				awsUser, err = performCreateUser(ctx, request, plugin, requestUrl, userVal)
				if err != nil {
					status := AddUserToGroupResponse{User: userVal.Email, Status: err.Error()}
					statuses <- status
					return
				}

				created = true
			}

			statusStr, err := addUserToGroup(ctx, request, plugin, requestUrl, groupId, awsUser.ID)
			if created {
				statusStr += ", created new user"
			}
			status := AddUserToGroupResponse{User: userVal.Email, Status: statusStr}
			statuses <- status
			if err != nil {
				errorToReturn = err
			}
		}()
	}
	wg.Wait()

	return statuses, errorToReturn
}

func createAndAddUsersToGroup(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest, plugin types.Plugin, requestUrl string, groupId string, users string) (chan AddUserToGroupResponse, error){
	var usersArr []UserInfo
	err := json.Unmarshal([]byte(users), &usersArr)
	if err != nil {
		return nil, errors.New("failed to unmarshal users param")
	}

	// validate that all users have an email address
	for _, user := range usersArr {
		if user.Email == "" {
			return nil, errors.New("not all users have an email address")
		}
	}

	return performCreateAndAddUsersToGroup(ctx, request, plugin, requestUrl, groupId, usersArr)
}

