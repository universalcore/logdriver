package main

import (
	"flag"
)

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
