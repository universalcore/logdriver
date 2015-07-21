package main

import (
	"flag"
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

func main() {

	var directory string
	flag.StringVar(&directory, "directory", "", "The directory to tail log files from.")
	flag.StringVar(&directory, "d", "", " (shorthand for -directory)")
	flag.Parse()

	if directory == "" {
		flag.Usage()
		return
	}

	ld := NewLogDriver(directory)
	ld.Start()
}
