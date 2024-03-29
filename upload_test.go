package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
)

func TestErrorPayloadEncoding(t *testing.T) {
	ua := newUploadAction("/some/path", "token", "appid", "variant", "f5ebaa009c043737f30bcb0f53d7614d09968e00",
		"user\"my-branch", false, make(map[string]string))

	ua.ciInfo = &ciInfo{
		gitBranch: "user\"my-branch",
	}

	output, err := ua.makeErrorPayload(fmt.Errorf("The following failed: \"%v\"", "some error message"))
	if err != nil {
		t.Error(err)
	}

	var jsonDecoded map[string]any
	err = json.Unmarshal([]byte(output), &jsonDecoded)
	if err != nil {
		t.Error(err)
	}
	if jsonDecoded["ciGitBranch"] != "user\"my-branch" {
		t.Errorf("Expected branch to be 'user\"my-branch', but got '%v'", jsonDecoded["ciGitBranch"])
	}
	if jsonDecoded["message"] != "The following failed: \"some error message\"" {
		t.Errorf("Expected error to be 'The following failed: \"some error message\"', but got '%v'", jsonDecoded["message"])
	}
}

func TestBuildURLEncoding(t *testing.T) {
	ua := newUploadAction("/some/path", "token", "appid", "variant", "f5ebaa009c043737f30bcb0f53d7614d09968e00",
		"user\"=+my-branch", false, make(map[string]string))

	ua.gitInfo = &gitInfo{
		branch: "user\"=+my-branch",
		access: ok,
	}

	ua.ciInfo = &ciInfo{
		gitBranch: "user\"=+my-branch",
	}

	output := ua.makeBuildURL()

	parsed, err := url.ParseRequestURI(output)
	if err != nil {
		t.Error(err)
	}

	if parsed.Query().Get("ciGitBranch") != "user\"=+my-branch" {
		t.Errorf("Expected branch to be 'user\"=+my-branch', but got '%v'", parsed.Query().Get("ciGitBranch"))
	}
}
