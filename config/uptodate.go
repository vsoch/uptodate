package config

import (
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

type Container struct {
	Name     string   `yaml:"name"`
	Filter   []string `yaml:"filter,omitempty"`
	StartAt  string   `yaml:"startat,omitempty"`
	EndAt    string   `yaml:"endat,omitempty"`
	Skips    []string `yaml:"skips,omitempty"`
	Includes []string `yaml:"includes,omitempty"`
}

type DockerHierarchy struct {
	Container Container `yaml:"container"`
}

// DockerBuild holds one or more build args
// We'd have to separate these later anyway
type DockerBuild struct {
	BuildArgs map[string]BuildArg `yaml:"build_args"`
	Matrix    map[string][]string `yaml:"matrix,omitempty"`
	Active    bool `yaml:"active,omitempty"`
}

// BuildArg expects metadata for building from a variable, spack, or other
type BuildArg struct {
	Name     string            `yaml:"name,omitempty"`
	Key      string            `yaml:"key,omitempty"`
	Type     string            `yaml:"type,omitempty"`
	StartAt  string            `yaml:"startat,omitempty"`
	EndAt    string            `yaml:"endat,omitempty"`
	Versions []string          `yaml:"versions,omitempty"`
	Values   []string          `yaml:"values,omitempty"`
	Filter   []string          `yaml:"filter,omitempty"`
	Skips    []string          `yaml:"skips,omitempty"`
	Includes []string          `yaml:"includes,omitempty"`
	Params   map[string]string `yaml:"params,omitempty"`
}

// Get the identifier for a build arg
func (b *BuildArg) GetKey() string {
	if b.Key != "" {
		return b.Key
	}
	return b.Name
}

// read the config and return a config type
func readConfig(yamlStr []byte) Conf {

	// First unmarshall into generic structure
	var data map[string]interface{}
	err := yaml.Unmarshal(yamlStr, &data)
	if err != nil {
		log.Fatalf("Unmarshal: %v\n", err)
	}

	// A config can hold multiple keyed sections
	c := Conf{}

	// If we have a dockerhierarchy, add it
	if item, ok := data["dockerhierarchy"]; ok {
		c.DockerHierarchy = convertDockerHierarchy(item)
	}

	// If we have a docker build, need to parse items
	c.DockerBuild = DockerBuild{}
	if item, ok := data["dockerbuild"]; ok {
		buildArgs := item.(map[string]interface{})
		if builditem, ok := buildArgs["build_args"]; ok {
			c.DockerBuild.BuildArgs = convertDockerBuildArgs(builditem)
		}

		// Otherwise it could be a matrix
		if matrixitem, ok := buildArgs["matrix"]; ok {
			c.DockerBuild.Matrix = convertDockerBuildMatrix(matrixitem)
		}

		// Or a boolean for active, default to true
		if active, ok := buildArgs["active"]; ok {
			c.DockerBuild.Active = active.(bool)
		} else {
			c.DockerBuild.Active = true	
		}

	}
	return c
}

// convertDockerHierarchy maps the dockerhierarchy portion to a DockerHierarchy
func convertDockerHierarchy(item interface{}) DockerHierarchy {
	hier := DockerHierarchy{}
	mapstructure.Decode(item, &hier)
	return hier
}

// convertDockerBuild maps the dockerbuild build params to a DockerBuild
func convertDockerBuildArgs(item interface{}) map[string]BuildArg {
	build := map[string]BuildArg{}
	mapstructure.Decode(item, &build)
	return build
}

// convertDockerBuildMatrix maps the dockerbuild matrix to a Matrix
func convertDockerBuildMatrix(item interface{}) map[string][]string {
	matrix := map[string][]string{}
	mapstructure.Decode(item, &matrix)
	return matrix
}

type Conf struct {
	DockerHierarchy DockerHierarchy `yaml:"dockerhierarchy,omitempty"`
	DockerBuild     DockerBuild     `yaml:"dockerbuild,omitempty"`
}

func Load(yamlfile string) Conf {
	yamlContent, err := ioutil.ReadFile(yamlfile)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	return readConfig(yamlContent)
}
