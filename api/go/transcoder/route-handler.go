package transcoder

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/AdityaP1502/livestreaming-platform-gcp/api/go/base"
	util "github.com/AdityaP1502/livestreaming-platform-gcp/api/go/util/json"
)

type StreamRequest struct {
	StreamLink  string `json:"stream-link"`
	StorageLink string `json:"storage-link"`
}

var API_URL string = os.Getenv("API_URL")

func checkHTTPRequest(r *http.Request) (base.Response, StreamRequest, error) {
	contentType := r.Header.Get("Content-Type")

	if contentType != "application/json" {
		return base.Response{
			Status:     "fail",
			Message:    "Invalid Content-Type. Request need to be in application/json",
			Data:       nil,
			StatusCode: http.StatusBadRequest,
		}, StreamRequest{}, nil
	}

	var request StreamRequest
	err := util.DecodeJSONBody(r.Body, &request)
	if err != nil {
		return base.Response{
			Status:     "fail",
			Message:    "Malformed JSON data",
			Data:       nil,
			StatusCode: http.StatusBadRequest,
		}, StreamRequest{}, err
	}

	if request.StreamLink == "" {
		return base.Response{
			Status:     "fail",
			Message:    "stream-link field cannot be empty",
			Data:       nil,
			StatusCode: http.StatusBadRequest,
		}, StreamRequest{}, nil
	}

	if request.StorageLink == "" {
		return base.Response{
			Status:     "fail",
			Message:    "storage-link field cannot be empty",
			Data:       nil,
			StatusCode: http.StatusBadRequest,
		}, StreamRequest{}, nil
	}

	return base.Response{Status: "success"}, request, nil
}

func createTranscoderHandler(initTranscoder bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response, request, err := checkHTTPRequest(r)

		if err != nil {
			fmt.Printf("[ERROR] error while checking request. %s\n", err.Error())
		}

		if response.Status == "fail" {
			jsonResponse, err := util.CreateJSONResponse(response)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(response.StatusCode)
			w.Write(jsonResponse)

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
			// cmd := exec.Command(
			// 	"/bin/sh",
			// 	path+"/scripts/start-transcoding.sh",
			// 	request.StreamLink,
			// 	request.StorageLink,
			// )

			cmd := exec.Command(
				"./scripts/start-transcoding.sh",
				request.StreamLink,
				request.StorageLink,
			)

			cmd.Stdout = os.Stdout

			if !initTranscoder {
				cmd = exec.Command(
					"./scripts/end-transcoding.sh",
					request.StorageLink,
				)
			}

			fmt.Println(cmd)

			err = cmd.Run()
			if err != nil {
				fmt.Println(err.Error())
				log.Fatal(err)
			}

			time.Sleep(1 * time.Second)

			// send HTTP post request to the api
			// in the format of API_URL/{username}/{stream-id}

			requestURL := fmt.Sprintf("%s/%s", API_URL, request.StorageLink)
			payload := `{"active":"true"}`

			req, err := http.NewRequest("PATCH", requestURL, strings.NewReader(payload))
			if err != nil {
				fmt.Println("Failed to send PATCH request to API server." + err.Error())
				return
			}

			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}

			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Failed to send PATCH request to API server." + err.Error())
				return
			}

			defer resp.Body.Close()

			if resp.StatusCode == 200 {
				return
			}

			fmt.Printf("Failed to send PATCH request to API server.Get %d status response", resp.StatusCode)
		}()

		response.Message = "transcoder server successfully created"

		if !initTranscoder {
			response.Message = "transcoder server successfully terminated"
		}

		jsonResponse, err := util.CreateJSONResponse(response)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}

// http handler
func initTranscoderHandler(w http.ResponseWriter, r *http.Request) {
	createTranscoderHandler(true)(w, r)
}

func terminateTranscoderHandler(w http.ResponseWriter, r *http.Request) {
	createTranscoderHandler(false)(w, r)
}
