package config

import (
	"gopkg.in/yaml.v2"
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

type Conf struct {
	DockerHierarchy DockerHierarchy `yaml:"dockerhierarchy,omitempty"`
}

func Load(yamlfile string) Conf {
	c := Conf{}
	yamlContent, err := ioutil.ReadFile(yamlfile)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlContent, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}
