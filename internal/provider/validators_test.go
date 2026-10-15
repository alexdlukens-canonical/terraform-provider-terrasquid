package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPortRangeValidator(t *testing.T) {
	for _, value := range []string{"1-1", "1-65535", "65535-65535", "30000-32767"} {
		t.Run(value, func(t *testing.T) {
			var resp validator.StringResponse
			portRangeValidator{}.ValidateString(context.Background(), validator.StringRequest{
				Path: path.Root("port_ranges").AtListIndex(0), ConfigValue: types.StringValue(value),
			}, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("valid range rejected: %v", resp.Diagnostics)
			}
		})
	}
	for _, value := range []string{"", "80", "0-80", "80-65536", "443-80", "-1-80", "80-90-100", " 80-90", "80-90 ", "80-90\n", "８０-９０"} {
		t.Run(value, func(t *testing.T) {
			var resp validator.StringResponse
			portRangeValidator{}.ValidateString(context.Background(), validator.StringRequest{
				Path: path.Root("port_ranges").AtListIndex(0), ConfigValue: types.StringValue(value),
			}, &resp)
			if !resp.Diagnostics.HasError() {
				t.Fatal("invalid range accepted")
			}
		})
	}
	for _, value := range []types.String{types.StringNull(), types.StringUnknown()} {
		var resp validator.StringResponse
		portRangeValidator{}.ValidateString(context.Background(), validator.StringRequest{ConfigValue: value}, &resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("unset or unknown value rejected: %v", resp.Diagnostics)
		}
	}
}
func TestInt64SliceToIntSlice(t *testing.T) {
	input := []int64{1, 2, 3}
	output := int64SliceToIntSlice(input)
	if len(output) != 3 || output[0] != 1 || output[1] != 2 || output[2] != 3 {
		t.Errorf("expected [1 2 3], got %v", output)
	}
}

func TestIntSliceToInt64Slice(t *testing.T) {
	input := []int{1, 2, 3}
	output := intSliceToInt64Slice(input)
	if len(output) != 3 || output[0] != 1 || output[1] != 2 || output[2] != 3 {
		t.Errorf("expected [1 2 3], got %v", output)
	}
}
