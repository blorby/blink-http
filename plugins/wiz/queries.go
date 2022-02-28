package wiz

const (
	listCloudConfigurationRulesQuery = `query CloudConfigurationSettingsTable($first: Int, $after: String, $filterBy: CloudConfigurationRuleFilters, $orderBy: CloudConfigurationRuleOrder, $projectId: [String!]) {
          cloudConfigurationRules(
            first: $first
            after: $after
            filterBy: $filterBy
            orderBy: $orderBy
          ) {
            nodes {
              id
              name
              description
              enabled
              severity
              cloudProvider
              subjectEntityType
              functionAsControl
              graphId
              opaPolicy
              builtin
              targetNativeType
              remediationInstructions
              control {
                id
              }
              securitySubCategories {
                id
                title
                description
                externalId
                category {
                  id
                  name
                  description
                  externalId
                  framework {
                    id
                    name
                  }
                }
              }
              analytics(selection: {projectId: $projectId}) {
                passCount
                failCount
              }
              scopeAccounts {
                id
                name
                cloudProvider
              }
            }
            pageInfo {
              endCursor
              hasNextPage
            }
            totalCount
          }
        }`
	listProjectsQuery = `query ProjectsTable($filterBy: ProjectFilters, $first: Int, $after: String, $orderBy: ProjectOrder, $analyticsSelection: ProjectIssueAnalyticsSelection) {
          projects(filterBy: $filterBy, first: $first, after: $after, orderBy: $orderBy) {
            nodes {
              id
              name
              slug
              cloudAccountCount
              repositoryCount
              securityScore
              entityCount
              technologyCount
              archived
              riskProfile {
                businessImpact
              }
              issueAnalytics(selection: $analyticsSelection) {
                issueCount
                scopeSize
                informationalSeverityCount
                lowSeverityCount
                mediumSeverityCount
                highSeverityCount
                criticalSeverityCount
              }
            }
            pageInfo {
              hasNextPage
              endCursor
            }
            totalCount
            LBICount
            MBICount
            HBICount
          }
        }`
	listControlsQuery = `query ManageControlsTable($first: Int, $after: String, $filterBy: ControlFilters, $issueAnalyticsSelection: ControlIssueAnalyticsSelection) {
          controls(filterBy: $filterBy, first: $first, after: $after) {
            nodes {
              id
              name
              description
              type
              severity
              query
              enabled
              lastRunAt
              lastSuccessfulRunAt
              lastRunError
              scopeProject {
                id
                name
              }
              sourceCloudConfigurationRule {
                name
              }
              securitySubCategories {
                id
                externalId
                title
                description
                category {
                  id
                  externalId
                  name
                  framework {
                    id
                    name
                  }
                }
              }
              enabledForLBI
              enabledForMBI
              enabledForHBI
              enabledForUnattributed
              createdBy {
                id
                name
                email
              }
              issueAnalytics(selection: $issueAnalyticsSelection) {
                issueCount
              }
            }
            pageInfo {
              hasNextPage
              endCursor
            }
            totalCount
          }
        }`
)