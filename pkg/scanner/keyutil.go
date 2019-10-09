package scanner

func ContainsKeys(a map[string]string, b map[string]bool) bool {
	for k := range a {
		_, ok := b[k]
		if !ok {
			return false
		}
	}
	return true
}
