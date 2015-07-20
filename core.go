package main

import (
    "github.com/ActiveState/tail"
    "fmt"
)

type LogDriver struct {
    directory string
    tails []*tail.Tail
}

func NewLogDriver(directory string) (l LogDriver) {
    return LogDriver{directory, make([]*tail.Tail, 0, 0)}
}

func (l LogDriver) Tail(filepath string) (t *tail.Tail, err error) {
    t, err = tail.TailFile(filepath, tail.Config{Follow: true})
    l.tails = append(l.tails, t)
    return t, err
}

func (l LogDriver) Start() {
    // when function completes, notify via the channel
}

func (l LogDriver) Stop() {
    for _, tail := range l.tails {
        err := tail.Stop()
        if err != nil {
            fmt.Println(err)
        }
        tail.Cleanup()
    }
}

func (l LogDriver) StopOnReceive(done <-chan bool) {
    <- done
    l.Stop()
}

/*

defer func() { done <- true }()

tail, err := LogDriver.tail()

func tailFile(filename string, messages chan<- string, done chan<- bool) {

    // start tailing
    t, err := tail.TailFile(filename, tail.Config{Follow: true})
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        return
    }
    for line := range t.Lines {
        messages <- line.Text
    }
    err = t.Wait()
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
    }
}

*/
