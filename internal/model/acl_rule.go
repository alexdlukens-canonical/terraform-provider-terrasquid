package model

type ACLRule struct {
	BaseResource
	Priority          int      `json:"priority"`
	Comment           string   `json:"comment"`
	Sources           []string `json:"sources"`
	Destinations      []string `json:"destinations"`
	DestinationGroups []string `json:"destination_groups"`
}

type ACLRuleInput struct {
	Name              string   `json:"name"`
	Priority          int      `json:"priority"`
	Comment           string   `json:"comment"`
	Sources           []string `json:"sources"`
	Destinations      []string `json:"destinations"`
	DestinationGroups []string `json:"destination_groups"`
}
