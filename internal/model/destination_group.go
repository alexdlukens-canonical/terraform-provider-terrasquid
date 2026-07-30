package model

type DestinationGroup struct {
	BaseResource
	Destinations []string `json:"destinations"`
	Comment      string   `json:"comment"`
}

type DestinationGroupInput struct {
	Name         string   `json:"name"`
	Destinations []string `json:"destinations"`
	Comment      string   `json:"comment"`
}
