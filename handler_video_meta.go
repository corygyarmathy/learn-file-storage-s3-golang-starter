package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerVideoMetaCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		database.CreateVideoParams
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

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	params.UserID = userID

	video, err := cfg.db.CreateVideo(params.CreateVideoParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create video", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, video)
}

func (cfg *apiConfig) handlerVideoMetaDelete(w http.ResponseWriter, r *http.Request) {
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

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get video", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You can't delete this video", err)
		return
	}

	err = cfg.db.DeleteVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete video", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerVideoGet(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid video ID", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}

func (cfg *apiConfig) handlerVideosRetrieve(w http.ResponseWriter, r *http.Request) {
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

	videos, err := cfg.db.GetVideos(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve videos", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videos)
}

func getVideoOrientation(filePath string) (string, error) {
	aspectRatio, err := getVideoAspectRatio(filePath)
	ratios := strings.Split(aspectRatio, ":")
	w := ratios[0]
	h := ratios[1]
	if err != nil {
		return "", fmt.Errorf("getting video orientation: %v", err)
	}

	log.Printf("video aspectRatio is %v:%v", w, h)

	if w == "16" && h == "9" {
		return "landscape", nil
	} else if w == "9" && h == "16" {
		return "portrait", nil
	} else {
		return "other", nil
	}
}

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-print_format", "json",
		"-show_streams",
		filePath)
	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("running ffprobe command: %v", err)
	}

	var metadata struct {
		Streams []struct {
			CodecType          string `json:"codec_type"`
			DisplayAspectRatio string `json:"display_aspect_ratio,omitempty"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(buffer.Bytes(), &metadata); err != nil {
		return "", fmt.Errorf("unmarshalling ffprobe json: %v", err)
	}

	if len(metadata.Streams) == 0 {
		return "", errors.New("no video streams found")
	}

	var aspectRatio string
	for i := range metadata.Streams {
		if metadata.Streams[i].CodecType == "video" {
			aspectRatio = metadata.Streams[i].DisplayAspectRatio
		}
	}
	if aspectRatio == "" {
		return "", errors.New("video display aspect ratio not found in ffprobe output")
	}

	return aspectRatio, nil
}
