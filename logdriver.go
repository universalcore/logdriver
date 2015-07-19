package main

import (
	"flag"
	"fmt"

	"github.com/ActiveState/tail"
)

func main() {

	var filename string
	config := tail.Config{Follow: true}

	flag.StringVar(&filename, "filename", "", "The file to tail.")
	flag.StringVar(&filename, "F", "", " (shorthand for -filename)")
	flag.Parse()

	done := make(chan bool)
	go tailFile(filename, config, done)
	<-done
}

func tailFile(filename string, config tail.Config, done chan bool) {
	defer func() { done <- true }()
	t, err := tail.TailFile(filename, config)
	if err != nil {
		fmt.Println(err)
		return
	}
	for line := range t.Lines {
		fmt.Println(line.Text)
	}
	err = t.Wait()
	if err != nil {
		fmt.Println(err)
	}
}
