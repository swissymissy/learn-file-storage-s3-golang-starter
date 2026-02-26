package main 

import (
	"os/exec"
	"bytes"
)


// take filepath and return new path to a file with "fast start" encoding
func processVideoForFastStart( filePath string) (string, error) {

	outputFilePath := filePath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputFilePath)
	var stderr bytes.Buffer		// create a buffer to capture stderr
	cmd.Stderr = &stderr		// write stderr to buffer
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	
	return outputFilePath, nil
}