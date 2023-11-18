package transcoder

import (
	"fmt"
	"net/http"

	"github.com/AdityaP1502/livestreaming-platform-gcp/api/go/base"
	"github.com/gorilla/mux"
)

func InitServer(port int, ip string) base.Server {
	r := mux.NewRouter()
	r.HandleFunc("/init", initTranscoderHandler).Methods("POST")
	apiServer := base.Server{
		Port: port,
		IP:   ip,
		Start: func() {
			http.ListenAndServe(fmt.Sprintf("%s:%d", ip, port), r)
		},
	}

	return apiServer
}
