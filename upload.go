package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//-----------------------------------------------------------------------------

type uploadAction struct {
	retryCount      int
	userAppID       string
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
	failureBody         any
	failureHeaders      any
	failureStatusCode   int
	flavor              string
	gitInfo             *gitInfo
	rtInfo              *rtInfo
	uploadID            string
	uploadMetadata      *UploadMetadata
	validated           bool
}

//-----------------------------------------------------------------------------

func newUploadAction(buildPath, uploadToken, appID, variantName, gitCommit, gitBranch string, verbose bool, overrides map[string]string) *uploadAction {
	return &uploadAction{
		retryCount:      0,
		rtInfo:          detectRTInfo(),
		userAppID:       appID,
		userBuildPath:   buildPath,
		userGitBranch:   gitBranch,
		userGitCommit:   gitCommit,
		userOverrides:   overrides,
		userUploadToken: uploadToken,
		userVariantName: variantName,
		userVerbose:     verbose}
}

//-----------------------------------------------------------------------------

type ErrorPayloadJSON struct {
	AgentName         string `json:"agentName,omitempty"`
	AgentVersion      string `json:"agentVersion,omitempty"`
	Arch              string `json:"arch,omitempty"`
	CI                string `json:"ci,omitempty"`
	CIGitBranch       string `json:"ciGitBranch,omitempty"`
	CIGitCommit       string `json:"ciGitCommit,omitempty"`
	FailureBody       any    `json:"failureBody,omitempty"`
	FailureHeaders    any    `json:"failureHeaders,omitempty"`
	FailureStatusCode int    `json:"failureStatusCode"`
	Message           string `json:"message,omitempty"`
	Platform          string `json:"platform,omitempty"`
	Retry             int    `json:"retry"`
	WrapperName       string `json:"wrapperName,omitempty"`
	WrapperVersion    string `json:"wrapperVersion,omitempty"`
}

//-----------------------------------------------------------------------------

func (ua *uploadAction) appID() string {
	return ua.userAppID
}

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

func (ua *uploadAction) retry() string {
	return strconv.Itoa(ua.retryCount)
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
		err = ua.uploadBuildWithRetry()
	}

	if err != nil {
		ua.uploadErrorWithRetry(err)
	}

	return err
}

func (ua *uploadAction) validate() error {
	if ua.validated {
		return nil
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
	ua.uploadID = randomUploadID()
	ua.validated = true

	return nil
}

//-----------------------------------------------------------------------------

func (ua *uploadAction) authorization() string {
	return fmt.Sprintf("Upload-Token %s", ua.userUploadToken)
}

func (ua *uploadAction) buildContentType() string {
	return binaryContentType
}

func (ua *uploadAction) checkBuildStatus(resp *http.Response) error {
	status := resp.StatusCode

	if status == 401 {
		return errors.New("Upload token is invalid or missing!")
	}

	if status == 403 && ua.isWAFResponse(resp) {
		return errors.New("Upload build blocked by WAF server!")
	}

	if status < 200 || status > 299 {
		return fmt.Errorf("Unable to upload build to Waldo, HTTP status: %d", status)
	}

	return nil
}

func (ua *uploadAction) checkErrorStatus(resp *http.Response) error {
	status := resp.StatusCode

	if status == 403 && ua.isWAFResponse(resp) {
		return errors.New("Upload error blocked by WAF server!")
	}

	if status < 200 || status > 299 {
		return fmt.Errorf("Unable to upload error to Waldo, HTTP status: %d", status)
	}

	return nil
}

func (ua *uploadAction) createBuildPayload() error {
	parentPath := filepath.Dir(ua.absBuildPath)
	buildName := filepath.Base(ua.absBuildPath)

	switch ua.buildSuffix {
	case "apk":
		if !isRegular(ua.absBuildPath) {
			return fmt.Errorf("Unable to read build at %q", ua.absBuildPath)
		}

		return nil

	case "app":
		if !isDir(ua.absBuildPath) {
			return fmt.Errorf("Unable to read build at %q", ua.absBuildPath)
		}

		return zipFolder(ua.absBuildPayloadPath, parentPath, buildName)

	default:
		return fmt.Errorf("Unable to read build at %q", ua.absBuildPath)
	}
}

func (ua *uploadAction) errorContentType() string {
	return jsonContentType
}

func (ua *uploadAction) extractUploadMetadata(resp *http.Response, host string) (*UploadMetadata, error) {
	ur, err := parseUploadResponse(resp)

	if err != nil {
		return nil, err
	}

	return &UploadMetadata{
		AppID:        ur.AppID,
		AppVersionID: ur.AppVersionID,
		Host:         host,
		UploadTime:   time.Now()}, nil
}

func (ua *uploadAction) fetchBody(resp *http.Response) any {
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil
	}

	var jsonBody any

	err = json.Unmarshal(body, &jsonBody)

	if err != nil {
		return nil
	}

	return jsonBody
}

func (ua *uploadAction) isWAFResponse(resp *http.Response) bool {
	server := resp.Header.Get("Server")

	if len(server) == 0 {
		return false
	}

	return strings.HasPrefix(server, "awselb/")
}

func (ua *uploadAction) makeBuildURL() string {
	buildURL := ua.userOverrides["apiBuildEndpoint"]

	if len(buildURL) == 0 {
		if strings.HasPrefix(ua.userUploadToken, "u-") {
			buildURL = strings.ReplaceAll(defaultAPIBuildNewEndpoint, "${APP_ID}", ua.userAppID)
		} else {
			buildURL = defaultAPIBuildOldEndpoint
		}
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
	addIfNotEmpty(&query, "retry", ua.retry())
	addIfNotEmpty(&query, "userGitBranch", ua.userGitBranch)
	addIfNotEmpty(&query, "userGitCommit", ua.userGitCommit)
	addIfNotEmpty(&query, "variantName", ua.userVariantName)
	addIfNotEmpty(&query, "wrapperName", ua.userOverrides["wrapperName"])
	addIfNotEmpty(&query, "wrapperVersion", ua.userOverrides["wrapperVersion"])

	buildURL += "?" + query.Encode()

	return buildURL
}

func (ua *uploadAction) makeErrorPayload(err error) (string, error) {
	payload := ErrorPayloadJSON{
		AgentName:         agentName,
		AgentVersion:      agentVersion,
		Arch:              ua.rtInfo.arch,
		CI:                ua.ciInfo.provider.string(),
		CIGitBranch:       ua.ciInfo.gitBranch,
		CIGitCommit:       ua.ciInfo.gitCommit,
		FailureBody:       ua.failureBody,
		FailureHeaders:    ua.failureHeaders,
		FailureStatusCode: ua.failureStatusCode,
		Message:           err.Error(),
		Platform:          ua.rtInfo.platform,
		Retry:             ua.retryCount,
		WrapperName:       ua.userOverrides["wrapperName"],
		WrapperVersion:    ua.userOverrides["wrapperVersion"],
	}

	data, err := json.Marshal(payload)

	if err != nil {
		return "", fmt.Errorf("Unable to encode JSON error payload, error: %v", err)
	}

	return string(data), nil
}

func (ua *uploadAction) makeErrorURL() string {
	errorURL := ua.userOverrides["apiErrorEndpoint"]

	if len(errorURL) == 0 {
		errorURL = defaultAPIErrorEndpoint
	}

	return errorURL
}

func (ua *uploadAction) uploadBuild(retryAllowed bool) (bool, error) {
	fmt.Printf("Uploading build to Waldo…\n")

	url := ua.makeBuildURL()

	file, err := os.Open(ua.absBuildPayloadPath)

	if err != nil {
		return false, ua.wrapUploadError("build", err, url)
	}

	defer file.Close()

	fi, err := file.Stat()

	if err != nil {
		return false, ua.wrapUploadError("build", err, url)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout:   0,
			ResponseHeaderTimeout: 0,
			ExpectContinueTimeout: 0}}

	req, err := http.NewRequest("POST", url, file)

	if err != nil {
		return false, ua.wrapUploadError("build", err, url)
	}

	req.ContentLength = fi.Size()

	req.Header.Add("Authorization", ua.authorization())
	req.Header.Add("Content-Type", ua.buildContentType())
	req.Header.Add("User-Agent", ua.userAgent())
	req.Header.Add("X-Upload-Id", ua.uploadID)

	dumpRequest(ua.userVerbose, req, false)

	resp, err := client.Do(req)

	if err != nil {
		return retryAllowed, ua.wrapUploadError("build", err, url)
	}

	dumpResponse(ua.userVerbose, resp, true)

	defer resp.Body.Close()

	err = ua.checkBuildStatus(resp)

	if err == nil {
		um, err2 := ua.extractUploadMetadata(resp, req.URL.Host)

		if err2 == nil {
			err2 = um.save()
		}

		if err2 == nil {
			ua.uploadMetadata = um
		} else {
			emitError(fmt.Errorf("Unable to save upload metadata locally, error: %v", err2))
		}
	} else {
		ua.failureBody = ua.fetchBody(resp)
		ua.failureHeaders = resp.Header
		ua.failureStatusCode = resp.StatusCode
	}

	return retryAllowed && shouldRetry(resp), err
}

func (ua *uploadAction) uploadBuildWithRetry() error {
	for attempts := 1; attempts <= maxNetworkAttempts; attempts++ {
		ua.retryCount = attempts - 1
		retry, err := ua.uploadBuild(attempts < maxNetworkAttempts)

		if !retry || err == nil {
			return err
		}

		emitError(err)

		fmt.Printf("\nFailed upload attempts: %d -- retrying…\n\n", attempts)
	}

	return nil
}

func (ua *uploadAction) uploadError(err error, retryAllowed bool) (bool, error) {
	url := ua.makeErrorURL()

	body, err := ua.makeErrorPayload(err)

	if err != nil {
		return false, err
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", url, strings.NewReader(body))

	if err != nil {
		return false, ua.wrapUploadError("error", err, url)
	}

	req.ContentLength = int64(len([]byte(body)))

	req.Header.Add("Authorization", ua.authorization())
	req.Header.Add("Content-Type", ua.errorContentType())
	req.Header.Add("User-Agent", ua.userAgent())
	req.Header.Add("X-Upload-Id", ua.uploadID)

	dumpRequest(ua.userVerbose, req, true)

	resp, err := client.Do(req)

	if err != nil {
		return retryAllowed, ua.wrapUploadError("error", err, url)
	}

	dumpResponse(ua.userVerbose, resp, true)

	defer resp.Body.Close()

	return retryAllowed && shouldRetry(resp), ua.checkErrorStatus(resp)
}

func (ua *uploadAction) uploadErrorWithRetry(err error) error {
	for attempts := 1; attempts <= maxNetworkAttempts; attempts++ {
		retry, tmpErr := ua.uploadError(err, attempts < maxNetworkAttempts)

		if !retry || tmpErr == nil {
			return tmpErr
		}

		// emitError(tmpErr)

		// fmt.Printf("\nFailed upload error attempts: %d -- retrying…\n\n", attempts)
	}

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

func (ua *uploadAction) wrapUploadError(desc string, err error, url string) error {
	return fmt.Errorf("Unable to upload %s to Waldo, error: %v, url: %q", desc, err, url)
}
