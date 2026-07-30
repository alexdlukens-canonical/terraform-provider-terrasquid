package provider

import (
	"fmt"
	"slices"

	"github.com/terrasquid/terraform-provider-terrasquid/internal/model"
)

func validateCreatedSourceACL(input model.SourceACLInput, result *model.SourceACL) error {
	if input.Name != result.Name || input.Comment != result.Comment || !sameStrings(input.CIDR, result.CIDR) {
		return fmt.Errorf("an existing source ACL named %q has different attributes; import it or choose a different name", input.Name)
	}
	return nil
}

func validateCreatedDestinationConfig(input model.DestinationConfigInput, result *model.DestinationConfig) error {
	if input.Name != result.Name || input.Dst != result.Dst || input.Type != result.Type || input.Comment != result.Comment || !sameInts(input.Ports, result.Ports) {
		return fmt.Errorf("an existing destination config named %q has different attributes; import it or choose a different name", input.Name)
	}
	return nil
}

func validateCreatedDestinationGroup(input model.DestinationGroupInput, result *model.DestinationGroup) error {
	if input.Name != result.Name || input.Comment != result.Comment || !sameStrings(input.Destinations, result.Destinations) {
		return fmt.Errorf("an existing destination group named %q has different attributes; import it or choose a different name", input.Name)
	}
	return nil
}

func validateCreatedACLRule(input model.ACLRuleInput, result *model.ACLRule) error {
	if input.Name != result.Name || input.Priority != result.Priority || input.Comment != result.Comment ||
		!sameStrings(input.Sources, result.Sources) || !sameStrings(input.Destinations, result.Destinations) ||
		!sameStrings(input.DestinationGroups, result.DestinationGroups) {
		return fmt.Errorf("an existing ACL rule named %q has different attributes; import it or choose a different name", input.Name)
	}
	return nil
}

func sameStrings(left, right []string) bool {
	left = slices.Clone(left)
	right = slices.Clone(right)
	slices.Sort(left)
	slices.Sort(right)
	return slices.Equal(left, right)
}

func sameInts(left, right []int) bool {
	left = slices.Clone(left)
	right = slices.Clone(right)
	slices.Sort(left)
	slices.Sort(right)
	return slices.Equal(left, right)
}
