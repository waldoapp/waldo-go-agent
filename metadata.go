package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

//-----------------------------------------------------------------------------

type UploadMetadata struct {
	AppVersionID string    `json:"appVersionID"`
	UploadTime   time.Time `json:"uploadTime,omitempty"`
}

//-----------------------------------------------------------------------------

func newUploadMetadata(ur *UploadResponse) *UploadMetadata {
	return &UploadMetadata{
		AppVersionID: ur.AppVersionID,
		UploadTime:   time.Now()}
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
