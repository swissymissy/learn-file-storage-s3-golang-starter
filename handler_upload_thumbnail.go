package main

import (
	"fmt"
	"net/http"
	"io"
	"errors"
	"database/sql"
	"encoding/base64"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20			// 10 shift left 20 times to get 10MB
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest , "Can't parse multipart form", err)
		return
	}
	fileData, fileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Can't get file data", err)
		return
	}
	defer fileData.Close()

	mediaType := fileHeader.Header.Get("Content-Type")
	imgData, err := io.ReadAll(fileData)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Can't read image data", err)
		return
	}

	// conver image data to a base64 string
	convertedimgData := base64.StdEncoding.EncodeToString(imgData)
	// create data URL
	dataURL := fmt.Sprintf("data:%s;base64,%s", mediaType, convertedimgData)

	// get video from video ID
	videoData, err := cfg.db.GetVideo(videoID)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusUnauthorized, "Not found", err)
		return
	} else if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong", err )
		return
	}

	// ownership check
	if videoData.UserID != userID {
		respondWithError(w , http.StatusUnauthorized, "Unauthorized", err)
		return
	} 
	
	// updatenew thumbnail url in db
	videoData.ThumbnailURL = &dataURL
	err = cfg.db.UpdateVideo(videoData)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Can't update thumbnail", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoData)
}
