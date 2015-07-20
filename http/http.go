package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Subscriber struct {
	filepath string
	messages <-chan string
}

type HTTPInterface struct {
	address              string
	registerSubscriber   chan Subscriber
	deregisterSubscriber chan Subscriber
}

func NewHTTPInterface(address string) *HTTPInterface {
	hi := &HTTPInterface{address, make(chan Subscriber), make(chan Subscriber)}
	router := mux.NewRouter()
	router.HandleFunc("/tail/{path:.+}", hi.ServeHTTP)
	http.Handle("/", router)
	http.ListenAndServe(address, nil)
	return hi
}

func (h HTTPInterface) AddSubscriber(filepath string) (s Subscriber) {
	s = Subscriber{filepath, make(chan string)}
	h.registerSubscriber <- s
	return s
}

func (h HTTPInterface) RemoveSubscriber(s Subscriber) {
	h.deregisterSubscriber <- s
}

func (h HTTPInterface) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported.", http.StatusInternalServerError)
		return
	}

	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	params := mux.Vars(r)
	subscriber := h.AddSubscriber(params["filepath"])

	// wait for closed notifications and deregister client when received.
	cn, ok := w.(http.CloseNotifier)
	if !ok {
		http.Error(w, "Cannot stream.", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case <-cn.CloseNotify():
			h.RemoveSubscriber(subscriber)
			log.Println("HTTP Connection closed.")
		case message := <-subscriber.messages:
			fmt.Fprint(w, "data: %s\n\n", message)
			f.Flush()
		}
	}
}
