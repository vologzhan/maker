package slices

import "slices"

func Delete[T comparable](slice []T, needle T) []T {
	for i := range slice {
		if needle == slice[i] {
			return slices.Delete(slice, i, i+1)
		}
	}
	return slice
}
