package pkg

import (
	"log"
	"net/http"
)

type restAPIHandler struct {
	mio *messageIO
	mux *http.ServeMux
}

func (rh *restAPIHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	rh.mux.ServeHTTP(writer, request)
}

func newRestAPIHandler(io *messageIO) *restAPIHandler {
	mux := http.NewServeMux()
	rh := &restAPIHandler{mio: io, mux: mux}
	mux.Handle("/messages", http.HandlerFunc(rh.messageHandler))
	return rh
}

// all the message handler.
func (rh *restAPIHandler) messageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	msg, err := rh.mio.ReadAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	_, err = w.Write(msg)
	if err != nil {
		log.Printf("ResponseWriter error: %v", err)
	}
}
