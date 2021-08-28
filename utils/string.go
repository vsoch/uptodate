package utils

// includesString to determine if a list include a string
func includesString(lookingFor string, list []string) bool {
	for _, b := range list {
		if b == lookingFor {
			return true
		}
	}
	return false
}
