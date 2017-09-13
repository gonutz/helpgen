package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

var generators = map[string]func(document) ([]byte, error){
	"-html": genHTML,
	"-rtf":  genRTF,
}

func main() {
	var (
		code      []byte
		generator = genHTML
	)

	args := os.Args[1:]
	delArg := func(i int) {
		args = append(args[:i], args[i+1:]...)
	}
	i := 0
	for i < len(args) {
		for flag, gen := range generators {
			if args[i] == flag {
				generator = gen
				delArg(i)
				i--
				break
			}
		}
		i++
	}

	if len(args) == 0 {
		// read input from std in
		var err error
		code, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			fail(1, "error reading input from STDIN: %s\n", err.Error())
		}
	} else if len(args) == 1 {
		// read input from file
		path := args[0]
		var err error
		code, err = ioutil.ReadFile(path)
		if err != nil {
			fail(1, "unable to read file '%s': %s\n", path, err.Error())
		}
	} else {
		fail(1, "too many parameters")
	}

	doc, err := parse(code)
	if err != nil {
		fail(2, "error parsing code: %s\n", err.Error())
	}

	output, err := generator(doc)
	if err != nil {
		fail(3, "error generating output: %s\n", err.Error())
	}

	fmt.Print(string(output))
}

func fail(exitCode int, format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(exitCode)
}
