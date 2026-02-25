package main

import (
	"fmt"
	"net/http"
	"io"
	"errors"
	"database/sql"
	"path/filepath"
	"os"
	"mime"

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
	uploadedFileData, fileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Can't get file data", err)
		return
	}
	defer uploadedFileData.Close()

	mediaType := fileHeader.Header.Get("Content-Type")
	// parse mediatype first to strip parameters
	parsedMediaType , _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Can't parse header",err)
		return
	}
	// get file extension
	fileExtension := mediaTypeToExt(parsedMediaType)

	filename := videoID.String() + fileExtension 				// name of the file will be in filesystem
	dataPath := filepath.Join(cfg.assetsRoot, filename)			// create filesystem path
	newFile, err := os.Create(dataPath)							// create new file in the filesystem
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "can't create new file", err)
		return
	}
	defer newFile.Close() 

	// get video from video ID
	videoData, err := cfg.db.GetVideo(videoID)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusNotFound, "Not found", err)
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

	// copying the uploaded file to the fiesystem on disk
	_, err = io.Copy(newFile, uploadedFileData)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Can't save file", err)
		return
	}

	thumbPath := "/assets/" + filename
	dataURL := fmt.Sprintf("http://localhost:%s%s", cfg.port, thumbPath)
	// updatenew thumbnail url in db
	videoData.ThumbnailURL = &dataURL
	err = cfg.db.UpdateVideo(videoData)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Can't update thumbnail", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoData)
}
