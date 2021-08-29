package utils

func GetLogo() string {
	return `              _            _       _       
  _   _ _ __ | |_ ___   __| | __ _| |_ ___ 
 | | | | '_ \| __/ _ \ / _  |/ _  | __/ _ \
 | |_| | |_) | || (_) | (_| | (_| | ||  __/
  \__,_| .__/ \__\___/ \__,_|\__,_|\__\___|
       |_|`
}

// includesString to determine if a list include a string
func IncludesString(lookingFor string, list []string) bool {
	for _, b := range list {
		if b == lookingFor {
			return true
		}
	}
	return false
}

// Return overlap in two
func FindOverlap(one []string, two []string) []string {

	var overlap []string

	// Loop through one, and see if present in two
	for _, string1 := range one {
		if IncludesString(string1, two) {
			overlap = append(overlap, string1)
		}
	}
	return overlap
}

// Return strings that are in first list, but not second
func FindMissingInSecond(one []string, two []string) []string {

	var difference []string

	// Loop through one, and see if present in two
	for _, string1 := range one {

		// It's not found in two
		if !IncludesString(string1, two) {
			difference = append(difference, string1)
		}
	}
	return difference
}
