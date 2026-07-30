package provider

import (
	"testing"

	"github.com/terrasquid/terraform-provider-terrasquid/internal/model"
)

func TestValidateCreatedResources(t *testing.T) {
	sourceInput := model.SourceACLInput{Name: "source", CIDR: []string{"10.0.0.0/8"}, Comment: "test"}
	source := &model.SourceACL{BaseResource: model.BaseResource{Name: "source"}, CIDR: []string{"10.0.0.0/8"}, Comment: "test"}
	if err := validateCreatedSourceACL(sourceInput, source); err != nil {
		t.Fatalf("matching source ACL rejected: %v", err)
	}
	source.CIDR = []string{"192.168.0.0/16"}
	if err := validateCreatedSourceACL(sourceInput, source); err == nil {
		t.Error("mismatched source ACL accepted")
	}

	destinationInput := model.DestinationConfigInput{Name: "destination", Dst: "example.com", Type: "CONNECT", Ports: []int{443}}
	destination := &model.DestinationConfig{BaseResource: model.BaseResource{Name: "destination"}, Dst: "example.com", Type: "CONNECT", Ports: []int{443}}
	if err := validateCreatedDestinationConfig(destinationInput, destination); err != nil {
		t.Fatalf("matching destination config rejected: %v", err)
	}
	destination.Ports = []int{80}
	if err := validateCreatedDestinationConfig(destinationInput, destination); err == nil {
		t.Error("mismatched destination config accepted")
	}

	groupInput := model.DestinationGroupInput{Name: "group", Destinations: []string{"a", "b"}}
	group := &model.DestinationGroup{BaseResource: model.BaseResource{Name: "group"}, Destinations: []string{"b", "a"}}
	if err := validateCreatedDestinationGroup(groupInput, group); err != nil {
		t.Fatalf("matching destination group rejected: %v", err)
	}
	group.Destinations = []string{"a"}
	if err := validateCreatedDestinationGroup(groupInput, group); err == nil {
		t.Error("mismatched destination group accepted")
	}

	ruleInput := model.ACLRuleInput{Name: "rule", Priority: 100, Sources: []string{"a", "b"}, Destinations: []string{"c"}, DestinationGroups: []string{}}
	rule := &model.ACLRule{BaseResource: model.BaseResource{Name: "rule"}, Priority: 100, Sources: []string{"b", "a"}, Destinations: []string{"c"}, DestinationGroups: []string{}}
	if err := validateCreatedACLRule(ruleInput, rule); err != nil {
		t.Fatalf("matching ACL rule rejected: %v", err)
	}
	rule.Priority = 200
	if err := validateCreatedACLRule(ruleInput, rule); err == nil {
		t.Error("mismatched ACL rule accepted")
	}
}
