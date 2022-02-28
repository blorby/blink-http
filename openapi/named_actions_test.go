package openapi

import (
	"github.com/stretchr/testify/suite"
	"os"
	"reflect"
	"testing"
)



type NamedActionsTestSuite struct {
	suite.Suite
	openApi ParsedOpenApi
}

func (suite *NamedActionsTestSuite) SetupSuite() {
	// this is the grafana OpenApi file with just the 'create dashboard' action
	grafanaOpenApi := []byte("openapi: 3.1.0\ninfo:\n  title: Grafana REST API\n  description: Grafana REST API.\n  version: '1.0'\nservers:\n  - url: 'http://localhost:3000'\npaths:\n  /api/dashboards/db:\n    post:\n      summary: 'Creates a new dashboard or updates an existing dashboard. When updating existing dashboards, if you do not define the folderId or the folderUid property, then the dashboard(s) are moved to the General folder. (You need to define only one property, not both).'\n      responses:\n        '200':\n          description: OK\n          content:\n            application/json:\n              schema:\n                description: ''\n                type: object\n                properties:\n                  id:\n                    type: number\n                  uid:\n                    type: string\n                    minLength: 1\n                  url:\n                    type: string\n                    minLength: 1\n                  status:\n                    type: string\n                    minLength: 1\n                  version:\n                    type: number\n                  slug:\n                    type: string\n                    minLength: 1\n                required:\n                  - id\n                  - uid\n                  - url\n                  - status\n                  - version\n                  - slug\n                x-examples:\n                  example-1:\n                    id: 1\n                    uid: cIBgcSjkk\n                    url: /d/cIBgcSjkk/production-overview\n                    status: success\n                    version: 1\n                    slug: production-overview\n      operationId: CreateDashboard\n      requestBody:\n        required: true\n        content:\n          application/json:\n            schema:\n              description: Folder details\n              type: object\n              properties:\n                dashboard:\n                  description: 'The complete dashboard model, id = null to create a new dashboard.'\n                  type: object\n                  properties:\n                    id:\n                      type: integer\n                      description: null to create a new dashboard.\n                    uid:\n                      type: string\n                      format: uuid\n                      description: Optional unique identifier when creating a dashboard. uid = null will generate a new uid.\n                    title:\n                      type: string\n                      minLength: 1\n                    tags:\n                      type: array\n                      items:\n                        required: []\n                        properties: {}\n                    timezone:\n                      type: string\n                      minLength: 1\n                    schemaVersion:\n                      type: number\n                    version:\n                      type: number\n                    refresh:\n                      description: 'Set the dashboard refresh interval. If this is lower than the minimum refresh interval, then Grafana will ignore it and will enforce the minimum refresh interval.'\n                      type: string\n                      minLength: 1\n                  required:\n                    - title\n                folderId:\n                  description: 'The id of the folder to save the dashboard in. When updating existing dashboards, if you do not define the folderId or the folderUid property, then the dashboard(s) are moved to the General folder. (You need to define only one property, not both).'\n                  type: number\n                folderUid:\n                  description: 'The UID of the folder to save the dashboard in. Overrides the folderId. When updating existing dashboards, if you do not define the folderId or the folderUid property, then the dashboard(s) are moved to the General folder. (You need to define only one property, not both).'\n                  type: string\n                  minLength: 1\n                message:\n                  description: Set a commit message for the version history.\n                  type: string\n                  minLength: 1\n                overwrite:\n                  description: 'Set to true if you want to overwrite existing dashboard with newer version, same dashboard title in folder or same dashboard uid.'\n                  type: boolean\n              required:\n                - dashboard\n              x-examples:\n                example-1:\n                  dashboard:\n                    id: null\n                    uid: null\n                    title: Production Overview\n                    tags:\n                      - templated\n                    timezone: browser\n                    schemaVersion: 16\n                    version: 0\n                    refresh: 25s\n                  folderId: 0\n                  folderUid: nErXDvfCkzz\n                  message: Made changes to xyz\n                  overwrite: false\ncomponents:\n  schemas: {}\n  securitySchemes:\n    BearerAuth:\n      scheme: bearer\n      type: http\nsecurity:\n  - BearerAuth: []\n")
	mask := []byte("request_url: grafana\nactions:\n  CreateDashboard:\n    alias: \"CreateOrUpdateDashboard\"\n    parameters:\n      dashboard_uid:\n        alias: \"Dashboard Unique ID\"\n      dashboard_title:\n        alias: \"Dashboard Title\"\n      dashboard_tags:\n        alias: \"Dashboard Tags\"\n      dashboard_timezone:\n        alias: \"Dashboard Timezone\"\n      dashboard_schemaVersion:\n        alias: \"Dashboard Schema Version\"\n      dashboard_version:\n        alias: \"Dashboard Version\"\n      dashboard_refresh:\n        alias: \"Dashboard Refresh Interval\"\n      folderUid:\n        alias: \"Folder Unique ID\"\n      message:\n        alias: \"Commit Message\"\n")

	err := os.WriteFile("openapi.yaml", grafanaOpenApi, 0644)
	suite.Require().Nil(err)
	err = os.WriteFile("m.yaml", mask, 0644)
	suite.Require().Nil(err)
	maskData, err := ParseMask("m.yaml")
	suite.Require().Nil(err)
	openApi, err := parseOpenApiFile("openapi.yaml")
	suite.Require().Len(openApi.actions, 1)
	openApi.mask = maskData
	suite.Require().Nil(err)
	suite.openApi = openApi
}

func (suite *NamedActionsTestSuite) TestGetActionParams() {
	actionParams := suite.openApi.getActionParams(suite.openApi.actions[0])
	suite.Require().Len(actionParams, 9)
	_, ok := actionParams["dashboard_title"]
	suite.Require().True(ok)
}

func (suite *NamedActionsTestSuite) TestGetActionParamValues() {
	action := suite.openApi.actions[0]
	originalName := suite.openApi.mask.ReplaceActionAlias(action.Name)
	opDefinition := suite.openApi.operations[originalName]
	actionParamValues := suite.openApi.getActionParamValues(action, opDefinition)
	expected := map[string]string{
		"body": "{\"dashboard\":{\"id\":\"{{int(params.dashboard_id)}}\",\"refresh\":\"{{params.dashboard_refresh}}\",\"schemaVersion\":\"{{int(params.dashboard_schemaVersion)}}\",\"tags\":\"[\\\"{{arr(params.dashboard_tags)}}\\\"]\",\"timezone\":\"{{params.dashboard_timezone}}\",\"title\":\"{{params.dashboard_title}}\",\"uid\":\"{{params.dashboard_uid}}\",\"version\":\"{{int(params.dashboard_version)}}\"},\"folderId\":\"{{int(params.folderId)}}\",\"folderUid\":\"{{params.folderUid}}\",\"message\":\"{{params.message}}\",\"overwrite\":\"{{bool(params.overwrite)}}\"}",
		"contentType": "application/json",
		"url": "{{connection.grafana.REQUEST_URL + '/api/dashboards/db'}}",
	}
	suite.Require().True(reflect.DeepEqual(actionParamValues, expected))
}

func (suite *NamedActionsTestSuite) TestGetDisplayName() {
	type TestCase struct {
		Name string
		Wanted string
	}
	for _, testCase := range []TestCase{
		{
			Name: "NameWithNoSpaces",
			Wanted: "Name With No Spaces",
		},
		{
			Name: "name.with.dots",
			Wanted: "Name With Dots",
		},
		{
			Name: "name_With_underscores",
			Wanted: "Name With Underscores",
		},
		{
			Name: "[]Mix_things.up url and ids",
			Wanted: "Mix Things Up URL And IDs",
		},
	} {
		suite.Require().Equal(testCase.Wanted,getDisplayName(testCase.Name))
	}
}


func (suite *NamedActionsTestSuite) TearDownSuite() {
	err := os.Remove("m.yaml")
	suite.Require().Nil(err)
	err = os.Remove("openapi.yaml")
	suite.Require().Nil(err)
}


func TestNamedActionsSuite(t *testing.T) {
	suite.Run(t, new(NamedActionsTestSuite))
}
