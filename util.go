package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func appendIfNotEmpty(payload *string, key string, value string) {
	if len(key) == 0 || len(value) == 0 {
		return
	}

	if len(*payload) > 0 {
		*payload += ","
	}

	*payload += fmt.Sprintf(`"%s":"%s"`, key, value)
}

func determineBuildPayloadPath(workingPath, buildPath, buildSuffix string) string {
	buildName := filepath.Base(buildPath)

	switch buildSuffix {
	case "app":
		return filepath.Join(workingPath, buildName+".zip")

	default:
		return buildPath
	}
}

func determineWorkingPath() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("WaldoGoAgent-%d", os.Getpid()))
}

func isDir(path string) bool {
	fi, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	}

	return fi.Mode().IsDir()
}

func isRegular(path string) bool {
	fi, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	}

	return fi.Mode().IsRegular()
}

func run(name string, args ...string) (string, string, error) {
	var (
		stderrBuffer bytes.Buffer
		stdoutBuffer bytes.Buffer
	)

	cmd := exec.Command(name, args...)

	cmd.Stderr = &stderrBuffer
	cmd.Stdout = &stdoutBuffer

	err := cmd.Run()

	stderr := strings.TrimRight(stderrBuffer.String(), "\n")
	stdout := strings.TrimRight(stdoutBuffer.String(), "\n")

	return stdout, stderr, err
}

func validateBuildPath(buildPath string) (string, string, string, error) {
	if len(buildPath) == 0 {
		return "", "", "", errors.New("Empty build path")
	}

	buildPath, err := filepath.Abs(buildPath)

	if err != nil {
		return "", "", "", err
	}

	buildSuffix := strings.TrimPrefix(filepath.Ext(buildPath), ".")

	switch buildSuffix {
	case "apk":
		return buildPath, buildSuffix, "Android", nil

	case "app", "ipa":
		return buildPath, buildSuffix, "iOS", nil

	default:
		return "", "", "", fmt.Errorf("File extension of build at ‘%s’ is not recognized", buildPath)
	}
}

func validateUploadToken(uploadToken string) error {
	if len(uploadToken) == 0 {
		return errors.New("Empty upload token")
	}

	return nil
}
