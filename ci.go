package main

import (
	"os"
	"strings"
)

type ciInfo struct {
	gitBranch string
	gitCommit string
	provider  ciProvider
	skipCount int
}

//-----------------------------------------------------------------------------

type ciProvider int

const (
	unknown ciProvider = iota // MUST be first
	appCenter
	azureDevOps
	bitrise
	circleCI
	codeBuild
	gitHubActions
	jenkins
	teamCity
	travisCI
	xcodeCloud
)

func (cp ciProvider) string() string {
	return [...]string{
		"Unknown",
		"App Center",
		"Azure DevOps",
		"Bitrise",
		"CircleCI",
		"CodeBuild",
		"GitHub Actions",
		"Jenkins",
		"TeamCity",
		"Travis CI",
		"Xcode Cloud"}[cp]
}

//-----------------------------------------------------------------------------

func detectCIInfo(fullInfo bool) *ciInfo {
	info := &ciInfo{
		provider: detectCIProvider()}

	if fullInfo {
		info.extractFullInfo()
	}

	return info
}

//-----------------------------------------------------------------------------

func (ci *ciInfo) extractFullInfo() {
	switch ci.provider {
	case appCenter:
		ci.extractFullInfoFromAppCenter()

	case azureDevOps:
		ci.extractFullInfoFromAzureDevOps()

	case bitrise:
		ci.extractFullInfoFromBitrise()

	case circleCI:
		ci.extractFullInfoFromCircleCI()

	case codeBuild:
		ci.extractFullInfoFromCodeBuild()

	case gitHubActions:
		ci.extractFullInfoFromGitHubActions()

	case jenkins:
		ci.extractFullInfoFromJenkins()

	case teamCity:
		ci.extractFullInfoFromTeamCity()

	case travisCI:
		ci.extractFullInfoFromTravisCI()

	case xcodeCloud:
		ci.extractFullInfoFromXcodeCloud()

	default:
		break
	}
}

func (ci *ciInfo) extractFullInfoFromAppCenter() {
	//
	// https://docs.microsoft.com/en-us/appcenter/build/custom/variables/
	//
	ci.gitBranch = os.Getenv("APPCENTER_BRANCH")
	ci.gitCommit = "" //os.Getenv("???") -- not currently supported?
}

func (ci *ciInfo) extractFullInfoFromAzureDevOps() {
	//
	// https://docs.microsoft.com/en-us/azure/devops/pipelines/build/variables#build-variables-devops-services
	//
	ci.gitBranch = os.Getenv("BUILD_SOURCEBRANCHNAME")
	ci.gitCommit = os.Getenv("BUILD_SOURCEVERSION")
}

func (ci *ciInfo) extractFullInfoFromBitrise() {
	//
	// https://devcenter.bitrise.io/en/references/available-environment-variables.html
	//
	ci.gitBranch = os.Getenv("BITRISE_GIT_BRANCH")
	ci.gitCommit = os.Getenv("BITRISE_GIT_COMMIT")
}

func (ci *ciInfo) extractFullInfoFromCircleCI() {
	//
	// https://circleci.com/docs2/2.0/env-vars#built-in-environment-variables
	//
	ci.gitBranch = os.Getenv("CIRCLE_BRANCH")
	ci.gitCommit = os.Getenv("CIRCLE_SHA1")
}

func (ci *ciInfo) extractFullInfoFromCodeBuild() {
	//
	// https://docs.aws.amazon.com/codebuild/latest/userguide/build-env-ref-env-vars.html
	//
	trigger := os.Getenv("CODEBUILD_WEBHOOK_TRIGGER")

	if strings.HasPrefix(trigger, "branch/") {
		ci.gitBranch = strings.TrimPrefix(trigger, "branch/")
	} else {
		ci.gitBranch = ""
	}

	ci.gitCommit = os.Getenv("CODEBUILD_WEBHOOK_PREV_COMMIT")
}

func (ci *ciInfo) extractFullInfoFromGitHubActions() {
	//
	// https://docs.github.com/en/actions/learn-github-actions/environment-variables#default-environment-variables
	//
	eventName := os.Getenv("GITHUB_EVENT_NAME")
	refType := os.Getenv("GITHUB_REF_TYPE")

	switch eventName {
	case "pull_request", "pull_request_target":
		if refType == "branch" {
			ci.gitBranch = os.Getenv("GITHUB_HEAD_REF")
		} else {
			ci.gitBranch = ""
		}

		//
		// The following environment variable must be set by us (most likely in
		// a custom action) to match the current value of
		// `github.event.pull_request.head.sha`:
		//
		ci.gitCommit = os.Getenv("GITHUB_EVENT_PULL_REQUEST_HEAD_SHA")

		ci.skipCount = 1

	case "push":
		if refType == "branch" {
			ci.gitBranch = os.Getenv("GITHUB_REF_NAME")
		} else {
			ci.gitBranch = ""
		}

		ci.gitCommit = os.Getenv("GITHUB_SHA")

	default:
		ci.gitBranch = ""
		ci.gitCommit = ""
	}
}

func (ci *ciInfo) extractFullInfoFromJenkins() {
	ci.gitBranch = "" //os.Getenv("???") -- not currently supported?
	ci.gitCommit = "" //os.Getenv("???") -- not currently supported?
}

func (ci *ciInfo) extractFullInfoFromTeamCity() {
	ci.gitBranch = "" //os.Getenv("???") -- not currently supported?
	ci.gitCommit = "" //os.Getenv("???") -- not currently supported?
}

func (ci *ciInfo) extractFullInfoFromTravisCI() {
	//
	// https://docs.travis-ci.com/user/environment-variables/#default-environment-variables
	//
	ci.gitBranch = os.Getenv("TRAVIS_BRANCH")
	ci.gitCommit = os.Getenv("TRAVIS_COMMIT")
}

func (ci *ciInfo) extractFullInfoFromXcodeCloud() {
	//
	// https://developer.apple.com/documentation/xcode/environment-variable-reference
	//
	ci.gitBranch = os.Getenv("CI_BRANCH")

	if ci.gitBranch == "" {
		ci.gitBranch = os.Getenv("CI_PULL_REQUEST_SOURCE_BRANCH")
	}

	ci.gitCommit = os.Getenv("CI_COMMIT")

	if ci.gitCommit == "" {
		ci.gitCommit = os.Getenv("CI_PULL_REQUEST_SOURCE_COMMIT")
	}
}

//-----------------------------------------------------------------------------

func detectCIProvider() ciProvider {
	switch {
	case onAppCenter():
		return appCenter

	case onAzureDevOps():
		return azureDevOps

	case onBitrise():
		return bitrise

	case onCircleCI():
		return circleCI

	case onCodeBuild():
		return codeBuild

	case onGitHubActions():
		return gitHubActions

	case onJenkins():
		return jenkins

	case onTeamCity():
		return teamCity

	case onTravisCI():
		return travisCI

	case onXcodeCloud():
		return xcodeCloud

	default:
		return unknown
	}
}

func onAppCenter() bool {
	return len(os.Getenv("APPCENTER_BUILD_ID")) > 0
}

func onAzureDevOps() bool {
	return len(os.Getenv("AGENT_ID")) > 0
}

func onBitrise() bool {
	return os.Getenv("BITRISE_IO") == "true"
}

func onCircleCI() bool {
	return os.Getenv("CIRCLECI") == "true"
}

func onCodeBuild() bool {
	return len(os.Getenv("CODEBUILD_BUILD_ID")) > 0
}

func onGitHubActions() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}

func onJenkins() bool {
	return len(os.Getenv("JENKINS_URL")) > 0
}

func onTeamCity() bool {
	return len(os.Getenv("TEAMCITY_VERSION")) > 0
}

func onTravisCI() bool {
	return os.Getenv("TRAVIS") == "true"
}

func onXcodeCloud() bool {
	return len(os.Getenv("CI_BUILD_ID")) > 0
}
