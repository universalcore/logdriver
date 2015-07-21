package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/ActiveState/tail"
	"github.com/gorilla/mux"
)

// The LogDriver object, use NewLogDriver(directory) to create an instance.
type LogDriver struct {
	directory string
}

// NewLogDriver creates a new LogDriver instance, only files in subdirectories
// of the given directory are available for tailing from.
func NewLogDriver(directory string) (l LogDriver) {
	return LogDriver{directory}
}

// Tail a file found in one of the subdirectories of the LogDriver's directory
func (l LogDriver) Tail(filepath string) (t *tail.Tail, err error) {
	t, err = tail.TailFile(
		filepath,
		tail.Config{Follow: true, MustExist: true, ReOpen: true})
	return t, err
}

// StartServer a webserver and bind it to the given address
func (l LogDriver) StartServer(address string) {
	// when function completes, notify via the channel
	router := mux.NewRouter()
	router.HandleFunc("/tail/{filepath:.+}", l.ServeHTTP).Methods("GET")
	http.Handle("/", router)
	log.Printf("Listening on %s.\n", address)
	http.ListenAndServe(address, nil)
}

// ServeHTTP handles the HTTP requests for tail paths
func (l LogDriver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the params from the URL path
	params := mux.Vars(r)
	filepath, err := filepath.Abs(filepath.Join(l.directory, params["filepath"]))

	// We need to be able to flush when new content arrives
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	log.Printf("Starting tail of %s\n.", filepath)

	tail, err := l.Tail(filepath)
	if err != nil {
		http.Error(w, "Failed to tail file.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	f.Flush()

	// Get notified when clients go away so we can clean up properly.
	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		tail.Stop()
		tail.Cleanup()
		log.Printf("HTTP connection just closed, closed tail for %s.\n", filepath)
	}()

	// Wait for data to appear on the tail & relay to the browser connection
	go func() {
		for line := range tail.Lines {
			fmt.Fprintf(w, "data: %s\n\n", line.Text)
			f.Flush()
		}
		tail.Stop()
		tail.Cleanup()
	}()

	err = tail.Wait()
	if err != nil {
		log.Println(err)
	}

}

func main() {

	var directory string
	var address string
	flag.StringVar(&directory, "directory", "", "The directory to tail log files from.")
	flag.StringVar(&directory, "d", "", " (shorthand for -directory)")
	flag.StringVar(&address, "address", "0.0.0.0:3000", "The address to bind to.")
	flag.StringVar(&address, "a", "0.0.0.0:3000", " (shorthand for -address)")
	flag.Parse()

	if directory == "" {
		flag.Usage()
		return
	}

	ld := NewLogDriver(directory)
	ld.StartServer(address)
}
