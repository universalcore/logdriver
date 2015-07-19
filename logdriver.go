package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ActiveState/tail"
)

func main() {

	var filename string
	flag.StringVar(&filename, "filename", "", "The file to tail.")
	flag.StringVar(&filename, "F", "", " (shorthand for -filename)")
	flag.Parse()

	if filename == "" {
		flag.Usage()
		return
	}

	done := make(chan bool)
	go tailFile(filename, tail.Config{Follow: true}, done)
	<-done
}

func tailFile(filename string, config tail.Config, done chan bool) {
	// when function completes, notify via the channel
	defer func() { done <- true }()

	// start tailing
	t, err := tail.TailFile(filename, config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for line := range t.Lines {
		fmt.Println(line.Text)
	}
	err = t.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
