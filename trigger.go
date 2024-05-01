package main

import (
	"fmt"
	"net/http"
	"strings"
)

type triggerAction struct {
	userGitCommit   string
	userOverrides   map[string]string
	userRuleName    string
	userUploadToken string
	userVerbose     bool

	ciInfo    *ciInfo
	rtInfo    *rtInfo
	validated bool
}

//-----------------------------------------------------------------------------

func newTriggerAction(uploadToken, ruleName, gitCommit string, verbose bool, overrides map[string]string) *triggerAction {
	return &triggerAction{
		rtInfo:          detectRTInfo(),
		userGitCommit:   gitCommit,
		userOverrides:   overrides,
		userRuleName:    ruleName,
		userUploadToken: uploadToken,
		userVerbose:     verbose}
}

//-----------------------------------------------------------------------------

func (ta *triggerAction) gitCommit() string {
	return ta.userGitCommit
}

func (ta *triggerAction) ruleName() string {
	return ta.userRuleName
}

func (ta *triggerAction) uploadToken() string {
	return ta.userUploadToken
}

func (ta *triggerAction) version() string {
	return ta.rtInfo.version()
}

//-----------------------------------------------------------------------------

func (ta *triggerAction) perform() error {
	return ta.triggerRunWithRetry()
}

func (ta *triggerAction) validate() error {
	if ta.validated {
		return nil
	}

	ta.ciInfo = detectCIInfo(false)
	ta.validated = true

	return nil
}

//-----------------------------------------------------------------------------

func (ta *triggerAction) authorization() string {
	return fmt.Sprintf("Upload-Token %s", ta.userUploadToken)
}

func (ta *triggerAction) checkTriggerStatus(resp *http.Response) error {
	status := resp.StatusCode

	if status == 401 {
		return fmt.Errorf("Upload token is invalid or missing!")
	}

	if status < 200 || status > 299 {
		return fmt.Errorf("Unable to trigger run on Waldo, HTTP status: %d", status)
	}

	return nil
}

func (ta *triggerAction) contentType() string {
	return jsonContentType
}

func (ta *triggerAction) makePayload() string {
	payload := ""

	appendIfNotEmpty(&payload, "agentName", agentName)
	appendIfNotEmpty(&payload, "agentVersion", agentVersion)
	appendIfNotEmpty(&payload, "arch", ta.rtInfo.arch)
	appendIfNotEmpty(&payload, "ci", ta.ciInfo.provider.string())
	appendIfNotEmpty(&payload, "gitSha", ta.userGitCommit)
	appendIfNotEmpty(&payload, "platform", ta.rtInfo.platform)
	appendIfNotEmpty(&payload, "ruleName", ta.userRuleName)
	appendIfNotEmpty(&payload, "wrapperName", ta.userOverrides["wrapperName"])
	appendIfNotEmpty(&payload, "wrapperVersion", ta.userOverrides["wrapperVersion"])

	payload = "{" + payload + "}"

	return payload
}

func (ta *triggerAction) makeURL() string {
	triggerURL := ta.userOverrides["apiTriggerEndpoint"]

	if len(triggerURL) == 0 {
		triggerURL = defaultAPITriggerEndpoint
	}

	return triggerURL
}

func (ta *triggerAction) triggerRun(retryAllowed bool) (bool, error) {
	fmt.Printf("Triggering run on Waldo…\n")

	url := ta.makeURL()
	body := ta.makePayload()

	client := &http.Client{}

	req, err := http.NewRequest("POST", url, strings.NewReader(body))

	if err != nil {
		return false, fmt.Errorf("Unable to trigger run on Waldo, error: %v, url: %q", err, url)
	}

	req.Header.Add("Authorization", ta.authorization())
	req.Header.Add("Content-Type", ta.contentType())
	req.Header.Add("User-Agent", ta.userAgent())

	dumpRequest(ta.userVerbose, req, true)

	resp, err := client.Do(req)

	if err != nil {
		return retryAllowed, fmt.Errorf("Unable to trigger run on Waldo, error: %v, url: %q", err, url)
	}

	dumpResponse(ta.userVerbose, resp, true)

	defer resp.Body.Close()

	return retryAllowed && shouldRetry(resp), ta.checkTriggerStatus(resp)
}

func (ta *triggerAction) triggerRunWithRetry() error {
	for attempts := 1; attempts <= maxNetworkAttempts; attempts++ {
		retry, err := ta.triggerRun(attempts < maxNetworkAttempts)

		if !retry || err == nil {
			return err
		}

		emitError(err)

		fmt.Printf("\nFailed trigger attempts: %d -- retrying…\n\n", attempts)
	}

	return nil

}

func (ta *triggerAction) userAgent() string {
	ci := ta.ciInfo.provider.string()

	if ci == "Unknown" {
		ci = "Go Agent"
	}

	version := ta.userOverrides["wrapperVersion"]

	if len(version) == 0 {
		version = agentVersion
	}

	return fmt.Sprintf("Waldo %s v%s", ci, version)
}
