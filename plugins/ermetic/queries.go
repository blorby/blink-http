package ermetic

const (
	listFindingsQuery = `query listFindings($Skip: Int!, $Take: Int! $StartTime: DateTime!, $EndTime: DateTime!, $Status: [FindingStatus!]!, $Severity: [Severity!]!){
	Findings(
		skip: $Skip
		take: $Take,
		filter: {
    Statuses: $Status,
    CreationTimeStart: $StartTime,
    CreationTimeEnd: $EndTime,
		Severities: $Severity
  }) {
    items {
      Title
			Status
			Severity
			CreationTime
			Provider: CloudProvider
			ID: Id
			Description
    }
		pageInfo {
      hasNextPage
      hasPreviousPage
    }
  }
	
}`

	listEntitiesQuery = `query listEntities($Skip: Int!, $Take: Int! $CloudProvider: [CloudProvider!], $Types: [EntityType!]){
	Entities(
		skip: $Skip
		take: $Take,
		filter: {
    Types: $Types,
    CloudProviders: $CloudProvider,
  }) {
    items {
			Name
			Tags{
				Key
				Value
			}
			Provider: CloudProvider
			ID: Id

    }
		pageInfo {
      hasNextPage
      hasPreviousPage
    }
  }
	
}
`
)
