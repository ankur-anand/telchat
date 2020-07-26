package pkg

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type restAPIHandler struct {
	mio           *messageIO
	mux           *http.ServeMux
	chatDataStore *chatDataStore
}

func (rh *restAPIHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	rh.mux.ServeHTTP(writer, request)
}

func newRestAPIHandler(io *messageIO, store *chatDataStore) *restAPIHandler {
	mux := http.NewServeMux()
	rh := &restAPIHandler{mio: io, mux: mux, chatDataStore: store}
	mux.Handle("/messages", http.HandlerFunc(rh.messageHandler))
	mux.Handle("/post", http.HandlerFunc(rh.postMessageHandler))
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

type message struct {
	Name string `json:"name"`
	Room string `json:"room"`
	Msg  string `json:"msg"`
}

// post message handler
func (rh *restAPIHandler) postMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}
	var m message
	err := json.NewDecoder(r.Body).Decode(&m)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "bad request body", http.StatusBadRequest)
		return
	}
	if len(m.Name) == 0 || len(m.Msg) == 0 || len(m.Room) == 0 {
		http.Error(w, "bad request body", http.StatusBadRequest)
		return
	}
	// req context can get closed anytime so don;t use request context.
	rh.chatDataStore.broadcastMsg(context.TODO(), m.Name, m.Room, []byte(formatDM(m.Name, m.Room, m.Msg)))
	rh.logWriter(m.Msg)
	w.WriteHeader(201)
}

func (rh *restAPIHandler) logWriter(command string) {
	_, err := rh.mio.Write([]byte(command + "\n\r")) // write message to the log file
	if err != nil {
		log.Println("error writing message to the log file")
	}
}
