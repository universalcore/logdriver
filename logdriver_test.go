package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ActiveState/tail"
)

type LogDriverTest struct {
	Name string
	path string
	*testing.T
}

// NOTE:	This is me learning more of go by studying the tail
//			test cases. A bunch of this code is copied & paste across
//			from tail's test suite.
func NewLogDriverTest(name string, t *testing.T) LogDriverTest {
	lt := LogDriverTest{name, ".test/" + name, t}
	err := os.MkdirAll(lt.path, os.ModeTemporary|0700)
	if err != nil {
		lt.Fatal(err)
	}
	return lt
}

func (lt LogDriverTest) CreateFile(name string, contents string) (path string) {
	err := ioutil.WriteFile(lt.path+"/"+name, []byte(contents), 0600)
	if err != nil {
		lt.Fatal(err)
	}
	return lt.path + "/" + name
}

func (lt LogDriverTest) AppendFile(name string, contents string) {
	f, err := os.OpenFile(lt.path+"/"+name, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		lt.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString(contents)
	if err != nil {
		lt.Fatal(err)
	}
}

// TestMain seems to be golang's way of providing hooks for setup & teardown
// stuff
func TestMain(m *testing.M) {
	ret := m.Run()
	os.RemoveAll(".test")
	os.Exit(ret)
}

func (lt LogDriverTest) AssertTailOutput(tail *tail.Tail, lines []string, done chan bool) {
	defer func() { done <- true }()

	for idx, line := range lines {
		tailedLine, ok := <-tail.Lines
		if !ok {
			// tail.Lines is closed and empty.
			err := tail.Err()
			if err != nil {
				lt.Fatalf("tail ended with error: %v", err)
			}
			lt.Fatalf("tail ended early; expecting more: %v", lines[idx:])
		}
		if tailedLine == nil {
			lt.Fatalf("tail.Lines returned nil; not possible")
		}
		// Note: not checking .Err as the `lines` argument is designed
		// to match error strings as well.
		if tailedLine.Text != line {
			lt.Fatalf(
				"unexpected line/err from tail: "+
					"expecting <<%s>>>, but got <<<%s>>>",
				line, tailedLine.Text)
		}
	}
}

func (lt LogDriverTest) AssertRecorderOutput(w *ClosableRecorder, lines []string, done chan<- bool) {
	defer func() { done <- true }()

	receivedLines := strings.Split(w.Body.String(), "\n\n")
	for index, expectedLine := range lines {
		receivedLine := receivedLines[index]
		if expectedLine != receivedLine {
			lt.Fatalf("unexpected response: expecting \"%s\" but got \"%s\"",
				expectedLine, receivedLine)
		}
	}
}

func TestTail(t *testing.T) {

	lt := NewLogDriverTest("test_tail_file", t)
	filePath := lt.CreateFile("test.txt", "foo\n")

	ld := NewLogDriver("test_tail_file")
	tail, _ := ld.Tail(filePath)

	lt.AppendFile("test.txt", "bar\nbaz\n")
	done := make(chan bool)
	go lt.AssertTailOutput(tail, []string{"foo", "bar", "baz"}, done)
	<-done
}

type ClosableRecorder struct {
	*httptest.ResponseRecorder
	closer chan bool
}

func NewClosableRecorder() *ClosableRecorder {
	r := httptest.NewRecorder()
	closer := make(chan bool)
	return &ClosableRecorder{r, closer}
}

func (r *ClosableRecorder) CloseNotify() <-chan bool {
	return r.closer
}

func TestServeHTTP(t *testing.T) {
	lt := NewLogDriverTest("test_serve_http", t)
	lt.CreateFile("foo.txt", "foo\n")

	ld := NewLogDriver(lt.path)

	r, _ := http.NewRequest("GET", "http://localhost:3000/tail/foo.txt", nil)
	w := NewClosableRecorder()

	router := ld.NewRouter()
	go router.ServeHTTP(w, r)

	lt.AppendFile("foo.txt", "bar\nbaz\n")

	<-time.After(100 * time.Millisecond)
	go lt.AssertRecorderOutput(w, []string{"data: foo", "data: bar", "data: baz"}, w.closer)

}
