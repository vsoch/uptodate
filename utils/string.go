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
