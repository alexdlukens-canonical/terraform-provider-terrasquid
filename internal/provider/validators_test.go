package provider

import (
	"testing"
)

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
