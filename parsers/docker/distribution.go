package docker

// Types and functions for manifests, configs, etc.

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/vsoch/uptodate/utils"
)

type ImageConfig struct {
	Architecture string `json:"architecture"`
	Config       struct {
		Hostname     string            `json:"Hostname"`
		Domainname   string            `json:"Domainname"`
		User         string            `json:"User"`
		AttachStdin  bool              `json:"AttachStdin"`
		AttachStdout bool              `json:"AttachStdout"`
		AttachStderr bool              `json:"AttachStderr"`
		Tty          bool              `json:"Tty"`
		OpenStdin    bool              `json:"OpenStdin"`
		StdinOnce    bool              `json:"StdinOnce"`
		Env          []string          `json:"Env"`
		Cmd          []string          `json:"Cmd"`
		Image        string            `json:"Image"`
		Volumes      interface{}       `json:"Volumes"`
		WorkingDir   string            `json:"WorkingDir"`
		Entrypoint   interface{}       `json:"Entrypoint"`
		OnBuild      interface{}       `json:"OnBuild"`
		Labels       map[string]string `json:"Labels"`
	} `json:"config"`
	Container       string `json:"container"`
	ContainerConfig struct {
		Hostname     string            `json:"Hostname"`
		Domainname   string            `json:"Domainname"`
		User         string            `json:"User"`
		AttachStdin  bool              `json:"AttachStdin"`
		AttachStdout bool              `json:"AttachStdout"`
		AttachStderr bool              `json:"AttachStderr"`
		Tty          bool              `json:"Tty"`
		OpenStdin    bool              `json:"OpenStdin"`
		StdinOnce    bool              `json:"StdinOnce"`
		Env          []string          `json:"Env"`
		Cmd          []string          `json:"Cmd"`
		Image        string            `json:"Image"`
		Volumes      interface{}       `json:"Volumes"`
		WorkingDir   string            `json:"WorkingDir"`
		Entrypoint   interface{}       `json:"Entrypoint"`
		OnBuild      interface{}       `json:"OnBuild"`
		Labels       map[string]string `json:"Labels"`
	} `json:"container_config"`
	Created       time.Time `json:"created"`
	DockerVersion string    `json:"docker_version"`
	History       []struct {
		Created    time.Time `json:"created"`
		CreatedBy  string    `json:"created_by"`
		EmptyLayer bool      `json:"empty_layer,omitempty"`
	} `json:"history"`
	Os     string `json:"os"`
	Rootfs struct {
		Type    string   `json:"type"`
		DiffIds []string `json:"diff_ids"`
	} `json:"rootfs"`
}

// GetImageConfig of an existing container
func GetImageConfig(container string) ImageConfig {

	// Get tags for current container image
	configUrl := "https://crane.ggcr.dev/config/" + container
	response := utils.GetRequest(configUrl, map[string]string{})

	imageConf := ImageConfig{}
	json.Unmarshal([]byte(response), &imageConf)

	// We don't care about the error - a missing config means we will build anyway
	return imageConf
}

// Get image tags for a container
func GetImageTags(container string) []string {
	tagsUrl := "https://crane.ggcr.dev/ls/" + container
	response := utils.GetRequest(tagsUrl, map[string]string{})
	tags := strings.Split(response, "\n")
	return tags
}
