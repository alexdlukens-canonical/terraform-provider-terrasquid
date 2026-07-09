package provider

import (
	"sort"
)

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
