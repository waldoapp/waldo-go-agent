package main

import (
	"testing"
)

func TestFetchBranchesCanHaveMultiple(t *testing.T) {
	names := fetchBranchNamesFromGitForEachRefResults("refs/remotes/origin/master\nrefs/heads/foo\nrefs/remotes/origin/bar")

	if len(names) != 3 {
		t.Errorf("Expected 1 element, got %v", names)
	}

	if names[0] != "master" {
		t.Errorf("Expected [\"master\"], got %v", names)
	}

	if names[1] != "foo" {
		t.Errorf("Expected [\"master\"], got %v", names)
	}

	if names[2] != "bar" {
		t.Errorf("Expected [\"master\"], got %v", names)
	}
}

func TestFetchBranchesDeduplicate(t *testing.T) {
	names := fetchBranchNamesFromGitForEachRefResults("refs/remotes/origin/master\nrefs/heads/master")

	if len(names) != 1 {
		t.Errorf("Expected 1 element, got %v", names)
	}

	if names[0] != "master" {
		t.Errorf("Expected [\"master\"], got %v", names)
	}
}

func TestFetchBranchesEmptyString(t *testing.T) {
	names := fetchBranchNamesFromGitForEachRefResults("")

	if names != nil {
		t.Errorf("Expected nil, got %v", names)
	}
}

func TestFetchBranchesFull(t *testing.T) {
	names := fetchBranchNamesFromGitForEachRefResults("refs/heads/master\nrefs/remotes/origin/HEAD\nrefs/remotes/origin/master\nrefs/tags/foo")

	if len(names) != 1 {
		t.Errorf("Expected 1 element, got %v", names)
	}

	if names[0] != "master" {
		t.Errorf("Expected [\"master\"], got %v", names)
	}
}

func TestFetchBranchesIgnoreHeads(t *testing.T) {
	names := fetchBranchNamesFromGitForEachRefResults("refs/remotes/origin/HEAD")

	if names != nil {
		t.Errorf("Expected nil, got %v", names)
	}
}

func TestFetchBranchesIgnoreTags(t *testing.T) {
	names := fetchBranchNamesFromGitForEachRefResults("refs/tags/foo")

	if names != nil {
		t.Errorf("Expected nil, got %v", names)
	}
}

func TestFetchBranchesLocalMaster(t *testing.T) {
	names := fetchBranchNamesFromGitForEachRefResults("refs/heads/master")

	if len(names) != 1 {
		t.Errorf("Expected 1 element, got %v", names)
	}

	if names[0] != "master" {
		t.Errorf("Expected [\"master\"], got %v", names)
	}
}

func TestFetchBranchesOriginMaster(t *testing.T) {
	names := fetchBranchNamesFromGitForEachRefResults("refs/remotes/origin/master")

	if len(names) != 1 {
		t.Errorf("Expected 1 element, got %v", names)
	}

	if names[0] != "master" {
		t.Errorf("Expected [\"master\"], got %v", names)
	}
}

func TestNameRevToBranchNameComplex(t *testing.T) {
	name := nameRevToBranchName("features/waldo/git-handling")
	expected := "features/waldo/git-handling"

	if name != expected {
		t.Errorf("Expected %s string, got %v", expected, name)
	}
}

func TestNameRevToBranchNameEmpty(t *testing.T) {
	name := nameRevToBranchName("")
	expected := ""

	if name != expected {
		t.Errorf("Expected %s string, got %v", expected, name)
	}
}

func TestNameRevToBranchNameHead(t *testing.T) {
	name := nameRevToBranchName("HEAD")
	expected := ""

	if name != expected {
		t.Errorf("Expected %s string, got %v", expected, name)
	}
}

func TestNameRevToBranchNameRemote(t *testing.T) {
	name := nameRevToBranchName("remotes/origin/foo")
	expected := "foo"

	if name != expected {
		t.Errorf("Expected %s string, got %v", expected, name)
	}
}

func TestNameRevToBranchNameSimple(t *testing.T) {
	name := nameRevToBranchName("foo")
	expected := "foo"

	if name != expected {
		t.Errorf("Expected %s string, got %v", expected, name)
	}
}

func TestNameRevToBranchNameTags(t *testing.T) {
	name := nameRevToBranchName("tags/bar")
	expected := ""

	if name != expected {
		t.Errorf("Expected %s string, got %v", expected, name)
	}
}
