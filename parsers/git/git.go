package git

import (
	"encoding/json"
	"fmt"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/vsoch/uptodate/parsers"
	"github.com/vsoch/uptodate/utils"
	"log"
)

type GitParser struct {
	Root string
}

// Keep a record of changed file name and change type
type GitChange struct {
	Name   string
	Action string
}

func getChangeName(change *object.Change) string {
	var empty = object.ChangeEntry{}
	if change.From != empty {
		return change.From.Name
	}
	return change.To.Name
}

// GetChangedFiles filters to changed
func GetChangedFiles(path string) []GitChange {
	allChanges := GetAllChanges(path)
	changedFiles := []GitChange{}
	for _, change := range allChanges {
		if change.Action == "Modify" || change.Action == "Add" {
			changedFiles = append(changedFiles, change)
		}
	}
	return changedFiles
}

// GetChangedFiles filters to changed
func GetChangedFilesStrings(path string) []string {
	changes := GetChangedFiles(path)
	names := []string{}
	for _, change := range changes {
		names = append(names, change.Name)
	}
	return names
}

// GetAllChanges for current git repository at some path
func GetAllChanges(path string) []GitChange {

	changedFiles := []GitChange{}
	repo, err := gogit.PlainOpen(path)
	if err != nil {
		log.Fatalf("Cannot find repo at %s\n", path)
	}

	// Get the branch pointed by HEAD
	ref, err := repo.Head()
	if err != nil {
		log.Fatalf("Cannot get branch pointed to by HEAD\n")
	}

	// Get the commit object
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		log.Fatalf("Cannot retrieve HEAD commit.\n")
	}

	prevCommit, err := commit.Parent(0)
	if err != nil {
		log.Fatalf("Cannot get previous commit\n")
	}

	// Get trees for current and previous commit
	currentTree, err := commit.Tree()
	prevTree, err := prevCommit.Tree()

	// Get new, modified, deleted files
	changes, err := currentTree.Diff(prevTree)
	for _, change := range changes {
		action, err := change.Action()
		if err != nil {
			fmt.Printf("Cannot get action for %s, skipping\n", change)
			continue
		}

		// Get list of involved files
		name := getChangeName(change)
		change := GitChange{Name: name, Action: action.String()}
		changedFiles = append(changedFiles, change)
	}
	return changedFiles
}

// Entrypoint to parse one or more Docker build matrices
func (s *GitParser) Parse(path string) error {

	changedFiles := GetAllChanges(path)

	// Format into a list of results
	results := []parsers.Result{}

	fmt.Println("\n  ⭐️ Changed Files ⭐️")
	for _, change := range changedFiles {
		newResult := parsers.Result{Name: change.Action, Filename: change.Name}
		fmt.Printf("    %s: %s\n", change.Name, change.Action)
		results = append(results, newResult)
	}

	// Parse into json
	outJson, _ := json.Marshal(results)

	// If we are running in a GitHub Action, set the outputs
	if utils.IsGitHubAction() {
		fmt.Printf("::set-output name=git_matrix::%s\n", string(outJson))
	}
	return nil
}
