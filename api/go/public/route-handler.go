package public

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/AdityaP1502/livestreaming-platform-gcp/api/go/base"
	"github.com/AdityaP1502/livestreaming-platform-gcp/api/go/util/db"
	pwdhash "github.com/AdityaP1502/livestreaming-platform-gcp/api/go/util/hash"
	jsonutil "github.com/AdityaP1502/livestreaming-platform-gcp/api/go/util/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2/google"
)

type Users struct {
	FullName string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Stream struct {
	ID      string `json:"stream-id"`
	RTSPURL string `json:"rtsp-url,omitempty"`
	URL     string `json:"stream-url,omitempty"`
	UserID  int    `json:"-"`
	Status  string `json:"status,omitempty"`
}

type StreamMetadata struct {
	Title     string `json:"name"`
	CreatedAt string `json:"created-at"`
	Thumbnail string `json:"thumbnail"`
	StreamID  int    `json:"-"`
}

// ExampleHandler struct holds the dependencies for the example handler.
type ServerHandler struct {
	App *base.App
}

// NewExampleHandler initializes and returns a new instance of ExampleHandler.
func NewServerHandler(app *base.App) *ServerHandler {
	return &ServerHandler{
		App: app,
	}
}

func createBucketSignedURL(filepath string) (string, error) {
	sakeyFile := os.Getenv("GCP_SERVICE_ACCOUNT_JSON_KEY_PATH")

	saKey, err := os.ReadFile(sakeyFile)

	if err != nil {
		return "", err
	}

	cfg, err := google.JWTConfigFromJSON(saKey)

	if err != nil {
		return "", err
	}

	url, err := storage.SignedURL(
		os.Getenv("BUCKET_NAME_2"),
		filepath,
		&storage.SignedURLOptions{
			GoogleAccessID: cfg.Email,
			PrivateKey:     cfg.PrivateKey,
			Method:         "PUT",
			Expires:        time.Now().Add(time.Minute * 5),
			ContentType:    "image/png",
		},
	)

	return url, err
}

func checkUserRequest(r *http.Request, mode int) (base.Response, interface{}, error) {
	contentType := r.Header.Get("Content-Type")

	if contentType != "application/json" {
		return base.Response{
			Status:     "fail",
			Message:    "Invalid Content-Type. Request need to be in application/json",
			Data:       nil,
			StatusCode: http.StatusBadRequest,
		}, Users{}, nil
	}

	var request Users
	err := jsonutil.DecodeJSONBody(r.Body, &request)
	if err != nil {
		return base.Response{
			Status:     "fail",
			Message:    "Malformed JSON data",
			Data:       nil,
			StatusCode: http.StatusBadRequest,
		}, Users{}, err
	}

	if request.Username == "" {
		return base.Response{
			Status:     "fail",
			Message:    "Username cannot be empty",
			Data:       nil,
			StatusCode: http.StatusBadRequest,
		}, Users{}, err
	}

	if mode == 0 && request.FullName == "" {
		return base.Response{
			Status:     "fail",
			Message:    "Fullname cannot be empty",
			Data:       nil,
			StatusCode: http.StatusBadRequest,
		}, Users{}, err
	}

	if mode != 2 && request.Password == "" {
		return base.Response{
			Status:     "fail",
			Message:    "Password cannot be empty",
			Data:       nil,
			StatusCode: http.StatusBadRequest,
		}, Users{}, err
	}

	return base.Response{Status: "success"}, request, nil
}

func checkUserMetadataRequest(r *http.Request) (base.Response, interface{}, error) {
	contentType := r.Header.Get("Content-Type")

	if contentType != "application/json" {
		return base.Response{
			Status:     "fail",
			Message:    "Invalid Content-Type. Request need to be in application/json",
			Data:       nil,
			StatusCode: http.StatusBadRequest,
		}, StreamMetadata{}, nil
	}

	var request StreamMetadata
	err := jsonutil.DecodeJSONBody(r.Body, &request)
	if err != nil {
		return base.Response{
			Status:     "fail",
			Message:    "Malformed JSON data",
			Data:       nil,
			StatusCode: http.StatusBadRequest,
		}, StreamMetadata{}, err
	}

	if request.Title == "" {
		return base.Response{
			Status:     "fail",
			Message:    "Title cannot be empty",
			Data:       nil,
			StatusCode: http.StatusBadRequest,
		}, StreamMetadata{}, err
	}

	if request.CreatedAt == "" {
		return base.Response{
			Status:     "fail",
			Message:    "CreatedAt cannot be empty",
			Data:       nil,
			StatusCode: http.StatusBadRequest,
		}, StreamMetadata{}, err
	}

	return base.Response{Status: "success"}, request, nil
}

func checkStreamStatusRequest(r *http.Request) (base.Response, interface{}, error) {
	contentType := r.Header.Get("Content-Type")

	if contentType != "application/json" {
		return base.Response{
			Status:     "fail",
			Message:    "Invalid Content-Type. Request need to be in application/json",
			Data:       nil,
			StatusCode: http.StatusBadRequest,
		}, Users{}, nil
	}

	var request Stream
	err := jsonutil.DecodeJSONBody(r.Body, &request)
	if err != nil {
		return base.Response{
			Status:     "fail",
			Message:    "Malformed JSON data",
			Data:       nil,
			StatusCode: http.StatusBadRequest,
		}, Users{}, err
	}

	if request.Status == "" {
		return base.Response{
			Status:     "fail",
			Message:    "status cannot be empty",
			Data:       nil,
			StatusCode: http.StatusBadRequest,
		}, Users{}, err
	}

	return base.Response{Status: "success"}, request, nil
}

func (sh *ServerHandler) registerHandler(w http.ResponseWriter, r *http.Request) {
	response, request, err := checkUserRequest(r, 0)

	if err != nil {
		fmt.Printf("[ERROR] error while checking request. %s\n", err.Error())
	}

	if response.Status == "fail" {
		jsonResponse, err := jsonutil.CreateJSONResponse(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(response.StatusCode)
		w.Write(jsonResponse)

		return
	}

	v := request.(Users)

	if exist, err := usernameExists(v.Username, sh.App); err != nil {
		fmt.Printf("[ERROR] sql transcation failed. %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if exist {
		response.Status = "fail"
		response.Message = "Username already exist"
		jsonResponse, err := jsonutil.CreateJSONResponse(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResponse)

		return
	}

	// Hash the user password before storing
	passwordHash, err := pwdhash.HashPassword(v.Password)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Store data into database
	v.Password = passwordHash

	err = insertNewUserToDatabase(&v, sh.App)

	if err != nil {
		fmt.Printf("[ERROR] sql transcation failed. %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.Message = "Your account is successfully created"

	data, err := jsonutil.CreateJSONResponse(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)

}

func (sh *ServerHandler) loginHandler(w http.ResponseWriter, r *http.Request) {
	response, request, err := checkUserRequest(r, 1)

	if err != nil {
		fmt.Printf("[ERROR] error while checking request. %s\n", err.Error())
	}

	if response.Status == "fail" {
		jsonResponse, err := jsonutil.CreateJSONResponse(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(response.StatusCode)
		w.Write(jsonResponse)

		return
	}

	v := request.(Users)

	// Get user info from db
	result := retrieveUserInformation(&v, sh.App)

	var pwdHash string

	err = result.Scan(&pwdHash)

	exist, err := db.CheckIfRowExist(err)

	if err != nil {
		fmt.Printf("[ERROR] sql transcation failed. %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !exist || !pwdhash.CheckPassword(v.Password, pwdHash) {
		response.Status = "fail"
		response.StatusCode = http.StatusUnauthorized
		response.Message = "Username or password is wrong"

		jsonResponse, err := jsonutil.CreateJSONResponse(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(response.StatusCode)
		w.Write(jsonResponse)
		return
	}

	response.StatusCode = http.StatusOK
	response.Message = "Login successful"

	data, err := jsonutil.CreateJSONResponse(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

func (sh *ServerHandler) createStreamHandler(w http.ResponseWriter, r *http.Request) {
	response, request, err := checkUserRequest(r, 2)

	if err != nil {
		fmt.Printf("[ERROR] error while checking request. %s\n", err.Error())
	}

	if response.Status == "fail" {
		jsonResponse, err := jsonutil.CreateJSONResponse(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(response.StatusCode)
		w.Write(jsonResponse)

		return
	}

	v := request.(Users)

	result := retrieveUserID(&v, sh.App)

	var id int
	err = result.Scan(&id)

	exist, err := db.CheckIfRowExist(err)

	if err != nil {
		fmt.Printf("[ERROR] sql transcation failed. %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !exist {
		response.Status = "fail"
		response.StatusCode = http.StatusUnauthorized
		response.Message = "Username not exist"

		jsonResponse, err := jsonutil.CreateJSONResponse(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(response.StatusCode)
		w.Write(jsonResponse)
		return
	}

	streamID := uuid.New().String()
	fmt.Println("Create a new stream with ID " + streamID)

	rtspURL := fmt.Sprintf("rtsp://%s:%s/%s/%s/%s",
		os.Getenv("RTSP_LB_IP"), os.Getenv("RTSP_PORT"), "stream", v.Username, streamID)

	storageURL := fmt.Sprintf("%s/%s/%s/%s/%s",
		os.Getenv("GCP_BUCKET_URL"), os.Getenv("BUCKET_NAME"), "stream", v.Username, streamID)

	// save into database

	stream := Stream{
		ID:      streamID,
		URL:     storageURL,
		RTSPURL: rtspURL,
		UserID:  id,
		Status:  "inactive",
	}

	err = saveStreamData(&stream, sh.App)

	if err != nil {
		fmt.Printf("[ERROR] sql transcation failed. %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	url, err := createBucketSignedURL(fmt.Sprintf("%s/%s/%s/%s", "stream", v.Username, streamID, "thumbnail.png"))

	if err != nil {
		fmt.Printf("[ERROR] creating signed url failed. %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	stream.URL = ""
	stream.Status = ""

	response.Message = "Your stream is successfully provisioned"

	response.Data = base.DataField{
		"stream":    stream,
		"BucketUrl": url,
	}

	data, err := jsonutil.CreateJSONResponse(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

func (sh *ServerHandler) postStreamMetadataHandler(w http.ResponseWriter, r *http.Request) {
	response, request, err := checkUserMetadataRequest(r)

	if err != nil {
		fmt.Printf("[ERROR] error while checking request. %s\n", err.Error())
	}

	if response.Status == "fail" {
		jsonResponse, err := jsonutil.CreateJSONResponse(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(response.StatusCode)
		w.Write(jsonResponse)

		return
	}

	vars := mux.Vars(r)
	streamID := vars["stream-id"]
	username := vars["username"]

	result := streamExist(username, streamID, sh.App)
	var id int

	err = result.Scan(&id)

	exist, err := db.CheckIfRowExist(err)

	if err != nil {
		fmt.Printf("[ERROR] creating signed url failed. %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !exist {
		response.Status = "fail"
		response.StatusCode = http.StatusNotFound
		response.Message = "stream not found"

		jsonResponse, err := jsonutil.CreateJSONResponse(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(response.StatusCode)
		w.Write(jsonResponse)

		return
	}

	v := request.(StreamMetadata)

	v.Thumbnail = fmt.Sprintf("%s/%s/%s/%s/%s/%s",
		os.Getenv("GCP_BUCKET_URL"),
		os.Getenv("BUCKET_NAME_2"), "stream", username, streamID, "thumbnail.png",
	)

	v.StreamID = id

	err = insertStreamMetadata(&v, sh.App)

	if err != nil {
		fmt.Printf("[ERROR] sql transcation failed. %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.Message = "Your stream metadata is successfully created"

	data, err := jsonutil.CreateJSONResponse(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)

}

func (sh *ServerHandler) endStreamHandler(w http.ResponseWriter, r *http.Request) {
	var response base.Response

	vars := mux.Vars(r)
	streamID := vars["stream-id"]
	username := vars["username"]

	result := streamExist(username, streamID, sh.App)
	var i int

	err := result.Scan(&i)

	exist, err := db.CheckIfRowExist(err)

	if err != nil {
		fmt.Printf("[ERROR] creating signed url failed. %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !exist {
		response.Status = "fail"
		response.StatusCode = http.StatusNotFound
		response.Message = "stream not found"

		jsonResponse, err := jsonutil.CreateJSONResponse(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		w.Write(jsonResponse)

		return
	}

	result = retrieveUserID(&Users{Username: username}, sh.App)

	var id int
	err = result.Scan(&id)

	exist, err = db.CheckIfRowExist(err)

	if err != nil {
		fmt.Printf("[ERROR] creating signed url failed. %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	if !exist {
		response.Status = "fail"
		response.StatusCode = http.StatusUnauthorized
		response.Message = "Username not exist"

		jsonResponse, err := jsonutil.CreateJSONResponse(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(response.StatusCode)
		w.Write(jsonResponse)
		return
	}

	err = deleteStream(i, sh.App)
	if err != nil {
		fmt.Printf("[ERROR] sql transcation failed. %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.Message = "Your stream is successfully terminated"

	data, err := jsonutil.CreateJSONResponse(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)

}

func (sh *ServerHandler) updateStreamStatusHadler(w http.ResponseWriter, r *http.Request) {
	response, request, err := checkStreamStatusRequest(r)

	if err != nil {
		fmt.Printf("[ERROR] error while checking request. %s\n", err.Error())
	}

	if response.Status == "fail" {
		jsonResponse, err := jsonutil.CreateJSONResponse(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(response.StatusCode)
		w.Write(jsonResponse)

		return
	}

	vars := mux.Vars(r)
	streamID := vars["stream-id"]
	username := vars["username"]

	result := streamExist(username, streamID, sh.App)
	var id int

	err = result.Scan(&id)

	exist, err := db.CheckIfRowExist(err)

	if err != nil {
		fmt.Printf("[ERROR] creating signed url failed. %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !exist {
		response.Status = "fail"
		response.StatusCode = http.StatusNotFound
		response.Message = "stream not found"

		jsonResponse, err := jsonutil.CreateJSONResponse(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(response.StatusCode)
		w.Write(jsonResponse)

		return
	}

	v := request.(Stream)

	err = updateStreamStatus(v.Status, id, sh.App)

	if err != nil {
		fmt.Printf("[ERROR] sql transcation failed. %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.Message = "Your stream status is successfully updated"

	data, err := jsonutil.CreateJSONResponse(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (sh *ServerHandler) getStreamHandler(w http.ResponseWriter, r *http.Request) {
	res, err := getAllStream(sh.App)

	if err != nil {
		fmt.Printf("[ERROR] sql transcation failed. %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer res.Close()

	streamResult := []base.DataField{}

	for res.Next() {
		var (
			streamURL, username         string
			title, createdAt, thumbnail string
		)

		if err := res.Scan(&streamURL, &username, &title, &createdAt, &thumbnail); err != nil {
			fmt.Printf("[ERROR] sql transcation failed. %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		v := base.DataField{
			"username":   username,
			"stream-url": streamURL,
			"metadata": base.DataField{
				"title":     title,
				"createdAt": createdAt,
				"thumbnail": thumbnail,
			},
		}

		streamResult = append(streamResult, v)
	}

	// Check for errors from iterating over rows
	if err := res.Err(); err != nil {
		fmt.Printf("[ERROR] sql transcation failed. %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := base.Response{
		Status:  "success",
		Message: "Get all active stream successfull",
		Data: base.DataField{
			"streams": streamResult,
		},
	}

	data, err := jsonutil.CreateJSONResponse(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (sh *ServerHandler) updateStreamMetadataHandler(w http.ResponseWriter, r *http.Request) {
	response := base.Response{
		Status:  "fail",
		Message: "This path is currently under development",
	}

	w.WriteHeader(http.StatusNotImplemented)

	data, err := jsonutil.CreateJSONResponse(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application-json")
	w.Write(data)
}
