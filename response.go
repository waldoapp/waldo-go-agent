package main

import (
	"encoding/json"
	"io"
	"net/http"
)

//-----------------------------------------------------------------------------

type UploadResponse struct {
	AgentType     string `json:"agentType"`
	AgentVersion  string `json:"agentVersion"`
	ApplicationID string `json:"applicationId"`
	AppName       string `json:"name,omitempty"`
	AppVersion    string `json:"version,omitempty"`
	AppVersionID  string `json:"id"`
	GitHash       string `json:"gitSha,omitempty"`
	// GitMetadata   string `json:"gitMetadata"`
	MinOSVersion  string   `json:"minimumOsVersion,omitempty"`
	PackageName   string   `json:"packageName,omitempty"`
	Size          int      `json:"size"`
	SupportedABIs []string `json:"supportedAbis"`
	UploadStatus  string   `json:"status"`
	UploadType    string   `json:"type"`
	VariantName   string   `json:"variantName,omitempty"`
}

//-----------------------------------------------------------------------------

func parseUploadResponse(resp *http.Response) (*UploadResponse, error) {
	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	ur := &UploadResponse{}

	if err = json.Unmarshal(data, ur); err != nil {
		return nil, err
	}

	return ur, nil
}
