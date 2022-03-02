package openapi

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type MaskTestSuite struct {
	suite.Suite
	Mask Mask
}

// Need to populate the global variables.
func (suite *MaskTestSuite) SetupSuite() {
	mask := []byte("actions:\n  AddTeamMember:\n    parameters:\n      teamId:\n        alias: \"Team ID\"\n      userId:\n        alias: \"User ID\"\n  InviteOrgMember:\n    parameters:\n      name:\n        alias: \"Name\"\n      loginOrEmail:\n        alias: \"Username/Email\"\n      role:\n        alias: \"Role\"\n      sendEmail:\n        alias: \"Send Email\"\n  RemoveOrgMember:\n    parameters:\n      userId:\n        alias: \"User ID\"\n  CreateFolder:\n    parameters:\n      uid:\n        alias: \"Folder ID (optional)\"\n      title:\n        alias: \"Folder Name\"\n  CreateDashboard:\n    parameters:\n      dashboard__uid:\n        alias: \"Dashboard Unique ID\"\n      dashboard__title:\n        alias: \"Dashboard Title\"\n      dashboard__tags:\n        alias: \"Dashboard Tags\"\n      dashboard__timezone:\n        alias: \"Dashboard Timezone\"\n      dashboard__schemaVersion:\n        alias: \"Dashboard Schema Version\"\n      dashboard__version:\n        alias: \"Dashboard Version\"\n      dashboard__refresh:\n        alias: \"Dashboard Refresh Interval\"\n      folderUid:\n        alias: \"Folder Unique ID\"\n      message:\n        alias: \"Commit Message\"")
	err := os.WriteFile("mask_test.yaml", mask, 0644)
	suite.Require().Nil(err)
	m, err := ParseMask("mask_test.yaml")
	if err != nil {
		panic("unable to parse mask from test data")
	}
	suite.Mask = m
}

func (suite *MaskTestSuite) TestBuildActionAliasMap() {
	suite.Mask.buildActionAliasMap()
	assert.Equal(suite.T(), len(suite.Mask.ReverseActionAliasMap), 0)
}

func (suite *MaskTestSuite) TestBuildParamAliasMap() {
	suite.Mask.buildParamAliasMap()
	reverseParameterAliasMap := suite.Mask.ReverseParameterAliasMap
	assert.Equal(suite.T(), len(reverseParameterAliasMap), 5)
	assert.Contains(suite.T(), reverseParameterAliasMap, "AddTeamMember")
	assert.Contains(suite.T(), reverseParameterAliasMap["AddTeamMember"], "Team ID")
	assert.Contains(suite.T(), reverseParameterAliasMap["AddTeamMember"], "User ID")
}

func (suite *MaskTestSuite) TestReplaceActionAlias() {
	input := "InviteOrgMember"
	result := suite.Mask.ReplaceActionAlias(input)
	assert.Equal(suite.T(), result, input)
}

func (suite *MaskTestSuite) TestGetAction() {
	action := suite.Mask.GetAction("CreateFolder")
	assert.Equal(suite.T(), action.Alias, "")
	assert.Equal(suite.T(), len(action.Parameters), 2)
	assert.Equal(suite.T(), action.Parameters["uid"].Alias, "Folder ID (optional)")
}

func (suite *MaskTestSuite) TestGetParameter() {
	actionParameter := suite.Mask.GetParameter("CreateFolder", "title")
	assert.Equal(suite.T(), actionParameter.Alias, "Folder Name")
}

func (suite *MaskTestSuite) TearDownSuite() {
	err := os.Remove("mask_test.yaml")
	suite.Require().Nil(err)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestMaskSuite(t *testing.T) {
	suite.Run(t, new(MaskTestSuite))
}
