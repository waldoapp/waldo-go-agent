package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

type rtInfo struct {
	arch     string
	platform string
}

//-----------------------------------------------------------------------------

func detectRTInfo() *rtInfo {
	return &rtInfo{
		arch:     detectArch(),
		platform: detectPlatform()}
}

//-----------------------------------------------------------------------------

func (ri *rtInfo) version() string {
	tmpVersion := fmt.Sprintf("%s %s (%s/%s)\n", agentName, agentVersion, ri.platform, ri.arch)

	wrapperName := os.Getenv("WALDO_WRAPPER_NAME_OVERRIDE")
	wrapperVersion := os.Getenv("WALDO_WRAPPER_VERSION_OVERRIDE")

	if len(wrapperName) > 0 && len(wrapperVersion) > 0 {
		return fmt.Sprintf("%s %s / %s", wrapperName, wrapperVersion, tmpVersion)
	}

	return tmpVersion
}

//-----------------------------------------------------------------------------

func detectArch() string {
	arch := runtime.GOARCH

	switch arch {
	case "amd64":
		return "x86_64"

	default:
		return arch
	}
}

func detectPlatform() string {
	platform := runtime.GOOS

	switch platform {
	case "darwin":
		return "macOS"

	default:
		return strings.Title(platform)
	}
}
