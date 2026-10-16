package model

type DestinationConfig struct {
	BaseResource
	Dst        string   `json:"dst"`
	Type       string   `json:"type"`
	Ports      []int    `json:"ports,omitempty"`
	PortRanges []string `json:"port_ranges,omitempty"`
	Comment    string   `json:"comment"`
}

type DestinationConfigInput struct {
	Name       string    `json:"name"`
	Dst        string    `json:"dst"`
	Type       string    `json:"type"`
	Ports      []int     `json:"ports"`
	PortRanges *[]string `json:"port_ranges,omitempty"`
	Comment    string    `json:"comment"`
}
