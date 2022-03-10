package jira

type SearchResult struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Search struct {
	Values []SearchResult `json:"values"`
}
