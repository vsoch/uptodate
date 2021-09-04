package parsers

// A Result object will store a path to some file that was changed, and
// an identifier for the parser, and some identifier for the changed file

type Result struct {
	Name       string `json:"name,omitempty"`
	Filename   string `json:"filename,omitempty"`
	Parser     string `json:"parser,omitempty"`
	Identifier string `json:"id,omitempty"`
}

// A BuildResult needs more information (e.g., versions) to be given to a build matrix
type BuildResult struct {
	Name          string            `json:"name,omitempty"`
	Filename      string            `json:"filename,omitempty"`
	Parser        string            `json:"parser,omitempty"`
	BuildArgs     map[string]string `json:"buildargs,omitempty"`
	CommandPrefix string            `json:"command_prefix,omitempty"`
}

// BuildVariable holds a key (name) and one or more values to parameterize over
type BuildVariable struct {
	Name   string
	Values []string
}

// VersionRegex matches a major and minor, optional third group (not semver)
var VersionRegex = "[0-9]+[.][0-9]+(?:[.][0-9]+)?"
