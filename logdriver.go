package main

import (
	"flag"
	"log"

	"github.com/ActiveState/tail"
	"github.com/smn/logdriver/http"
)

type Subscriber struct {
	filepath string
	messages chan<- string
}

type ClientInterface struct {
	registerSubscriber   chan Subscriber
	deregisterSubscriber chan Subscriber
}

type LogDriver struct {
	directory   string
	tails       []*tail.Tail
	subscribers []*Subscriber
}

func NewLogDriver(directory string) (l LogDriver) {
	return LogDriver{
		directory,
		make([]*tail.Tail, 0, 0),
		make([]*Subscriber, 0, 0)}
}

func (l LogDriver) Tail(filepath string) (t *tail.Tail, err error) {
	t, err = tail.TailFile(filepath, tail.Config{Follow: true})
	return t, err
}

func (l LogDriver) RegisterClientInterface(v interface{}) {
	ci := v.(ClientInterface)
	go func() {
		select {
		case s := <-ci.registerSubscriber:
			tail, _ := l.Tail(s.filepath)
			go func() {
				for {
					tailLine, _ := <-tail.Lines
					s.messages <- tailLine.Text
				}
			}()
			l.subscribers = append(l.subscribers, &s)
			log.Println("Subscribed " + s.filepath)
		case s := <-ci.deregisterSubscriber:
			for index, subscriber := range l.subscribers {
				if *subscriber == s {
					l.subscribers = append(l.subscribers[:index], l.subscribers[index+1:]...)
					log.Println("Unsubscribed " + s.filepath)
				}
			}
		}
	}()
}

func main() {

	var directory string
	var address string
	flag.StringVar(&directory, "directory", "", "The directory to tail log files from.")
	flag.StringVar(&directory, "d", "", " (shorthand for -directory)")
	flag.StringVar(&address, "address", ":3000", "The address to bind to.")
	flag.StringVar(&address, "-a", ":3000", " (shorthand for -address)")
	flag.Parse()

	if directory == "" {
		flag.Usage()
		return
	}

	ld := NewLogDriver(directory)
	ld.RegisterClientInterface(http.NewHTTPInterface(address))
}
