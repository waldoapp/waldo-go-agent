package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type uploadAction struct {
	userBuildPath   string
	userGitBranch   string
	userGitCommit   string
	userOverrides   map[string]string
	userUploadToken string
	userVariantName string
	userVerbose     bool

	absBuildPath        string
	absBuildPayloadPath string
	absWorkingPath      string
	buildSuffix         string
	ciInfo              *ciInfo
	flavor              string
	gitInfo             *gitInfo
	rtInfo              *rtInfo
	validated           bool
}

//-----------------------------------------------------------------------------

func newUploadAction(buildPath, uploadToken, variantName, gitCommit, gitBranch string, verbose bool, overrides map[string]string) *uploadAction {
	return &uploadAction{
		rtInfo:          detectRTInfo(),
		userBuildPath:   buildPath,
		userGitBranch:   gitBranch,
		userGitCommit:   gitCommit,
		userOverrides:   overrides,
		userUploadToken: uploadToken,
		userVariantName: variantName,
		userVerbose:     verbose}
}

//-----------------------------------------------------------------------------

func (ua *uploadAction) buildPath() string {
	if ua.validated {
		return ua.absBuildPath
	}

	return ua.userBuildPath
}

func (ua *uploadAction) buildPayloadPath() string {
	return ua.absBuildPayloadPath
}

func (ua *uploadAction) ciGitBranch() string {
	return ua.ciInfo.gitBranch
}

func (ua *uploadAction) ciGitCommit() string {
	return ua.ciInfo.gitCommit
}

func (ua *uploadAction) ciProvider() string {
	return ua.ciInfo.provider.string()
}

func (ua *uploadAction) gitAccess() string {
	return ua.gitInfo.access.String()
}

func (ua *uploadAction) gitBranch() string {
	return ua.userGitBranch
}

func (ua *uploadAction) gitCommit() string {
	return ua.userGitCommit
}

func (ua *uploadAction) inferredGitBranch() string {
	return ua.gitInfo.branch
}

func (ua *uploadAction) inferredGitCommit() string {
	return ua.gitInfo.commit
}

func (ua *uploadAction) uploadToken() string {
	return ua.userUploadToken
}

func (ua *uploadAction) variantName() string {
	return ua.userVariantName
}

func (ua *uploadAction) version() string {
	return ua.rtInfo.version()
}

//-----------------------------------------------------------------------------

func (ua *uploadAction) perform() error {
	err := os.RemoveAll(ua.absWorkingPath)

	if err == nil {
		err = os.MkdirAll(ua.absWorkingPath, 0755)
	}

	defer os.RemoveAll(ua.absWorkingPath)

	if err == nil {
		err = ua.createBuildPayload()
	}

	if err == nil {
		err = ua.uploadBuild()
	}

	if err != nil {
		ua.uploadError(err)
	}

	return err
}

func (ua *uploadAction) validate() error {
	if ua.validated {
		return nil
	}

	err := validateUploadToken(ua.userUploadToken)

	if err != nil {
		return err
	}

	buildPath, buildSuffix, flavor, err := validateBuildPath(ua.userBuildPath)

	if err != nil {
		return err
	}

	workingPath := determineWorkingPath()

	ua.absBuildPath = buildPath
	ua.absBuildPayloadPath = determineBuildPayloadPath(workingPath, buildPath, buildSuffix)
	ua.absWorkingPath = workingPath
	ua.buildSuffix = buildSuffix
	ua.ciInfo = detectCIInfo(true)
	ua.flavor = flavor
	ua.gitInfo = inferGitInfo(ua.ciInfo.skipCount)
	ua.validated = true

	return nil
}

//-----------------------------------------------------------------------------

func (ua *uploadAction) authorization() string {
	return fmt.Sprintf("Upload-Token %s", ua.userUploadToken)
}

func (ua *uploadAction) buildContentType() string {
	switch ua.buildSuffix {
	case "app":
		return zipContentType

	default:
		return binaryContentType
	}
}

func (ua *uploadAction) checkBuildStatus(resp *http.Response) error {
	status := resp.StatusCode

	if status == 401 {
		return fmt.Errorf("Upload token is invalid or missing!")
	}

	if status < 200 || status > 299 {
		return fmt.Errorf("Unable to upload build to Waldo, HTTP status: %d", status)
	}

	return nil
}

func (ua *uploadAction) createBuildPayload() error {
	parentPath := filepath.Dir(ua.absBuildPath)
	buildName := filepath.Base(ua.absBuildPath)

	switch ua.buildSuffix {
	case "app":
		if !isDir(ua.absBuildPath) {
			return fmt.Errorf("Unable to read build at ‘%s’", ua.absBuildPath)
		}

		return zipFolder(ua.absBuildPayloadPath, parentPath, buildName)

	default:
		if !isRegular(ua.absBuildPath) {
			return fmt.Errorf("Unable to read build at ‘%s’", ua.absBuildPath)
		}

		return nil
	}
}

func (ua *uploadAction) errorContentType() string {
	return jsonContentType
}

func (ua *uploadAction) makeBuildURL() string {
	buildURL := ua.userOverrides["apiBuildEndpoint"]

	if len(buildURL) == 0 {
		buildURL = defaultAPIBuildEndpoint
	}

	query := make(url.Values)

	addIfNotEmpty(&query, "agentName", agentName)
	addIfNotEmpty(&query, "agentVersion", agentVersion)
	addIfNotEmpty(&query, "arch", ua.rtInfo.arch)
	addIfNotEmpty(&query, "ci", ua.ciInfo.provider.string())
	addIfNotEmpty(&query, "ciGitBranch", ua.ciInfo.gitBranch)
	addIfNotEmpty(&query, "ciGitCommit", ua.ciInfo.gitCommit)
	addIfNotEmpty(&query, "flavor", ua.flavor)
	addIfNotEmpty(&query, "gitAccess", ua.gitInfo.access.String())
	addIfNotEmpty(&query, "gitBranch", ua.gitInfo.branch)
	addIfNotEmpty(&query, "gitCommit", ua.gitInfo.commit)
	addIfNotEmpty(&query, "platform", ua.rtInfo.platform)
	addIfNotEmpty(&query, "userGitBranch", ua.userGitBranch)
	addIfNotEmpty(&query, "userGitCommit", ua.userGitCommit)
	addIfNotEmpty(&query, "variantName", ua.userVariantName)
	addIfNotEmpty(&query, "wrapperName", ua.userOverrides["wrapperName"])
	addIfNotEmpty(&query, "wrapperVersion", ua.userOverrides["wrapperVersion"])

	buildURL += "?" + query.Encode()

	return buildURL
}

func (ua *uploadAction) makeErrorPayload(err error) string {
	payload := ""

	appendIfNotEmpty(&payload, "agentName", agentName)
	appendIfNotEmpty(&payload, "agentVersion", agentVersion)
	appendIfNotEmpty(&payload, "arch", ua.rtInfo.arch)
	appendIfNotEmpty(&payload, "ci", ua.ciInfo.provider.string())
	appendIfNotEmpty(&payload, "ciGitBranch", ua.ciInfo.gitBranch)
	appendIfNotEmpty(&payload, "ciGitCommit", ua.ciInfo.gitCommit)
	appendIfNotEmpty(&payload, "message", err.Error())
	appendIfNotEmpty(&payload, "platform", ua.rtInfo.platform)
	appendIfNotEmpty(&payload, "wrapperName", ua.userOverrides["wrapperName"])
	appendIfNotEmpty(&payload, "wrapperVersion", ua.userOverrides["wrapperVersion"])

	payload = "{" + payload + "}"

	return payload
}

func (ua *uploadAction) makeErrorURL() string {
	errorURL := ua.userOverrides["apiErrorEndpoint"]

	if len(errorURL) == 0 {
		errorURL = defaultAPIErrorEndpoint
	}

	return errorURL
}

func (ua *uploadAction) uploadBuild() error {
	url := ua.makeBuildURL()

	file, err := os.Open(ua.absBuildPayloadPath)

	if err != nil {
		return fmt.Errorf("Unable to upload build to Waldo, error: %v, url: %s", err, url)
	}

	defer file.Close()

	client := &http.Client{}

	req, err := http.NewRequest("POST", url, file)

	if err != nil {
		return fmt.Errorf("Unable to upload build to Waldo, error: %v, url: %s", err, url)
	}

	req.Header.Add("Authorization", ua.authorization())
	req.Header.Add("Content-Type", ua.buildContentType())
	req.Header.Add("User-Agent", ua.userAgent())

	dumpRequest(ua.userVerbose, req, false)

	resp, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("Unable to upload build to Waldo, error: %v, url: %s", err, url)
	}

	dumpResponse(ua.userVerbose, resp, true)

	defer resp.Body.Close()

	return ua.checkBuildStatus(resp)
}

func (ua *uploadAction) uploadError(err error) error {
	url := ua.makeErrorURL()
	body := ua.makeErrorPayload(err)

	client := &http.Client{}

	req, err := http.NewRequest("POST", url, strings.NewReader(body))

	if err != nil {
		return err
	}

	req.Header.Add("Authorization", ua.authorization())
	req.Header.Add("Content-Type", ua.errorContentType())
	req.Header.Add("User-Agent", ua.userAgent())

	// dumpRequest(ua.userVerbose, req, true)

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// dumpResponse(ua.userVerbose, resp, true)

	return nil
}

func (ua *uploadAction) userAgent() string {
	ci := ua.ciInfo.provider.string()

	if ci == "Unknown" {
		ci = "Go Agent"
	}

	version := ua.userOverrides["wrapperVersion"]

	if len(version) == 0 {
		version = agentVersion
	}

	return fmt.Sprintf("Waldo %s/%s v%s", ci, ua.flavor, version)
}
