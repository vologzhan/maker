package slices

func Delete[T comparable](slice []T, needle T) []T {
	for i := range slice {
		if needle == slice[i] {
			slice = append(slice[:i], slice[i+1:]...)
			break
		}
	}
	return slice
}
