package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/hpcloud/tail"
	"github.com/gorilla/mux"
)

// The LogDriver object, use NewLogDriver(directory) to create an instance.
type LogDriver struct {
	directory string
	cors      StringSliceVar
	Logger    *log.Logger
}

// StringSliceVar allows one to accept multiple command line arguments as string
// values and collect them in a slice of strings
type StringSliceVar []string

var cors StringSliceVar

func (i *StringSliceVar) String() string {
	return fmt.Sprintf("%s", *i)
}

// Set adds a value to the String Slice
func (i *StringSliceVar) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// NewLogDriver creates a new LogDriver instance, only files in subdirectories
// of the given directory are available for tailing from.
func NewLogDriver(directory string, cors StringSliceVar, logger *log.Logger) (l LogDriver) {
	return LogDriver{directory, cors, logger}
}

// Tail a file found in one of the subdirectories of the LogDriver's directory
func (l LogDriver) Tail(filepath string, offset int64) (t *tail.Tail, err error) {
	config := tail.Config{
		Follow:    true,
		MustExist: true,
		ReOpen:    true,
		Logger:    l.Logger,
	}
	if offset > 0 {
		config.Location = &tail.SeekInfo{offset, os.SEEK_SET}
	} else if offset < 0 {
		config.Location = &tail.SeekInfo{offset, os.SEEK_END}
	} else {
		config.Location = &tail.SeekInfo{0, os.SEEK_END}
	}
	t, err = tail.TailFile(filepath, config)
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
	n := r.Header.Get("Last-Event-ID")
	if n == "" {
		n = r.FormValue("n")
		if n == "" {
			n = "0"
		}
	}
	offset, _ := strconv.ParseInt(n, 10, 64)
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

	l.Logger.Printf("Starting tail of %s from %s.\n", filepath, n)

	tail, err := l.Tail(filepath, offset)
	if err != nil {
		http.Error(w, "Failed to tail file.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	for _, originAllowed := range l.cors {
		w.Header().Add("Access-Control-Allow-Origin", originAllowed)
	}
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
			offset, _ := tail.Tell()
			fmt.Fprint(w, "event: log\n")
			fmt.Fprintf(w, "id: %d\n", offset)
			fmt.Fprintf(w, "data: %s\n", line.Text)
			fmt.Fprint(w, "retry: 0\n")
			fmt.Fprint(w, "\n")
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
	flag.StringVar(&directory, "directory", "", "The directory to tail log files from.")
	flag.StringVar(&directory, "d", "", " (shorthand for -directory)")

	var address string
	flag.StringVar(&address, "address", "0.0.0.0:3000", "The address to bind to.")
	flag.StringVar(&address, "a", "0.0.0.0:3000", " (shorthand for -address)")

	var logfile string
	flag.StringVar(&logfile, "logfile", "", "Which file to log to (defaults to stdout)")
	flag.StringVar(&logfile, "l", "", "(shorthand for -logfile)")

	var cors StringSliceVar
	flag.Var(&cors, "cors", "Whitelisted URLs for CORS (defaults to [*])")
	flag.Var(&cors, "-c", "(shorthand for -cors)")

	flag.Parse()

	if directory == "" {
		flag.Usage()
		return
	}

	if len(cors) == 0 {
		cors.Set("*")
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

	ld := NewLogDriver(directory, cors, log.New(logOutput, "", log.LstdFlags))
	ld.StartServer(address)
}
