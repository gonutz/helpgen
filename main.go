package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

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
		if args[i] == "-html" {
			generator = genHTML
			delArg(i)
		} else if args[i] == "-rtf" {
			generator = genRTF
			delArg(i)
		}
		i++
	}

	if len(args) == 0 {
		// read input from std in
		var err error
		code, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Printf("error reading input from STDIN: %s\n", err.Error())
			os.Exit(1)
		}
	} else if len(args) == 1 {
		// read input from file
		path := args[0]
		var err error
		code, err = ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("unable to read file '%s': %s\n", path, err.Error())
			os.Exit(1)
		}
	} else {
		fmt.Println("too many parameters")
		os.Exit(1)
	}

	doc, err := parse(code)
	if err != nil {
		fmt.Printf("error parsing code: %s\n", err.Error())
		os.Exit(2)
	}

	output, err := generator(doc)
	if err != nil {
		fmt.Printf("error generating output: %s\n", err.Error())
		os.Exit(3)
	}
	fmt.Println(string(output))
}

type document struct {
	title string
	parts []docPart
}

type docPart interface {
	isDocPart()
}

type docText string

type docImage struct {
	name string
}

type largerDocText struct {
	text  string
	scale int
}

func (docText) isDocPart()       {}
func (docImage) isDocPart()      {}
func (largerDocText) isDocPart() {}
