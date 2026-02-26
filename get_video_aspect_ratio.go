package main

import (
	
	"os/exec"
	"bytes"
	"encoding/json"
)

type FFProbeOutput struct {
	Streams []struct {
		Width	int		`json:"width"`
		Height	int 	`json:"height"`
	} `json:"streams"`
}

// function takes filepath and returns aspect ratio as string
func getVideoAspectRatio(filepath string) (string, error) {

	// create a bytes buffer
	var b bytes.Buffer

	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filepath )
	cmd.Stdout = &b
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	var data FFProbeOutput
	err = json.Unmarshal(b.Bytes(), &data)
	if err != nil {
		return "", err
	}

	// determine the ratio
	var ratioStr string
	ratio := data.Streams[0].Width / data.Streams[0].Height
	if ratio == 0 {
		ratioStr = "portrait"
	} else if ratio == 1 {
		ratioStr = "landscape"
	} else {
		ratioStr = "other"
	}

	return ratioStr, nil
}