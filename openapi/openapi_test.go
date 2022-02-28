package openapi

import (
	"encoding/json"
	ops "github.com/blinkops/blink-http/openapi/operations"
	plugin_sdk "github.com/blinkops/blink-sdk/plugin"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

var (
	myPlugin *ParsedOpenApi
	schemaByte = []byte(`{
 "description": "Folder details",
 "properties": {
  "dashboard": {
   "description": "dashboard description",
   "properties": {
    "id": {},
    "refresh": {
     "minLength": 1,
     "type": "string"
    },
    "schemaVersion": {
     "type": "number"
    },
    "tags": {
     "items": {},
     "type": "array"
    },
    "timezone": {
	"description": "my test description",
     "minLength": 1,
     "type": "string"
    },
    "title": {
     "minLength": 1,
     "type": "string"
    },
    "uid": {},
    "version": {
     "type": "number"
    }
   },
   "required": [
    "title",
    "tags",
    "timezone",
    "schemaVersion",
    "version",
    "refresh"
   ],
   "type": "object"
  },
  "folderId": {
   "type": "number"
  },
  "folderUid": {
   "minLength": 1,
   "type": "string"
  },
  "message": {
   "minLength": 1,
   "type": "string"
  },
  "overwrite": {
   "type": "boolean"
  }
 },
 "required": [
  "dashboard",
  "folderId",
  "folderUid",
  "message",
  "overwrite"
 ],
 "type": "object",
 "x-examples": {
  "example-1": {
   "dashboard": {
    "id": null,
    "refresh": "25s",
    "schemaVersion": 16,
    "tags": [
     "templated"
    ],
    "timezone": "browser",
    "title": "Production Overview",
    "uid": null,
    "version": 0
   },
   "folderId": 0,
   "folderUid": "nErXDvfCkzz",
   "message": "Made changes to xyz",
   "overwrite": false
  }
 }
}`)
)

type OpenApiTestSuite struct {
	suite.Suite
}

func (suite *OpenApiTestSuite) SetupSuite() {
	myPlugin = &ParsedOpenApi{
		requestUrl: "http://127.0.0.1:8888",

		actions: []plugin_sdk.Action{
			{
				Name:        "AddTeamMember",
				Description: "AddTeamMember",
				Enabled:     true,
				EntryPoint:  "/api/teams/{teamId}/members",
				Parameters: map[string]plugin_sdk.ActionParameter{
					"Team ID": {
						Type:        "integer",
						Description: "Team ID to add member to",
						Placeholder: "",
						Required:    true,
						Default:     "",
						Pattern:     "",
						Options:     []string{},
					},
				},
				Output: nil,
			},
			{
				Name:        "InviteOrgMember",
				Description: "InviteOrgMember",
				Enabled:     true,
				EntryPoint:  "/api/org/invites",
				Parameters: map[string]plugin_sdk.ActionParameter{
					"Name": {
						Type:        "integer",
						Description: "User to invite",
						Placeholder: "",
						Required:    true,
						Default:     "",
						Pattern:     "",
						Options:     []string{},
					},
				},
				Output: nil,
			},
		},
		description: "description",
	}
}

func (suite *OpenApiTestSuite) TestHandleBodyParams() {
	var schema *openapi3.Schema
	_ = json.Unmarshal(schemaByte, &schema)
	parentPath, schemaPath := "", ""
	action := myPlugin.actions[0]

	handleBodyParams(&action, schema, parentPath, schemaPath, true)

	assert.Equal(suite.T(), 13, len(myPlugin.actions[0].Parameters))
	assert.Contains(suite.T(), myPlugin.actions[0].Parameters, "dashboard_id")
	assert.Contains(suite.T(), myPlugin.actions[0].Parameters, "dashboard_timezone")
	assert.Equal(suite.T(), "my test description", myPlugin.actions[0].Parameters["dashboard_timezone"].Description)
	assert.Equal(suite.T(), true, myPlugin.actions[0].Parameters["dashboard_timezone"].Required)
}

func (suite *OpenApiTestSuite) TestParseActionParam() {
	var schema *openapi3.Schema
	_ = json.Unmarshal(schemaByte, &schema)
	paramName := "dashboard"
	pathParam := schema.Properties[paramName]

	actionParam := parseActionParam(pathParam, false, pathParam.Value.Description)

	assert.False(suite.T(), actionParam.Required)
	assert.Equal(suite.T(), actionParam.Description, "dashboard description")
}

func (suite *OpenApiTestSuite) TestDefineOperations() {
	// handlers.DefineOperations populates a global variable (OperationDefinitions) that is REQUIRED for this run.
	// The only convenient option for populating this var is to load an api from file
	// the other one - loading from file is just too inconvenient
	openApi, err := loadOpenApi("https://raw.githubusercontent.com/blinkops/blink-grafana/master/grafana-openapi.yaml")
	suite.Require().Nil(err)

	opsDef, err := ops.DefineOperations(openApi)
	suite.Require().Nil(err)
	suite.Require().NotZero(opsDef)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestPluginSuite(t *testing.T) {
	suite.Run(t, new(OpenApiTestSuite))
}

