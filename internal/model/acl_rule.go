package model

type ACLRule struct {
	BaseResource
	Priority     int      `json:"priority"`
	Sources      []string `json:"sources"`
	Destinations []string `json:"destinations"`
}

type ACLRuleInput struct {
	Name         string   `json:"name"`
	Priority     int      `json:"priority"`
	Sources      []string `json:"sources"`
	Destinations []string `json:"destinations"`
}
