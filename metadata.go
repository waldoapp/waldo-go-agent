package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

//-----------------------------------------------------------------------------

type UploadMetadata struct {
	AppID        string    `json:"appID"`
	AppVersionID string    `json:"appVersionID"`
	Host         string    `json:"host"`
	UploadTime   time.Time `json:"uploadTime"`
}

//-----------------------------------------------------------------------------

func (um *UploadMetadata) string() string {
	if um == nil {
		return ""
	}

	data, err := json.Marshal(um)

	if err != nil {
		return ""
	}

	return string(data)
}

//-----------------------------------------------------------------------------

func (um *UploadMetadata) save() error {
	path, err := os.UserHomeDir()

	if err != nil {
		return err
	}

	dirPath := filepath.Join(path, ".waldo", "builds")

	if err := os.MkdirAll(dirPath, 0700); err != nil {
		return err
	}

	fileName := um.UploadTime.Format("2006-01-02-15-04-05")
	dataPath := filepath.Join(dirPath, fileName+".json")

	data, err := json.Marshal(um)

	if err != nil {
		return err
	}

	if err := os.WriteFile(dataPath, data, 0600); err != nil {
		return err
	}

	return nil
}
