package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/ActiveState/tail"
)

type LogDriverTest struct {
	Name string
	path string
	*testing.T
}

// NOTE:    This is me learning more of go by studying the tail
//          test cases. A bunch of this code is copied & paste across
//          from tail's test suite.
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

type TestInterface struct {
}

func (ti TestInterface) Register(l LogDriver) {
	log.Println("Registered")
}

// func TestNewLogDriver(t *testing.T) {
//     ldt := NewLogDriverTest("test_new_log_driver", t)
//     ldt.RegisterClientInterface(&TestInterface{})
//     log.Println(ldt.clients)
// }
