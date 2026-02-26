package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

// turn video's url to a presigned url and store in memory
func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {

	// split video's url to get bucket and key
	if video.VideoURL == nil {
		return video, nil
	}
	splitString := strings.Split(*video.VideoURL, ",")
	if len(splitString) != 2 {
		return video, nil
	}
	bucket := splitString[0]
	key := splitString[1]

	// set expire time for the url
	expiration := 5 * time.Minute

	// get a presigned URL for the video
	presignedURL, err := generatePresignedURL(cfg.s3Client, bucket, key, expiration)
	if err != nil {
		return database.Video{}, fmt.Errorf("Error getting presigned url: %w", err)
	}

	// update videoURL field of video, store in memory
	video.VideoURL = &presignedURL

	return video, nil
}