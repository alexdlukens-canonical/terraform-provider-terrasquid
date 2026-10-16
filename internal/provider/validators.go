package provider

import (
	"context"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var portRangePattern = regexp.MustCompile(`^[0-9]{1,5}-[0-9]{1,5}$`)

type portRangeValidator struct{}

func (portRangeValidator) Description(context.Context) string {
	return "Port ranges must use start-end format with endpoints between 1 and 65535 and start <= end."
}

func (v portRangeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v portRangeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	if portRangePattern.MatchString(value) {
		parts := strings.SplitN(value, "-", 2)
		start, startErr := strconv.Atoi(parts[0])
		end, endErr := strconv.Atoi(parts[1])
		if startErr == nil && endErr == nil && start >= 1 && start <= end && end <= 65535 {
			return
		}
	}
	resp.Diagnostics.AddAttributeError(req.Path, "Invalid Port Range", v.Description(ctx))
}
func int64SliceToIntSlice(in []int64) []int {
	out := make([]int, len(in))
	for i, v := range in {
		out[i] = int(v)
	}
	return out
}

func intSliceToInt64Slice(in []int) []int64 {
	out := make([]int64, len(in))
	for i, v := range in {
		out[i] = int64(v)
	}
	return out
}

func sortIntSlice(in []int) []int {
	out := make([]int, len(in))
	copy(out, in)
	sort.Ints(out)
	return out
}

func sortInt64Slice(in []int64) []int64 {
	out := make([]int64, len(in))
	copy(out, in)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
