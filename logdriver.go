package main

import (
	"fmt"

	"github.com/ActiveState/tail"
)

func main() {

	config := tail.Config{Follow: true}
	done := make(chan bool)
	go tailFile("foo.txt", config, done)
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
