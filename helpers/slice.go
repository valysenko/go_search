package helpers

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func HasAny(realValues, expectedValues []string) bool {
	set := make(map[string]struct{})
	for _, v := range realValues {
		set[v] = struct{}{}
	}

	for _, v := range expectedValues {
		if _, exists := set[v]; exists {
			return true
		}
	}
	return false
}
