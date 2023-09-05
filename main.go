package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	agentName    = "Waldo Agent"
	agentVersion = "2.1.2"

	defaultAPIBuildEndpoint   = "https://api.waldo.com/versions"
	defaultAPIErrorEndpoint   = "https://api.waldo.com/uploadError"
	defaultAPITriggerEndpoint = "https://api.waldo.com/suites"

	maxNetworkAttempts = 2
)

var (
	agentBuildPath   string
	agentCommand     string
	agentGitBranch   string
	agentGitCommit   string
	agentRuleName    string
	agentUploadToken string
	agentVariantName string
	agentVerbose     bool
)

func checkBuildPath() {
	if len(agentBuildPath) == 0 {
		failMissingArg("build-path")
	}
}

func checkUploadToken() {
	if len(agentUploadToken) == 0 {
		agentUploadToken = os.Getenv("WALDO_UPLOAD_TOKEN")
	}

	if len(agentUploadToken) == 0 {
		failMissingOpt("--upload_token")
	}
}

func displaySummary(context any) {
	switch {
	case isTriggerCommand():
		ta := context.(*triggerAction)

		fmt.Printf("\n")
		fmt.Printf("Git commit:          %s\n", summarize(ta.gitCommit()))
		fmt.Printf("Rule name:           %s\n", summarize(ta.ruleName()))
		fmt.Printf("Upload token:        %s\n", summarizeSecure(ta.uploadToken()))
		fmt.Printf("\n")

	case isUploadCommand():
		ua := context.(*uploadAction)

		fmt.Printf("\n")
		fmt.Printf("Build path:          %s\n", summarize(ua.buildPath()))
		fmt.Printf("Git branch:          %s\n", summarize(ua.gitBranch()))
		fmt.Printf("Git commit:          %s\n", summarize(ua.gitCommit()))
		fmt.Printf("Upload token:        %s\n", summarizeSecure(ua.uploadToken()))
		fmt.Printf("Variant name:        %s\n", summarize(ua.variantName()))

		if agentVerbose {
			fmt.Printf("\n")
			fmt.Printf("Build payload path:  %s\n", summarize(ua.buildPayloadPath()))
			fmt.Printf("CI git branch:       %s\n", summarize(ua.ciGitBranch()))
			fmt.Printf("CI git commit:       %s\n", summarize(ua.ciGitCommit()))
			fmt.Printf("CI provider:         %s\n", summarize(ua.ciProvider()))
			fmt.Printf("Git access:          %s\n", summarize(ua.gitAccess()))
			fmt.Printf("Inferred git branch: %s\n", summarize(ua.inferredGitBranch()))
			fmt.Printf("Inferred git commit: %s\n", summarize(ua.inferredGitCommit()))
		}

		fmt.Printf("\n")
	}
}

func displayUsage() {
	switch {
	case isTriggerCommand():
		fmt.Printf(`OVERVIEW: Trigger run on Waldo

USAGE: waldo trigger [options]

OPTIONS:

  --git_commit <value>    Hash of originating git commit
  --help                  Display available options and exit
  --rule_name <value>     Rule name
  --upload_token <value>  Upload token (overrides WALDO_UPLOAD_TOKEN)
  --verbose               Display extra verbiage
  --version               Display version and exit
`)

	case isUploadCommand():
		fallthrough

	default:
		fmt.Printf(`OVERVIEW: Upload build to Waldo

USAGE: waldo upload [options] <build-path>

OPTIONS:

  --git_branch <value>    Branch name for originating git commit
  --git_commit <value>    Hash of originating git commit
  --help                  Display available options and exit
  --upload_token <value>  Upload token (overrides WALDO_UPLOAD_TOKEN)
  --variant_name <value>  Variant name
  --verbose               Display extra verbiage
  --version               Display version and exit
`)
	}
}

func displayVersion() {
	fmt.Printf("%s\n", detectRTInfo().version())
}

func emitError(err error) {
	fmt.Printf("\n") // flush stdout

	os.Stderr.WriteString(fmt.Sprintf("waldo: %v\n", err))
}

func fail(err error) {
	emitError(err)

	os.Exit(1)
}

func failMissingArg(arg string) {
	failUsage(fmt.Errorf("Missing required argument: ‘%s’", arg))
}

func failMissingOpt(opt string) {
	failUsage(fmt.Errorf("Missing required option: ‘%s’", opt))
}

func failMissingOptValue(opt string) {
	failUsage(fmt.Errorf("Missing required value for option: ‘%s’", opt))
}

func failUnknownArg(arg string) {
	failUsage(fmt.Errorf("Unknown argument: ‘%s’", arg))
}

func failUnknownOpt(opt string) {
	failUsage(fmt.Errorf("Unknown option: ‘%s’", opt))
}

func failUsage(err error) {
	emitError(err)

	displayUsage()

	os.Exit(1)
}

func getOverrides() map[string]string {
	overrides := map[string]string{}

	if apiBuildEndpoint := os.Getenv("WALDO_API_BUILD_ENDPOINT_OVERRIDE"); len(apiBuildEndpoint) > 0 {
		overrides["apiBuildEndpoint"] = apiBuildEndpoint
	}

	if apiErrorEndpoint := os.Getenv("WALDO_API_ERROR_ENDPOINT_OVERRIDE"); len(apiErrorEndpoint) > 0 {
		overrides["apiErrorEndpoint"] = apiErrorEndpoint
	}

	if apiTriggerEndpoint := os.Getenv("WALDO_API_TRIGGER_ENDPOINT_OVERRIDE"); len(apiTriggerEndpoint) > 0 {
		overrides["apiTriggerEndpoint"] = apiTriggerEndpoint
	}

	if wrapperName := os.Getenv("WALDO_WRAPPER_NAME_OVERRIDE"); len(wrapperName) > 0 {
		overrides["wrapperName"] = wrapperName
	}

	if wrapperVersion := os.Getenv("WALDO_WRAPPER_VERSION_OVERRIDE"); len(wrapperVersion) > 0 {
		overrides["wrapperVersion"] = wrapperVersion
	}

	return overrides
}

func isTriggerCommand() bool {
	return agentCommand == "trigger"
}

func isUploadCommand() bool {
	return agentCommand == "upload"
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fail(fmt.Errorf("Unhandled panic: %v", err))
		}
	}()

	displayVersion()

	parseArgs()

	switch {
	case isTriggerCommand():
		performTriggerAction()

	case isUploadCommand():
		performUploadAction()
	}
}

func parseArgs() {
	args := os.Args[1:]

	if len(args) == 0 {
		displayUsage()

		os.Exit(0)
	}

	agentCommand, args = parseCommand(args)

	for len(args) > 0 {
		arg := args[0]
		args = args[1:]

		switch arg {
		case "--help":
			displayUsage()

			os.Exit(0)

		case "--git_branch":
			if isUploadCommand() {
				agentGitBranch, args = parseOption(arg, args)
			} else {
				failUnknownOpt(arg)
			}

		case "--git_commit":
			agentGitCommit, args = parseOption(arg, args)

		case "--rule_name":
			if isTriggerCommand() {
				agentRuleName, args = parseOption(arg, args)
			} else {
				failUnknownOpt(arg)
			}

		case "--upload_token":
			agentUploadToken, args = parseOption(arg, args)

		case "--variant_name":
			if isUploadCommand() {
				agentVariantName, args = parseOption(arg, args)
			} else {
				failUnknownOpt(arg)
			}

		case "--verbose":
			agentVerbose = true

		case "--version":
			os.Exit(0) // version already displayed

		default:
			if strings.HasPrefix(arg, "-") {
				failUnknownOpt(arg)
			}

			if isUploadCommand() && len(agentBuildPath) == 0 {
				agentBuildPath = arg
			} else {
				failUnknownArg(arg)
			}
		}
	}
}

func parseCommand(args []string) (string, []string) {
	switch args[0] {
	case "trigger", "upload":
		return args[0], args[1:]

	default:
		return "upload", args
	}
}

func parseOption(arg string, args []string) (string, []string) {
	if len(args) == 0 || len(args[0]) == 0 || strings.HasPrefix(args[0], "-") {
		failMissingOptValue(arg)
	}

	return args[0], args[1:]
}

func performTriggerAction() {
	checkUploadToken()

	ta := newTriggerAction(
		agentUploadToken,
		agentRuleName,
		agentGitCommit,
		agentVerbose,
		getOverrides())

	if err := ta.validate(); err != nil {
		fail(err)
	}

	displaySummary(ta)

	if err := ta.perform(); err != nil {
		fail(err)
	}

	fmt.Printf("\nRun successfully triggered on Waldo!\n")
}

func performUploadAction() {
	checkBuildPath()
	checkUploadToken()

	ua := newUploadAction(
		agentBuildPath,
		agentUploadToken,
		agentVariantName,
		agentGitCommit,
		agentGitBranch,
		agentVerbose,
		getOverrides())

	if err := ua.validate(); err != nil {
		fail(err)
	}

	displaySummary(ua)

	if err := ua.perform(); err != nil {
		fail(err)
	}

	fmt.Printf("\nBuild ‘%s’ successfully uploaded to Waldo!\n", filepath.Base(agentBuildPath))
}

func summarize(value string) string {
	if len(value) > 0 {
		return fmt.Sprintf("‘%s’", value)
	} else {
		return "(none)"
	}
}

func summarizeSecure(value string) string {
	if len(value) == 0 {
		return "(none)"
	}

	if !agentVerbose {
		prefixLen := len(value)

		if prefixLen > 6 {
			prefixLen = 6
		}

		prefix := value[0:prefixLen]
		suffixLen := len(value) - len(prefix)
		secure := "********************************"

		value = prefix + secure[0:suffixLen]
	}

	return fmt.Sprintf("‘%s’", value)
}
