package utils

func MapToValuesSlice[K comparable, V any](m map[K]V) []V{
	slice := make([]V, 0, len(m))
	for _, value := range m {
		slice = append(slice, value)
	}
	
	return slice
}
