package public

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/AdityaP1502/livestreaming-platform-gcp/api/go/base"
	"github.com/AdityaP1502/livestreaming-platform-gcp/api/go/util/db"
	"github.com/gorilla/mux"
)

func InitServer(port int, ip string) base.Server {
	r := mux.NewRouter()

	sqlPort, err := strconv.Atoi(os.Getenv("MYSQL_CONNECTION_PORT"))

	if err != nil {
		log.Fatal("Invalid sql port")
	}

	app := db.OpenDatabase(
		os.Getenv("MYSQL_USERNAME"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_CONNECTION_ADDRESS"),
		sqlPort,
		os.Getenv("MYSQL_DATABASE_NAME"),
	)

	handler := NewServerHandler(&app)

	r.HandleFunc("/register", handler.registerHandler).Methods("POST")
	r.HandleFunc("/login", handler.loginHandler).Methods("POST")
	r.HandleFunc("/stream", handler.getStreamHandler).Methods("GET")
	r.HandleFunc("/stream", handler.createStreamHandler).Methods("POST")
	r.HandleFunc("/stream/{username}/{stream-id}", handler.updateStreamStatusHadler).Methods("PATCH")
	r.HandleFunc("/stream/{username}/{stream-id}/metadata", handler.postStreamMetadataHandler).Methods("POST")
	r.HandleFunc("/stream/{username}/{stream-id}/metadata", handler.updateStreamMetadataHandler).Methods("PATCH")
	r.HandleFunc("/stream/{username}/{stream-id}", handler.endStreamHandler).Methods("DELETE")

	apiServer := base.Server{
		Port: port,
		IP:   ip,
		Start: func() {
			http.ListenAndServe(fmt.Sprintf("%s:%d", ip, port), r)
		},

		App: &app,
	}

	return apiServer
}
