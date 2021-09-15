package github

import (
	"encoding/json"
	"github.com/vsoch/uptodate/utils"
	"log"
)

func GetReleases(name string) Releases {

	url := "https://api.github.com/repos/" + name + "/releases"

	headers := make(map[string]string)
	headers["Accept"] = "application/vnd.github.v3+json"
	response := utils.GetRequest(url, headers)

	// The response gets parsed into a spack package
	releases := Releases{}
	err := json.Unmarshal([]byte(response), &releases)
	if err != nil {
		log.Fatalf("Issue unmarshalling releases data structure\n")
	}
	return releases
}

func GetCommits(name string, branch string) Commits {
	url := "https://api.github.com/repos/" + name + "/commits"

	headers := make(map[string]string)
	headers["Accept"] = "application/vnd.github.v3+json"
	headers["Sha"] = branch
	response := utils.GetRequest(url, headers)

	commits := Commits{}
	err := json.Unmarshal([]byte(response), &commits)
	if err != nil {
		log.Fatalf("Issue unmarshalling commits data structure\n")
	}
	return commits
}
