package transcoder

import (
	"fmt"
	"net/http"

	"github.com/AdityaP1502/livestreaming-platform-gcp/api/go/base"
	"github.com/gorilla/mux"
)

func InitServer(port int, ip string) base.Server {
	r := mux.NewRouter()

	initTranscoderHandler := createTranscoderHandler(true)
	terminateTranscoderHandler := createTranscoderHandler(false)

	r.HandleFunc("/init", initTranscoderHandler).Methods("POST")
	r.HandleFunc("/end", terminateTranscoderHandler).Methods("POST")

	apiServer := base.Server{
		Port: port,
		IP:   ip,
		Start: func() {
			http.ListenAndServe(fmt.Sprintf("%s:%d", ip, port), r)
		},
		App: nil,
	}

	return apiServer
}
