package http

import (
    "github.com/gorilla/mux"
)

type HTTPInterface struct {
    logdriver *LogDriver
}

func NewHTTPInterface(address string) {
    return HTTPInterface{nil}
}

func (h HTTPInterface) Register(*l LogDriver) {
    h.logdriver = l
    h.Start()
}

func NewRouter() {
    router := mux.NewRouter()
    router.HandleFunc("/tail/{path:.+}", l.ServeHTTP)
    http.Handle("/", router)
    http.ListenAndServe(address, nil)
}

func (h HTTPInterface) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    filepath := params["filepath"]

    tail, err := l.Tail(filepath)
    go func () {
        for {
            tailLine, err := <- tail.Lines
            for client := range l.clients {

            }
        }
    }

    f, ok := w.(http.Flusher)
    if !ok {
		http.Error(w, "Streaming unsupported.", http.StatusInternalServerError)
		return
	}

    // Set the headers related to event streaming.
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    messageChan := make(chan string)
    // got a new client, register it.
	l.newClient <- messageChan

    // wait for closed notifications and deregister client when received.
    cn, ok := w.(http.CloseNotifier)
    if !ok {
        http.Error(w, "Cannot stream.", http.StatusInternalServerError)
        return
    }

    for {
        select {
        case <-cn.CloseNotify():
            l.defunctClients <- messageChan
            log.Println("HTTP Connection closed.")
        case message := <- messageChan:
            fmt.Fprint(w, "data: %s\n\n", message)
            f.Flush()
        }
    }
}
