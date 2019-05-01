package util

//GetUniqueStrings returns the given slice of strings without the duplicates
func GetUniqueStrings(in []string) []string {
	out := []string{}

	for _, str := range in {
		shouldAdd := true
		for _, alreadyAdded := range out {
			if alreadyAdded == str {
				shouldAdd = false
			}
		}
		if shouldAdd {
			out = append(out, str)
		}
	}
	return out
}
