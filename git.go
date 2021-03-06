package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type gitInfo struct {
	access gitAccess
	branch string
	commit string
}

//-----------------------------------------------------------------------------

type gitAccess int

const (
	ok gitAccess = iota + 1 // MUST be first
	noGitCommandFound
	notGitRepository
)

func (ga gitAccess) String() string {
	return [...]string{
		"ok",
		"noGitCommandFound",
		"notGitRepository"}[ga-1]
}

//-----------------------------------------------------------------------------

func inferGitInfo(skipCount int) *gitInfo {
	access := ok
	branch := ""
	commit := ""

	if !isGitInstalled() {
		access = noGitCommandFound
	} else if !hasGitRepository() {
		access = notGitRepository
	} else {
		commit = inferGitCommit(skipCount)
		branch = inferGitBranch(commit)
	}

	return &gitInfo{
		access: access,
		branch: branch,
		commit: commit}
}

//-----------------------------------------------------------------------------

func fetchBranchNamesFromGitForEachRefResults(results string) []string {
	lines := strings.Split(results, "\n")

	var branchNames []string

	for _, line := range lines {
		branchName := refNameToBranchName(line)

		if len(branchName) > 0 {
			branchNames = append(branchNames, branchName)
		}
	}

	return removeDuplicates(branchNames)
}

func hasGitRepository() bool {
	_, _, err := run("git", "rev-parse")

	return err == nil
}

func inferGitBranch(commit string) string {
	if len(commit) > 0 {
		fromForEachRev := inferGitBranchFromForEachRef(commit)

		if len(fromForEachRev) > 0 {
			return fromForEachRev
		}

		fromNameRev := inferGitBranchFromNameRev(commit)

		if len(fromNameRev) > 0 {
			return fromNameRev
		}
	}

	return inferGitBranchFromRevParse()
}

func inferGitBranchFromForEachRef(commit string) string {
	stdout, _, err := run("git", "for-each-ref", fmt.Sprintf("--points-at=%s", commit), "--format=%(refname)")

	if err == nil {
		branchNames := fetchBranchNamesFromGitForEachRefResults(stdout)

		if len(branchNames) > 0 {
			//
			// Since we don’t know which branch is the correct one, arbitrarily
			// return the first one:
			//
			return branchNames[0]
		}
	}

	return ""
}

func inferGitBranchFromNameRev(commit string) string {
	name, _, err := run("git", "name-rev", "--always", "--name-only", commit)

	if err == nil {
		return nameRevToBranchName(name)
	}

	return ""
}

func inferGitBranchFromRevParse() string {
	name, _, err := run("git", "rev-parse", "--abbrev-ref", "HEAD")

	if err == nil && name != "HEAD" {
		return name
	}

	return ""
}

func inferGitCommit(skipCount int) string {
	skip := fmt.Sprintf("--skip=%d", skipCount)

	hash, _, err := run("git", "log", "--format=%H", skip, "-1")

	if err != nil {
		return ""
	}

	return hash
}

func isGitInstalled() bool {
	var name string

	if runtime.GOOS == "windows" {
		name = "git.exe"
	} else {
		name = "git"
	}

	_, err := exec.LookPath(name)

	return err == nil
}

func nameRevToBranchName(refName string) string {
	branchName := strings.TrimSpace(refName)

	if strings.HasPrefix(branchName, "tags/") {
		return ""
	}

	if strings.HasPrefix(branchName, "remotes/") {
		branchName = strings.TrimPrefix(branchName, "remotes/")

		//
		// Remove the remote name:
		//
		slash := strings.Index(branchName, "/")

		if slash == -1 {
			return ""
		}

		branchName = branchName[slash+1:]
	}

	if branchName == "HEAD" {
		return ""
	}

	return branchName
}

func refNameToBranchName(refName string) string {
	branchName := strings.TrimSpace(refName)

	if strings.HasPrefix(branchName, "refs/heads/") {
		branchName = strings.TrimPrefix(branchName, "refs/heads/")
	} else if strings.HasPrefix(branchName, "refs/remotes/") {
		branchName = strings.TrimPrefix(branchName, "refs/remotes/")

		//
		// Remove the remote name:
		//
		slash := strings.Index(branchName, "/")

		if slash == -1 {
			return ""
		}

		branchName = branchName[slash+1:]
	} else {
		return ""
	}

	if branchName == "HEAD" {
		return ""
	}

	return branchName
}

func removeDuplicates(strings []string) []string {
	seen := make(map[string]bool)

	var result []string

	for _, s := range strings {
		if _, ok := seen[s]; !ok {
			seen[s] = true
			result = append(result, s)
		}
	}

	return result
}
