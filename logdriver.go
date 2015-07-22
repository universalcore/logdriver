package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ActiveState/tail"
	"github.com/gorilla/mux"
)

// The LogDriver object, use NewLogDriver(directory) to create an instance.
type LogDriver struct {
	directory string
	Logger    *log.Logger
}

// NewLogDriver creates a new LogDriver instance, only files in subdirectories
// of the given directory are available for tailing from.
func NewLogDriver(directory string, logger *log.Logger) (l LogDriver) {
	return LogDriver{directory, logger}
}

// Tail a file found in one of the subdirectories of the LogDriver's directory
func (l LogDriver) Tail(filepath string) (t *tail.Tail, err error) {
	t, err = tail.TailFile(
		filepath,
		tail.Config{Follow: true, MustExist: true, ReOpen: true, Logger: l.Logger})
	return t, err
}

// NewRouter returns the router to be used for routing HTTP requests
func (l LogDriver) NewRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/tail/{filepath:.+}", l.ServeHTTP).Methods("GET")
	return router
}

// StartServer a webserver and bind it to the given address
func (l LogDriver) StartServer(address string) {
	// when function completes, notify via the channel
	http.Handle("/", l.NewRouter())
	l.Logger.Printf("Listening on %s.\n", address)
	http.ListenAndServe(address, nil)
}

// ServeHTTP handles the HTTP requests for tail paths
func (l LogDriver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the params from the URL path
	params := mux.Vars(r)
	filepath, err := filepath.Abs(filepath.Join(l.directory, params["filepath"]))

	// check if there's something there to begin with
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		http.Error(w, "File does not exist.", http.StatusNotFound)
		return
	}

	// We need to be able to flush when new content arrives
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	l.Logger.Printf("Starting tail of %s\n.", filepath)

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
		l.Logger.Printf("HTTP connection just closed, closed tail for %s.\n", filepath)
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
		l.Logger.Println(err)
	}

}

func main() {

	var directory string
	var address string
	var logfile string
	flag.StringVar(&directory, "directory", "", "The directory to tail log files from.")
	flag.StringVar(&directory, "d", "", " (shorthand for -directory)")
	flag.StringVar(&address, "address", "0.0.0.0:3000", "The address to bind to.")
	flag.StringVar(&address, "a", "0.0.0.0:3000", " (shorthand for -address)")
	flag.StringVar(&logfile, "logfile", "", "Which file to log to (defaults to stdout)")
	flag.StringVar(&logfile, "l", "", "Which file to log to (defaults to stdout)")
	flag.Parse()

	if directory == "" {
		flag.Usage()
		return
	}

	var logOutput *os.File
	var err error
	if logfile != "" {
		logOutput, err = os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Error opening file: %v", err)
		}
	} else {
		logOutput = os.Stdout
	}

	ld := NewLogDriver(directory, log.New(logOutput, "", log.LstdFlags))
	ld.StartServer(address)
}
