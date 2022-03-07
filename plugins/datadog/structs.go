package datadog

type IncidentAttributes struct {
	CustomerImpactEnd   string                 `json:"customer_impact_end,omitempty"`
	CustomerImpactScope string                 `json:"customer_impact_scope,omitempty"`
	CustomerImpactStart string                 `json:"customer_impact_start,omitempty"`
	CustomerImpacted    bool                   `json:"customer_impacted,omitempty"`
	Detected            string                 `json:"detected,omitempty"`
	Fields              map[string]interface{} `json:"fields,omitempty"`
	Resolved            string                 `json:"resolved,omitempty"`
	Title               string                 `json:"title,omitempty"`
}

type CommanderUserData struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type CommanderUser struct {
	Data CommanderUserData `json:"data"`
}

type IncidentRelationships struct {
	CommanderUser CommanderUser `json:"commander_user,omitempty"`
}

type IncidentData struct {
	Id            string                `json:"id,omitempty"`
	Type          string                `json:"type"`
	Attributes    IncidentAttributes    `json:"attributes"`
	Relationships IncidentRelationships `json:"relationships,omitempty"`
}

type IncidentRequestBody struct {
	Data IncidentData `json:"data"`
}
