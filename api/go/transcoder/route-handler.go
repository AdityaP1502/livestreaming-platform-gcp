package transcoder

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	util "github.com/AdityaP1502/livestreaming-platform-gcp/api/go/util/json"
)

type StreamRequest struct {
	StreamLink  string `json:"stream-link"`
	StorageLink string `json:"storage-link"`
}

// http handler
func initTranscoderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		response, err := util.CreateJSONResponse("fail", "Invalid Content-Type header", nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	var request StreamRequest

	err := util.DecodeJSONBody(r.Body, &request)

	fmt.Println(request)

	if err != nil {
		response, err := util.CreateJSONResponse("fail", "Invalid JSON Format", nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
	}

	// check for the fields
	if request.StreamLink == "" {
		response, err := util.CreateJSONResponse("fail", "stream-link field cannot be empty", nil)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	if request.StorageLink == "" {
		response, err := util.CreateJSONResponse("fail", "storage-link field cannot be empty", nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	// start go routine to run command in background
	newDirPath := fmt.Sprintf("./stream/%s", request.StorageLink)
	err = os.MkdirAll(newDirPath, 0755)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// c := fmt.Sprintf("ffmpeg -i %s -vsync 0 -copyts -vcodec copy -f "+
	// 	"hls -hls_time 1 -hls_list_size 3 -hls_list_size 3 "+
	// 	"-hls_segment_filename '../../stream/%s/%%d.ts' '../../stream/%s/index.m3u8'",
	// 	request.StreamLink, request.StorageLink, request.StorageLink)

	go func() {
		cmd := exec.Command("/bin/sh", "scripts/start-transcoding.sh")
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()

	response, err := util.CreateJSONResponse("success",
		"transcoder server successfully created", nil)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
