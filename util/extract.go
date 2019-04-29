package util

//ExtractStringMap extracts a map[string]interface from a map[string]interface
func ExtractStringMap(in map[string]interface{}, key string) (map[string]interface{}, bool) {
	if in == nil {
		return nil, false
	}
	iOut, ok := in[key]
	if !ok || iOut == nil {
		return nil, false
	}
	out, ok := iOut.(map[string]interface{})
	if !ok {
		return nil, false
	}
	return out, true
}
