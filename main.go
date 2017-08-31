package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	var code []byte

	// TODO remove this debug code
	if len(os.Args) == 1 {
		os.Args = append(os.Args, `example/slofec_viewer.help`)
	}

	if len(os.Args) == 1 {
		// read input from std in
		var err error
		code, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Printf("error reading input from STDIN: %s\n", err.Error())
			os.Exit(1)
		}
	} else if len(os.Args) == 2 {
		// read input from file
		path := os.Args[1]
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

	// TODO remove this debug code
	//doc = document{
	//	title: "SLOFEC Viewer",
	//	parts: []docPart{
	//		docText("Hello World!\nLine 2\n"),
	//		docImage{name: "setup 1 language"},
	//	},
	//}

	html, err := genHTML(doc)
	if err != nil {
		fmt.Printf("error generating HTML: %s\n", err.Error())
		os.Exit(3)
	}
	fmt.Println(html)
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
