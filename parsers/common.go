package parsers

// A Result object will store a path to some file that was changed, and
// an identifier for the parser, and some identifier for the changed file

type Result struct {
	Name      string `json:"name,omitempty"`
	Filename  string `json:"filename,omitempty"`
	Parser    string `json:"parser,omitempty"`
	Identifer string `json:"id,omitempty"`
}

// VersionRegex matches a major and minor, optional third group (not semver)
var VersionRegex = "[0-9]+[.][0-9]+(?:[.][0-9]+)?"
