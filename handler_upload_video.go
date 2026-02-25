package main

import (
	"fmt"
	"net/http"
	"mime"
	"os"
	"io"
	"crypto/rand"
	"encoding/hex"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {

	// set upload limit 1GB
	r.Body = http.MaxBytesReader(w, r.Body , 1 << 30)

	videoIDString := r.PathValue("videoID")		// extract videoID from URL
	videoID, err := uuid.Parse(videoIDString)	// convert ID string to UUID
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return 
	}

	token, err := auth.GetBearerToken(r.Header)  	// get token from request header
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}
	
	// validate user's token
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	// get video metadata from db
	videoInfo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, 500 , "Something went wrong", err)
		return
	}

	// check ownership of user with the video
	if userID != videoInfo.UserID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized user", err)
		return
	}

	// parse the uploaded video file from form data
	multipartFile, multipartFileHeader, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest , "Can't get file data", err)
		return
	}
	defer multipartFile.Close()

	// validate uploaded file. make sure it's MP4 vid
	mediaType := multipartFileHeader.Header.Get("Content-Type")
	parsedMediaType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid media type", err)
		return
	}
	if parsedMediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Invalid file type", err)
		return
	}


	// save uploaded file to temporary file on disk
	tempFile, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError( w, 500 , "Something went wrong", err)
		return
	}
	defer os.Remove(tempFile.Name()) 	// clean up
	defer tempFile.Close()				// close the file before deleting it

	// copy the uploaded video file to the temp file on disk
	_, err = io.Copy(tempFile, multipartFile)
	if err != nil {
		respondWithError(w, 500 , "Something went wrong", err)
		return
	}

	// reset file's pointer to beginning of file, allowing reading file again from the start
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError( w, 500 , "Something went wrong", err)
		return
	}

	// generate randome 32-byte name for the file
	randBytes := make([]byte, 32)
	_, err = rand.Read(randBytes)		// fill slice with random bytes
	if err != nil {
		respondWithError(w, 500 , "Something went wrong", err)
		return
	}
	hexCodedName := hex.EncodeToString(randBytes)		// encode the bytes to hex string
	fileExtension := mediaTypeToExt(parsedMediaType)	// get extension of the file
	filename := hexCodedName + fileExtension			// create file name

	// create putObjectInput
	putObjectInput := s3.PutObjectInput{
		Bucket: &cfg.s3Bucket,
		Key: &filename,
		Body: tempFile,
		ContentType: &parsedMediaType,
	}

	// put the object into s3 bucket
	_, err = cfg.s3Client.PutObject(r.Context(), &putObjectInput )
	if err != nil {
		respondWithError(w , 500 , "Can't upload file to bucket", err)
		return
	}

	// create video new URL
	videoURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, filename)
	
	// update video URL in db
	videoInfo.VideoURL = &videoURL
	err = cfg.db.UpdateVideo(videoInfo)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Can't update video url", err)
		return
	}
	respondWithJSON(w, http.StatusOK, videoInfo)
}
