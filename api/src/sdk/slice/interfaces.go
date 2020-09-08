package slice

// StringSliceToInterfaceSlice transform a slice of strings to a slice of interfaces
// It is particularly useful when building SQLBoiler WhereIn query
// because the function only takes slice of interfaces as argument
func StringSliceToInterfaceSlice(x []string) []interface{} {
	result := make([]interface{}, len(x))
	for index, elt := range x {
		result[index] = elt
	}
	return result
}
