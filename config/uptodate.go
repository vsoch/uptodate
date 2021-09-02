package config

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

type Container struct {
	Name    string   `yaml:"name"`
	Filter  []string `yaml:"filter,omitempty"`
	StartAt string   `yaml:"startat,omitempty"`
	Skips   []string `yaml:"skips,omitempty"`
}

type DockerHierarchy struct {
	Container Container `yaml:"container"`
}

// DockerBuild holds one or more build args
// We'd have to separate these later anyway
type DockerBuild struct {
	BuildArgs map[string]BuildArg `yaml:"build_args"`
}

// BuildArg expects metadata for building from a variable, spack, or other
type BuildArg struct {
	Name     string            `yaml:"name,omitempty"`
	Type     string            `yaml:"type,omitempty"`
	StartAt  string            `yaml:"startat,omitempty"`
	Versions []string          `yaml:"versions,omitempty"`
	Values   []string          `yaml:"values,omitempty"`
	Filter   []string          `yaml:"filter,omitempty"`
	Skips    []string          `yaml:"skips,omitempty"`
	Params   map[string]string `yaml:"params,omitempty"`
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
	if item, ok := data["dockerbuild"]; ok {
		buildArgs := item.(map[string]interface{})
		if builditem, ok := buildArgs["build_args"]; ok {
			c.DockerBuild = convertDockerBuild(builditem)
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
func convertDockerBuild(item interface{}) DockerBuild {
	build := map[string]BuildArg{}
	mapstructure.Decode(item, &build)
	db := DockerBuild{BuildArgs: build}
	return db
}

type Conf struct {
	DockerHierarchy DockerHierarchy `yaml:"dockerhierarchy,omitempty"`
	DockerBuild     DockerBuild     `yaml:"dockerbuild,omitempty"`
}

func Load(yamlfile string) Conf {
	fmt.Println(yamlfile)
	yamlContent, err := ioutil.ReadFile(yamlfile)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	return readConfig(yamlContent)
}
