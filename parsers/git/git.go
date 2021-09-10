package git

import (
	"encoding/json"
	"fmt"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/vsoch/uptodate/parsers"
	"github.com/vsoch/uptodate/utils"
	"log"
	"strings"
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
func GetChangedFiles(path string, branch string) []GitChange {
	allChanges := GetAllChanges(path, branch)
	changedFiles := []GitChange{}
	for _, change := range allChanges {
		if change.Action == "Modify" || change.Action == "Add" {
			changedFiles = append(changedFiles, change)
		}
	}
	return changedFiles
}

// GetChangedFiles filters to changed
func GetChangedFilesStrings(path string, branch string) []string {
	changes := GetChangedFiles(path, branch)
	names := []string{}
	for _, change := range changes {
		names = append(names, change.Name)
	}
	return names
}

// GetAllChanges for current git repository at some path
func GetAllChanges(path string, main string) []GitChange {

	changedFiles := []GitChange{}
	repo, err := gogit.PlainOpen(path)
	if err != nil {
		log.Fatalf("Cannot find repo at %s: %s\n", path, err)
	}

	// Get the branch pointed by HEAD
	ref, err := repo.Head()
	if err != nil {
		log.Fatalf("Cannot get branch pointed to by HEAD: %s\n", err)
	}

	// Get the main branch from the remote
	remote, err := repo.Remote("origin")
	if err != nil {
		panic(err)
	}
	refList, err := remote.List(&gogit.ListOptions{})
	if err != nil {
		panic(err)
	}
	refPrefix := "refs/heads/"
	var branch *plumbing.Reference
	for _, ref := range refList {
		refName := ref.Name().String()
		if !strings.HasPrefix(refName, refPrefix) {
			continue
		}
		branchName := refName[len(refPrefix):]
		if branchName == main {
			branch = ref
		}
	}

	// If we didn't find the main branch
	if branch == nil {
		log.Fatalf("Could not find main branch %s\n", main)
	}

	// Get the commit object
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		log.Fatalf("Cannot retrieve HEAD commit: %s\n", err)
	}

	prevCommit, err := repo.CommitObject(branch.Hash())
	if err != nil {
		log.Fatalf("Cannot get previous commit: %s\n", err)
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
func (s *GitParser) Parse(path string, branch string) error {

	changedFiles := GetAllChanges(path, branch)

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
