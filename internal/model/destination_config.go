package model

type DestinationConfig struct {
	BaseResource
	Dst     string `json:"dst"`
	Type    string `json:"type"`
	Ports   []int  `json:"ports,omitempty"`
	Comment string `json:"comment"`
}

type DestinationConfigInput struct {
	Name    string `json:"name"`
	Dst     string `json:"dst"`
	Type    string `json:"type"`
	Ports   []int  `json:"ports"`
	Comment string `json:"comment"`
}
