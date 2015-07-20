package main

import (
    "flag"
    "github.com/ActiveState/tail"
    "fmt"
    "log"
    "net/http"
)

type LogDriver struct {
    directory string
    tails []*tail.Tail
    clients map[chan string]bool
    newClient chan chan string
    defunctClients chan chan string
    messages chan string
}

func NewLogDriver(directory string) (l LogDriver) {
    return LogDriver{
        directory,
        make([]*tail.Tail, 0, 0),
        make(map[chan string]bool),
        make(chan (chan string)),
        make(chan (chan string)),
        make(chan string),
    }
}

func (l LogDriver) Tail(filepath string) (t *tail.Tail, err error) {
    t, err = tail.TailFile(filepath, tail.Config{Follow: true})
    l.tails = append(l.tails, t)
    return t, err
}

// NOTE:    A bunch of this stuff comes from
//          https://github.com/kljensen/golang-html5-sse-example/
func (l LogDriver) Start(directory string, address string) {
    go func () {
        for {
            select {

            case s := <- l.newClient:
                l.clients[s] = true
                log.Println("Added new client.")

            case s := <- l.defunctClients:
                delete(l.clients, s)
                log.Println("Removed new client.")

            case msg := <- l.messages:
                for s, _ := range l.clients {
                    s <- msg
                }
                log.Println("Sent " + msg + " to all clients.")
            }
        }
    }()
    l.StartWebserver(address)
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

type Subscriber interface {
    Register(*l LogDriver)
    messages chan<- string
}

func (l LogDriver) AddClient(filepath string, c ClientInterface) (tail tail.Tail) {
    tail, err := l.Tail(filepath)
    l.clients[c] = tail
    go func() {
        for {
            tailLine, ok := <- tail.Lines
            c.messages <- tailLine.Text
        }
    }()
    return tail
}

func (l LogDriver) RemoveClient(filepath string, c ClientInterface) {
    tail := l.clients[c]
    delete(l.clients, c)
    tail.Close()
    tail.Cleanup()
}

func (l LogDriver) StopOnReceive(done <-chan bool) {
    <- done
    l.Stop()
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
    ld.AddSubscriber(NewHTTPInterface(address))
}
